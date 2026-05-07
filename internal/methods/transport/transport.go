package transport

import (
	"errors"
	"fmt"
	"math"

	ml "optimization/internal/methods/linear"
)

// Problem описывает сбалансированную транспортную задачу.
type Problem struct {
	Costs  [][]float64
	Supply []float64
	Demand []float64
}

// Result хранит транспортный план и результат решателя ЛП.
type Result struct {
	Plan   [][]float64
	Cost   float64
	Linear ml.Result
}

// Solve решает транспортную задачу через сведение к ЛП.
func Solve(p Problem, eps float64) (Result, error) {
	if eps <= 0 || math.IsNaN(eps) || math.IsInf(eps, 0) {
		return Result{}, errors.New("eps must be positive finite number")
	}

	m := len(p.Costs)
	if m == 0 {
		return Result{}, errors.New("empty transport problem")
	}
	n := len(p.Costs[0])
	if n == 0 {
		return Result{}, errors.New("empty cost matrix")
	}
	for i := 0; i < m; i++ {
		if len(p.Costs[i]) != n {
			return Result{}, fmt.Errorf("cost matrix row %d has length %d, expected %d", i, len(p.Costs[i]), n)
		}
	}
	if len(p.Supply) != m {
		return Result{}, fmt.Errorf("supply length mismatch: got %d, expected %d", len(p.Supply), m)
	}
	if len(p.Demand) != n {
		return Result{}, fmt.Errorf("demand length mismatch: got %d, expected %d", len(p.Demand), n)
	}

	supplySum := 0.0
	for _, v := range p.Supply {
		supplySum += v
	}
	demandSum := 0.0
	for _, v := range p.Demand {
		demandSum += v
	}
	if math.Abs(supplySum-demandSum) > eps {
		return Result{}, fmt.Errorf("transport problem must be balanced: supply=%.6f demand=%.6f", supplySum, demandSum)
	}

	varCount := m * n
	c := make([]float64, varCount)
	a := make([][]float64, m+n)
	b := make([]float64, m+n)
	sense := make([]ml.ConstraintSense, m+n)

	for i := range a {
		a[i] = make([]float64, varCount)
		sense[i] = ml.SenseEQ
	}

	for i := 0; i < m; i++ {
		b[i] = p.Supply[i]
		for j := 0; j < n; j++ {
			idx := i*n + j
			c[idx] = -p.Costs[i][j]
			a[i][idx] = 1
		}
	}

	for j := 0; j < n; j++ {
		b[m+j] = p.Demand[j]
		for i := 0; i < m; i++ {
			idx := i*n + j
			a[m+j][idx] = 1
		}
	}

	linearProblem := ml.Problem{
		C:     c,
		A:     a,
		B:     b,
		Sense: sense,
	}

	linearRes, err := ml.SolveSimplex(linearProblem, eps)
	if err != nil {
		return Result{}, err
	}

	plan := make([][]float64, m)
	for i := range plan {
		plan[i] = make([]float64, n)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			idx := i*n + j
			if idx < len(linearRes.X) {
				plan[i][j] = linearRes.X[idx]
			}
		}
	}

	return Result{
		Plan:   plan,
		Cost:   -linearRes.Objective,
		Linear: linearRes,
	}, nil
}

// solveByPotentials реализует метод потенциалов (MODI) с инициализацией "northwest corner".
func solveByPotentials(p Problem, eps float64) ([][]float64, []ml.IterationLP, error) {
	// будем собирать итерационную историю в формате ml.IterationLP
	trace := make([]ml.IterationLP, 0, 32)
	m := len(p.Costs)
	n := len(p.Costs[0])

	// Инициализация плана NW-corner
	plan := make([][]float64, m)
	basic := make([][]bool, m)
	for i := 0; i < m; i++ {
		plan[i] = make([]float64, n)
		basic[i] = make([]bool, n)
	}

	supply := make([]float64, m)
	demand := make([]float64, n)
	copy(supply, p.Supply)
	copy(demand, p.Demand)

	i, j := 0, 0
	basicCount := 0
	for i < m && j < n {
		x := math.Min(supply[i], demand[j])
		plan[i][j] = x
		basic[i][j] = true
		basicCount++
		supply[i] -= x
		demand[j] -= x
		if math.Abs(supply[i]) < eps && math.Abs(demand[j]) < eps {
			// оба исчерпаны — продвинем один индекс и оставим возможную вырожденность
			if j+1 < n {
				j++
			} else {
				i++
			}
		} else if math.Abs(supply[i]) < eps {
			i++
		} else if math.Abs(demand[j]) < eps {
			j++
		}
	}

	// Если базисных меньше, добавим нулевые базисные клетки
	for bi := 0; bi < m && basicCount < m+n-1; bi++ {
		for bj := 0; bj < n && basicCount < m+n-1; bj++ {
			if !basic[bi][bj] {
				basic[bi][bj] = true
				plan[bi][bj] = 0
				basicCount++
			}
		}
	}

	// Сохраним начальное состояние (k=0)
	func() {
		xvec := make([]float64, m*n)
		for ri := 0; ri < m; ri++ {
			for cj := 0; cj < n; cj++ {
				xvec[ri*n+cj] = plan[ri][cj]
			}
		}
		// objective as sum(c*x) with c = -costs (to match LP objective used elsewhere)
		obj := 0.0
		for ri := 0; ri < m; ri++ {
			for cj := 0; cj < n; cj++ {
				obj += -p.Costs[ri][cj] * xvec[ri*n+cj]
			}
		}
		trace = append(trace, ml.IterationLP{K: 0, EnterVar: 0, LeaveVar: 0, Objective: obj, X: xvec})
	}()

	// Основной цикл метода потенциалов
	for iter := 1; iter < 10_000; iter++ {
		// вычислим потенциалы u (строки) и v (столбцы)
		u := make([]float64, m)
		v := make([]float64, n)
		uKnown := make([]bool, m)
		vKnown := make([]bool, n)
		uKnown[0] = true
		u[0] = 0

		changed := true
		for changed {
			changed = false
			for ri := 0; ri < m; ri++ {
				for cj := 0; cj < n; cj++ {
					if !basic[ri][cj] {
						continue
					}
					if uKnown[ri] && !vKnown[cj] {
						v[cj] = p.Costs[ri][cj] - u[ri]
						vKnown[cj] = true
						changed = true
					}
					if !uKnown[ri] && vKnown[cj] {
						u[ri] = p.Costs[ri][cj] - v[cj]
						uKnown[ri] = true
						changed = true
					}
				}
			}
		}

		// Если некоторые потенциалы остались неизвестными (несвязный базис),
		// инициализируем каждую несвязанную компоненту, задав для одной строки
		// компоненты u=0 и повторим распространение. Это гарантирует, что
		// uKnown/vKnown будут установлены для всех строк и столбцов, чтобы
		// не пропускать редуцированные стоимости при выборе входящей клетки.
		for ri := 0; ri < m; ri++ {
			if !uKnown[ri] {
				uKnown[ri] = true
				u[ri] = 0
				// propagate for this new seed
				changed = true
				for changed {
					changed = false
					for r2 := 0; r2 < m; r2++ {
						for c2 := 0; c2 < n; c2++ {
							if !basic[r2][c2] {
								continue
							}
							if uKnown[r2] && !vKnown[c2] {
								v[c2] = p.Costs[r2][c2] - u[r2]
								vKnown[c2] = true
								changed = true
							}
							if !uKnown[r2] && vKnown[c2] {
								u[r2] = p.Costs[r2][c2] - v[c2]
								uKnown[r2] = true
								changed = true
							}
						}
					}
				}
			}
		}
		for cj := 0; cj < n; cj++ {
			if !vKnown[cj] {
				vKnown[cj] = true
				v[cj] = 0
				// propagate again
				changed = true
				for changed {
					changed = false
					for r2 := 0; r2 < m; r2++ {
						for c2 := 0; c2 < n; c2++ {
							if !basic[r2][c2] {
								continue
							}
							if uKnown[r2] && !vKnown[c2] {
								v[c2] = p.Costs[r2][c2] - u[r2]
								vKnown[c2] = true
								changed = true
							}
							if !uKnown[r2] && vKnown[c2] {
								u[r2] = p.Costs[r2][c2] - v[c2]
								uKnown[r2] = true
								changed = true
							}
						}
					}
				}
			}
		}

		// найдем наименее положительный редуцированный ценовой (cost - (u+v))
		enterI, enterJ := -1, -1
		minDelta := 0.0
		for ri := 0; ri < m; ri++ {
			for cj := 0; cj < n; cj++ {
				if basic[ri][cj] {
					continue
				}
				// если потенциалы неизвестны — считаем их бесконечно большими, пропускаем
				if !uKnown[ri] || !vKnown[cj] {
					continue
				}
				delta := p.Costs[ri][cj] - (u[ri] + v[cj])
				if delta < minDelta-eps {
					minDelta = delta
					enterI = ri
					enterJ = cj
				}
			}
		}

		// Debug: print potentials and reduced costs
		fmt.Println("MODI iteration potentials:")
		fmt.Printf("u: %v\n", u)
		fmt.Printf("v: %v\n", v)
		fmt.Println("Reduced costs (delta) for non-basic cells:")
		for ri := 0; ri < m; ri++ {
			for cj := 0; cj < n; cj++ {
				if basic[ri][cj] {
					fmt.Printf("   -- ")
					continue
				}
				if !uKnown[ri] || !vKnown[cj] {
					fmt.Printf("   N/A ")
					continue
				}
				delta := p.Costs[ri][cj] - (u[ri] + v[cj])
				fmt.Printf("%6.2f ", delta)
			}
			fmt.Println()
		}
		fmt.Printf("Chosen enter: (%d,%d) minDelta=%.6f\n", enterI, enterJ, minDelta)

		if enterI == -1 {
			// оптимум
			break
		}

		// добавляем входящую клетку в базис и ищем цикл
		basic[enterI][enterJ] = true

		// построим список базисных клеток для поиска цикла
		var cycle [][2]int
		startR, startC := enterI, enterJ

		// рекурсивный DFS поиска цикла, чередуя ход по строке и столбцу
		var dfs func(r, c int, visited map[[2]int]bool, path [][2]int, expectRow bool) bool
		dfs = func(r, c int, visited map[[2]int]bool, path [][2]int, expectRow bool) bool {
			pos := [2]int{r, c}
			if visited[pos] {
				// если вернулись в старт и путь достаточно длинный — цикл найден
				if r == startR && c == startC && len(path) >= 4 {
					cycle = append([][2]int(nil), path...)
					return true
				}
				return false
			}
			visited[pos] = true
			path = append(path, pos)

			if expectRow {
				// можем двигаться по той же строке в любую другую базисную клетку
				for nc := 0; nc < n; nc++ {
					if nc == c {
						continue
					}
					if basic[r][nc] {
						if dfs(r, nc, visited, path, !expectRow) {
							return true
						}
					}
				}
			} else {
				// двигаться по столбцу
				for nr := 0; nr < m; nr++ {
					if nr == r {
						continue
					}
					if basic[nr][c] {
						if dfs(nr, c, visited, path, !expectRow) {
							return true
						}
					}
				}
			}

			delete(visited, pos)
			return false
		}

		visited := make(map[[2]int]bool)
		_ = dfs(startR, startC, visited, nil, true)

		if len(cycle) == 0 {
			// не нашли цикл — откат и ошибку
			basic[enterI][enterJ] = false
			return nil, nil, fmt.Errorf("cannot find cycle for entering cell (%d,%d)", enterI, enterJ)
		}

		// цикл найден — определить позиции с '-' (каждый второй после первой)
		// убедимся, что цикл замкнут (последний элемент равен первому) — наш dfs возвращал путь, возможно без повторного закрытия
		if cycle[0][0] != cycle[len(cycle)-1][0] || cycle[0][1] != cycle[len(cycle)-1][1] {
			cycle = append(cycle, cycle[0])
		}

		// найдем минимальный поток на '-' позициях (индекс 1,3,5...)
		theta := math.Inf(1)
		for k := 1; k < len(cycle); k += 2 {
			r := cycle[k][0]
			c := cycle[k][1]
			if plan[r][c] < theta {
				theta = plan[r][c]
			}
		}

		// Debug: print found cycle and theta
		fmt.Printf("cycle found: ")
		for _, p := range cycle {
			fmt.Printf("(%d,%d) ", p[0], p[1])
		}
		fmt.Printf(" theta=%.6f\n", theta)
		fmt.Println("plan after update:")
		for ri := 0; ri < m; ri++ {
			for cj := 0; cj < n; cj++ {
				fmt.Printf("%8.3f", plan[ri][cj])
			}
			fmt.Println()
		}

		// изменяем потоки вдоль цикла: +theta на чётных индексах, -theta на нечётных
		var removedR, removedC int = -1, -1
		for k := 0; k < len(cycle); k++ {
			r := cycle[k][0]
			c := cycle[k][1]
			if k%2 == 0 {
				plan[r][c] += theta
			} else {
				plan[r][c] -= theta
				if math.Abs(plan[r][c]) < eps {
					// пометим для удаления из базиса позже
					removedR, removedC = r, c
				}
			}
		}

		// удалим одну базисную клетку (если есть) — предпочитаем ту, которая стала нулём и не является входящей
		if removedR >= 0 {
			if !(removedR == enterI && removedC == enterJ) {
				basic[removedR][removedC] = false
			}
		}

		// после применения theta запишем итерацию
		xvec := make([]float64, m*n)
		for ri := 0; ri < m; ri++ {
			for cj := 0; cj < n; cj++ {
				xvec[ri*n+cj] = plan[ri][cj]
			}
		}
		obj := 0.0
		for ri := 0; ri < m; ri++ {
			for cj := 0; cj < n; cj++ {
				obj += -p.Costs[ri][cj] * xvec[ri*n+cj]
			}
		}
		enterIdx := enterI*n + enterJ
		leaveIdx := -1
		if removedR >= 0 {
			leaveIdx = removedR*n + removedC
		}
		ent := 0
		lev := 0
		if enterIdx >= 0 {
			ent = enterIdx + 1
		}
		if leaveIdx >= 0 {
			lev = leaveIdx + 1
		}
		trace = append(trace, ml.IterationLP{K: iter, EnterVar: ent, LeaveVar: lev, Objective: obj, X: xvec})
		_ = iter
	}

	return plan, trace, nil
}
