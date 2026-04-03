package methods

import "math"

// GradientOptimal реализует метод наискорейшего спуска (оптимальный шаг).

type GradientOptimal struct{}

func (GradientOptimal) Name() string {
	return "Метод наискорейшего спуска (оптимальный шаг)"
}

func (GradientOptimal) Minimize2D(f Func2, grad GradFunc2, x0 Vec2, eps float64) Result2 {
	const (
		alphaMax = 2.0 // верхняя граница поиска шага
		maxIter  = 100_000
	)

	x := x0
	iter := 0

	for iter < maxIter {
		g1, g2 := grad(x.X1, x.X2)
		if Norm2(g1, g2) < eps {
			break
		}

		// φ(α) = f(x − α·g): минимизируем методом золотого сечения
		phi := func(alpha float64) float64 {
			return f(x.X1-alpha*g1, x.X2-alpha*g2)
		}
		alphaOpt := goldenLine(phi, 0, alphaMax, eps)

		x = Vec2{x.X1 - alphaOpt*g1, x.X2 - alphaOpt*g2}
		iter++
	}

	return Result2{X: x, FMin: f(x.X1, x.X2), Iterations: iter}
}

// goldenLine — метод золотого сечения для функции одной переменной на [a, b].
// Используется внутри метода наискорейшего спуска для поиска оптимального шага.
func goldenLine(phi func(float64) float64, a, b, eps float64) float64 {
	ratio := (math.Sqrt(5) - 1) / 2 // φ ≈ 0.618

	x1 := b - ratio*(b-a)
	x2 := a + ratio*(b-a)
	f1, f2 := phi(x1), phi(x2)

	for (b - a) > eps {
		if f1 > f2 {
			a = x1
			x1, f1 = x2, f2
			x2 = a + ratio*(b-a)
			f2 = phi(x2)
		} else {
			b = x2
			x2, f2 = x1, f1
			x1 = b - ratio*(b-a)
			f1 = phi(x1)
		}
	}
	return (a + b) / 2
}
