package methods

import "fmt"

// HookeJeeves реализует метод Хука–Дживса (поисковый метод без градиента).

type HookeJeeves struct {
	Step0     float64 // начальный шаг по координатам; если 0 — используется 0.5
	Reduction float64 // коэффициент уменьшения шага (0,1); если некорректен — 0.5
	MaxIter   int     // ограничение числа обновлений точки; если 0 — 100000
}

func (HookeJeeves) Name() string {
	return "Метод Хука–Дживса"
}

func (m HookeJeeves) Minimize2D(f Func2, _ GradFunc2, x0 Vec2, eps float64) Result2 {
	step := m.Step0
	if step <= 0 {
		step = 0.5
	}

	reduction := m.Reduction
	if reduction <= 0 || reduction >= 1 {
		reduction = 0.5
	}

	maxIter := m.MaxIter
	if maxIter <= 0 {
		maxIter = 100_000
	}

	base := x0
	iter := 0
	trace := make([]Iteration2D, 0, 512)

	for iter < maxIter && step >= eps {
		iter++
		explored, improved := hookeExplore(f, base, step)
		meta := "reduced step"
		if !improved {
			trace = append(trace, Iteration2D{
				K:     iter,
				X1:    base.X1,
				X2:    base.X2,
				FX:    f(base.X1, base.X2),
				GNorm: 0,
				Step:  step,
				Meta:  meta,
			})
			step *= reduction
			continue
		}

		meta = "explore improved"

		for iter < maxIter {
			// pattern move: продолжаем движение в успешном направлении.
			pattern := Vec2{
				X1: explored.X1 + (explored.X1 - base.X1),
				X2: explored.X2 + (explored.X2 - base.X2),
			}

			patternExplored, patternImproved := hookeExplore(f, pattern, step)
			if patternImproved && f(patternExplored.X1, patternExplored.X2) < f(explored.X1, explored.X2) {
				base = explored
				explored = patternExplored
				meta = "pattern improved"
				continue
			}

			base = explored
			trace = append(trace, Iteration2D{
				K:     iter,
				X1:    base.X1,
				X2:    base.X2,
				FX:    f(base.X1, base.X2),
				GNorm: 0,
				Step:  step,
				Meta:  fmt.Sprintf("%s; accepted new base", meta),
			})
			break
		}
	}

	return Result2{X: base, FMin: f(base.X1, base.X2), Iterations: iter, Trace: trace}
}

func hookeExplore(f Func2, start Vec2, step float64) (Vec2, bool) {
	x := start
	improvedAny := false

	x, improved := exploreCoord(f, x, step, true)
	if improved {
		improvedAny = true
	}

	x, improved = exploreCoord(f, x, step, false)
	if improved {
		improvedAny = true
	}

	return x, improvedAny
}

func exploreCoord(f Func2, x Vec2, step float64, byX1 bool) (Vec2, bool) {
	fx := f(x.X1, x.X2)

	plus := x
	minus := x
	if byX1 {
		plus.X1 += step
		minus.X1 -= step
	} else {
		plus.X2 += step
		minus.X2 -= step
	}

	fPlus := f(plus.X1, plus.X2)
	fMinus := f(minus.X1, minus.X2)

	if fPlus < fx && fPlus <= fMinus {
		return plus, true
	}
	if fMinus < fx {
		return minus, true
	}

	return x, false
}
