package methods

import (
	"fmt"
	"math"
)

// Newton реализует метод Ньютона для минимизации (метод второго порядка).
//
// Использует первую df(x) и вторую d2f(x) производные.
//
// Алгоритм:
//
//	Начиная с x0 = (a+b)/2, на каждом шаге:
//	    x_{k+1} = x_k − f'(x_k) / f''(x_k)
//	Применимость: f''(x) > 0 на всём [a, b] (функция выпукла).
//	Итерации пока |x_{k+1} − x_k| > eps.
type Newton struct{}

func (Newton) Name() string {
	return "Метод Ньютона (2-й порядок)"
}

func (Newton) Minimize(f, df, d2f Func, a, b, eps float64) Result {
	// можно взять серидину отрезка
	//x := (a + b) / 2
	x := 0.25
	iter := 0
	trace := make([]Iteration1D, 0, 128)

	for {
		iter++

		d1 := df(x)
		d2 := d2f(x)
		trace = append(trace, Iteration1D{
			K:    iter,
			A:    a,
			B:    b,
			X:    x,
			FX:   f(x),
			Meta: fmt.Sprintf("f'(x)=%.10f; f''(x)=%.10f", d1, d2),
		})

		if math.Abs(d2) < 1e-15 {
			break // вторая производная близка к нулю — метод неприменим
		}

		xNew := x - d1/d2

		// защита от выхода за пределы отрезка
		if xNew < a {
			xNew = a
		}
		if xNew > b {
			xNew = b
		}

		if math.Abs(xNew-x) < eps {
			x = xNew
			break
		}
		x = xNew
	}

	return Result{
		XMin:       x,
		FMin:       f(x),
		Iterations: iter,
		Trace:      trace,
	}
}
