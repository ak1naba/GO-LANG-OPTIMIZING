package linear

import (
	"errors"
	"fmt"
	"math"
)

type ConstraintSense string

const (
	SenseLE ConstraintSense = "<="
	SenseGE ConstraintSense = ">="
	SenseEQ ConstraintSense = "="
)

// Problem описывает задачу линейного программирования в канонической форме:
// max c^T x, при ограничениях A*x <= b, x >= 0.
type Problem struct {
	C []float64
	A [][]float64
	B []float64
	// Sense определяет тип каждого ограничения: <=, >= или =.
	// Если поле пустое, ограничения считаются типа <= (обратная совместимость).
	Sense []ConstraintSense
}

// Result хранит результат работы симплекс-метода.
type Result struct {
	X          []float64
	Objective  float64
	Iterations int
	Status     string
	Trace      []IterationLP
}

// IterationLP хранит одну строку итерационной таблицы симплекс-метода.
type IterationLP struct {
	K         int
	Phase     string
	EnterVar  int
	LeaveVar  int
	Objective float64
	X         []float64
}

const (
	statusOptimal    = "optimal"
	statusUnbounded  = "unbounded"
	statusInfeasible = "infeasible"

	maxIter = 10_000
)

// SolveSimplex решает задачу ЛП методом двухфазного симплекс-метода.
// Поддерживаются ограничения <=, >=, = и произвольные знаки правых частей.
func SolveSimplex(p Problem, eps float64) (Result, error) {
	if eps <= 0 || math.IsNaN(eps) || math.IsInf(eps, 0) {
		return Result{}, errors.New("eps must be positive finite number")
	}

	m := len(p.B)
	n := len(p.C)
	if m == 0 || n == 0 {
		return Result{}, errors.New("empty LP problem")
	}
	if len(p.A) != m {
		return Result{}, fmt.Errorf("A rows mismatch: got %d, expected %d", len(p.A), m)
	}
	for i := 0; i < m; i++ {
		if len(p.A[i]) != n {
			return Result{}, fmt.Errorf("A[%d] cols mismatch: got %d, expected %d", i, len(p.A[i]), n)
		}
	}

	sense := make([]ConstraintSense, m)
	if len(p.Sense) == 0 {
		for i := 0; i < m; i++ {
			sense[i] = SenseLE
		}
	} else {
		if len(p.Sense) != m {
			return Result{}, fmt.Errorf("Sense len mismatch: got %d, expected %d", len(p.Sense), m)
		}
		copy(sense, p.Sense)
	}

	A := make([][]float64, m)
	B := make([]float64, m)
	for i := 0; i < m; i++ {
		A[i] = make([]float64, n)
		copy(A[i], p.A[i])
		B[i] = p.B[i]

		if B[i] < -eps {
			for j := 0; j < n; j++ {
				A[i][j] = -A[i][j]
			}
			B[i] = -B[i]
			sense[i] = flipSense(sense[i])
		}
		if sense[i] != SenseLE && sense[i] != SenseGE && sense[i] != SenseEQ {
			return Result{}, fmt.Errorf("unsupported constraint sense at row %d: %q", i, sense[i])
		}
	}

	slackCount := 0
	surplusCount := 0
	artCount := 0
	for i := 0; i < m; i++ {
		switch sense[i] {
		case SenseLE:
			slackCount++
		case SenseGE:
			surplusCount++
			artCount++
		case SenseEQ:
			artCount++
		}
	}

	totalVars := n + slackCount + surplusCount + artCount
	rows := m + 1
	cols := totalVars + 1 // последняя колонка — правая часть

	tab := make([][]float64, rows)
	for i := range tab {
		tab[i] = make([]float64, cols)
	}
	basis := make([]int, m)
	isArtificial := make([]bool, totalVars)

	nextSlack := n
	nextSurplus := n + slackCount
	nextArt := n + slackCount + surplusCount

	for i := 0; i < m; i++ {
		copy(tab[i][:n], A[i])

		switch sense[i] {
		case SenseLE:
			tab[i][nextSlack] = 1
			basis[i] = nextSlack
			nextSlack++
		case SenseGE:
			tab[i][nextSurplus] = -1
			tab[i][nextArt] = 1
			basis[i] = nextArt
			isArtificial[nextArt] = true
			nextSurplus++
			nextArt++
		case SenseEQ:
			tab[i][nextArt] = 1
			basis[i] = nextArt
			isArtificial[nextArt] = true
			nextArt++
		}

		tab[i][cols-1] = B[i]
	}

	iter := 0
	trace := make([]IterationLP, 0, 32)
	if artCount > 0 {
		// Фаза I: max(-sum(a_i)), где a_i — искусственные переменные.
		for j := 0; j < totalVars; j++ {
			if isArtificial[j] {
				tab[m][j] = 1
			}
		}
		for i := 0; i < m; i++ {
			coef := tab[m][basis[i]]
			if math.Abs(coef) <= eps {
				continue
			}
			for j := 0; j < cols; j++ {
				tab[m][j] -= coef * tab[i][j]
			}
		}
		trace = appendTrace(trace, iter, "I", -1, -1, tab, basis, m, n, cols-1)

		for iter < maxIter {
			enterCol := chooseEntering(tab[m], cols-1, eps)
			if enterCol == -1 {
				break
			}

			leaveRow := chooseLeaving(tab, enterCol, m, cols-1, eps)
			if leaveRow == -1 {
				return Result{}, errors.New("phase I became unbounded unexpectedly")
			}

			leaveVar := basis[leaveRow]
			pivot(tab, leaveRow, enterCol, rows, cols)
			basis[leaveRow] = enterCol
			iter++
			trace = appendTrace(trace, iter, "I", enterCol, leaveVar, tab, basis, m, n, cols-1)
		}
		if iter >= maxIter {
			return Result{}, errors.New("simplex reached iteration limit in phase I")
		}

		if tab[m][cols-1] < -eps {
			return Result{
				X:          nil,
				Objective:  math.Inf(-1),
				Iterations: iter,
				Status:     statusInfeasible,
				Trace:      trace,
			}, nil
		}

		for i := 0; i < m; i++ {
			if !isArtificial[basis[i]] {
				continue
			}
			pivotCol := -1
			for j := 0; j < totalVars; j++ {
				if isArtificial[j] {
					continue
				}
				if math.Abs(tab[i][j]) > eps {
					pivotCol = j
					break
				}
			}
			if pivotCol != -1 {
				pivot(tab, i, pivotCol, rows, cols)
				basis[i] = pivotCol
			}
		}

		keepCols := make([]int, 0, totalVars)
		colMap := make([]int, totalVars)
		for j := 0; j < totalVars; j++ {
			if isArtificial[j] {
				colMap[j] = -1
				continue
			}
			colMap[j] = len(keepCols)
			keepCols = append(keepCols, j)
		}

		newCols := len(keepCols) + 1
		tab2 := make([][]float64, rows)
		for i := 0; i < rows; i++ {
			tab2[i] = make([]float64, newCols)
			for newJ, oldJ := range keepCols {
				tab2[i][newJ] = tab[i][oldJ]
			}
			tab2[i][newCols-1] = tab[i][cols-1]
		}

		for i := 0; i < m; i++ {
			mapped := colMap[basis[i]]
			if mapped == -1 {
				return Result{}, errors.New("failed to eliminate artificial variable from basis")
			}
			basis[i] = mapped
		}

		tab = tab2
		cols = newCols
		totalVars = len(keepCols)
	}

	for j := 0; j < cols; j++ {
		tab[m][j] = 0
	}
	for j := 0; j < n; j++ {
		tab[m][j] = -p.C[j]
	}
	for i := 0; i < m; i++ {
		coef := tab[m][basis[i]]
		if math.Abs(coef) <= eps {
			continue
		}
		for j := 0; j < cols; j++ {
			tab[m][j] -= coef * tab[i][j]
		}
	}
	trace = appendTrace(trace, iter, "II", -1, -1, tab, basis, m, n, cols-1)

	for iter < maxIter {
		enterCol := chooseEntering(tab[m], cols-1, eps)
		if enterCol == -1 {
			x := make([]float64, n)
			for i := 0; i < m; i++ {
				if basis[i] < n {
					x[basis[i]] = tab[i][cols-1]
				}
			}

			return Result{
				X:          x,
				Objective:  tab[m][cols-1],
				Iterations: iter,
				Status:     statusOptimal,
				Trace:      trace,
			}, nil
		}

		leaveRow := chooseLeaving(tab, enterCol, m, cols-1, eps)
		if leaveRow == -1 {
			return Result{
				X:          nil,
				Objective:  math.Inf(1),
				Iterations: iter,
				Status:     statusUnbounded,
				Trace:      trace,
			}, nil
		}

		leaveVar := basis[leaveRow]
		pivot(tab, leaveRow, enterCol, rows, cols)
		basis[leaveRow] = enterCol
		iter++
		trace = appendTrace(trace, iter, "II", enterCol, leaveVar, tab, basis, m, n, cols-1)
	}

	return Result{}, errors.New("simplex reached iteration limit")
}

func appendTrace(trace []IterationLP, iter int, phase string, enterCol, leaveVar int,
	tab [][]float64, basis []int, m, n, rhsCol int,
) []IterationLP {
	x := extractPrimal(tab, basis, m, n, rhsCol)
	ent := 0
	lev := 0
	if enterCol >= 0 {
		ent = enterCol + 1
	}
	if leaveVar >= 0 {
		lev = leaveVar + 1
	}

	trace = append(trace, IterationLP{
		K:         iter,
		Phase:     phase,
		EnterVar:  ent,
		LeaveVar:  lev,
		Objective: tab[m][rhsCol],
		X:         x,
	})
	return trace
}

func extractPrimal(tab [][]float64, basis []int, m, n, rhsCol int) []float64 {
	x := make([]float64, n)
	for i := 0; i < m; i++ {
		if basis[i] < n {
			x[basis[i]] = tab[i][rhsCol]
		}
	}
	return x
}

func flipSense(s ConstraintSense) ConstraintSense {
	switch s {
	case SenseLE:
		return SenseGE
	case SenseGE:
		return SenseLE
	default:
		return s
	}
}

func chooseEntering(objRow []float64, lastCol int, eps float64) int {
	col := -1
	minVal := -eps

	for j := 0; j < lastCol; j++ {
		if objRow[j] < minVal {
			minVal = objRow[j]
			col = j
		}
	}

	return col
}

func chooseLeaving(tab [][]float64, enterCol, m, rhsCol int, eps float64) int {
	row := -1
	minRatio := math.Inf(1)

	for i := 0; i < m; i++ {
		a := tab[i][enterCol]
		if a <= eps {
			continue
		}

		ratio := tab[i][rhsCol] / a
		if ratio < minRatio-eps {
			minRatio = ratio
			row = i
		}
	}

	return row
}

func pivot(tab [][]float64, pivotRow, pivotCol, rows, cols int) {
	pivotVal := tab[pivotRow][pivotCol]

	for j := 0; j < cols; j++ {
		tab[pivotRow][j] /= pivotVal
	}

	for i := 0; i < rows; i++ {
		if i == pivotRow {
			continue
		}

		factor := tab[i][pivotCol]
		if factor == 0 {
			continue
		}

		for j := 0; j < cols; j++ {
			tab[i][j] -= factor * tab[pivotRow][j]
		}
	}
}
