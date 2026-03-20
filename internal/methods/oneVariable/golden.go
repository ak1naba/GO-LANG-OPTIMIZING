package methods

import "math"

// GoldenSection реализует метод золотого сечения.
//
// Метод нулевого порядка: использует только значения f(x).
// Производные df и d2f игнорируются.
//
// Алгоритм:
//
//	На каждом шаге отрезок [a, b] делится двумя точками x1, x2
//	в отношении золотого сечения φ = (√5−1)/2 ≈ 0.618.
//	Та половина, где функция меньше, остаётся как новый отрезок.
//	Итерации продолжаются пока |b−a| > eps.
type GoldenSection struct{}

func (GoldenSection) Name() string {
	return "Метод золотого сечения (0-й порядок)"
}

func (GoldenSection) Minimize(f, _, _ Func, a, b, eps float64) Result {
	ratio := (math.Sqrt(5) - 1) / 2 // φ ≈ 0.618

	x1 := b - ratio*(b-a)
	x2 := a + ratio*(b-a)
	f1 := f(x1)
	f2 := f(x2)

	iter := 0
	for (b - a) > eps {
		iter++
		if f1 > f2 {
			// минимум правее x1, сужаем левую границу
			a = x1
			x1, f1 = x2, f2
			x2 = a + ratio*(b-a)
			f2 = f(x2)
		} else {
			// минимум левее x2, сужаем правую границу
			b = x2
			x2, f2 = x1, f1
			x1 = b - ratio*(b-a)
			f1 = f(x1)
		}
	}

	xMin := (a + b) / 2
	return Result{
		XMin:       xMin,
		FMin:       f(xMin),
		Iterations: iter,
	}
}
