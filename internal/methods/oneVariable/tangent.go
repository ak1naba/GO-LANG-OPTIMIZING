package methods

import "math"

// Tangent реализует метод касательных (метод первого порядка).
//
// Использует первую производную df(x).
// Вторая производная d2f игнорируется.
//
// Алгоритм (применяется к нахождению нуля f'(x)):
//
//	На отрезке [a, b] проводятся касательные к f(x) в точках a и b.
//	Их пересечение даёт новое приближение x*:
//	    x* = ( f(b)−f(a) + f'(a)·a − f'(b)·b ) / ( f'(a)−f'(b) )
//	Знак f'(x*) определяет, какая граница обновляется:
//	  f'(x*) < 0  →  a = x*   (минимум правее)
//	  f'(x*) ≥ 0  →  b = x*   (минимум левее)
//	Итерации пока |b−a| > eps.
type Tangent struct{}

func (Tangent) Name() string {
	return "Метод касательных (1-й порядок)"
}

func (Tangent) Minimize(f, df, _ Func, a, b, eps float64) Result {
	iter := 0
	for (b - a) > eps {
		iter++

		fa, fb := df(a), df(b)
		denom := fa - fb
		if math.Abs(denom) < 1e-15 {
			break // производные совпали — деление нестабильно
		}

		// пересечение касательных
		xNew := (f(b) - f(a) + fa*a - fb*b) / denom

		// защита от выхода за границы из-за численных ошибок
		if xNew < a {
			xNew = a
		}
		if xNew > b {
			xNew = b
		}

		if df(xNew) < 0 {
			a = xNew // минимум справа от xNew
		} else {
			b = xNew // минимум слева от xNew
		}
	}

	xMin := (a + b) / 2
	return Result{
		XMin:       xMin,
		FMin:       f(xMin),
		Iterations: iter,
	}
}
