package methods

// GradientConst реализует градиентный метод с постоянным шагом (дроблением шага).
//
// Алгоритм:
//
//	На каждой итерации вычисляется градиент ∇f(xk).
//	Если ‖∇f(xk)‖ < eps — остановка (достигнута точность).
//	Пробуем шаг: x_new = xk − α · ∇f(xk), начиная с α = α₀.
//	Если f(x_new) < f(xk) — принимаем шаг и переходим к следующей итерации.
//	Иначе — дробим шаг: α ← α/2, и повторяем пробу.
type GradientConst struct {
	Alpha0 float64 // начальный шаг; если 0 — используется 0.5
}

func (m GradientConst) Name() string {
	return "Градиентный метод с дроблением шага"
}

func (m GradientConst) Minimize2D(f Func2, grad GradFunc2, x0 Vec2, eps float64) Result2 {
	alpha0 := m.Alpha0
	if alpha0 <= 0 {
		alpha0 = 0.5
	}

	const maxIter = 100_000

	x := x0
	iter := 0

	for iter < maxIter {
		g1, g2 := grad(x.X1, x.X2)
		if Norm2(g1, g2) < eps {
			break
		}

		// дробление шага: ищем наименьшее k, при котором f уменьшается
		alpha := alpha0
		fx := f(x.X1, x.X2)
		moved := false
		for alpha >= 1e-15 {
			xNew := Vec2{x.X1 - alpha*g1, x.X2 - alpha*g2}
			if f(xNew.X1, xNew.X2) < fx {
				x = xNew
				moved = true
				break
			}
			alpha /= 2
		}
		if !moved {
			break // шаг стал ничтожным — сошлись
		}
		iter++
	}

	return Result2{X: x, FMin: f(x.X1, x.X2), Iterations: iter}
}
