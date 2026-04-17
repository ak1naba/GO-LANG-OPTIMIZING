package methods

import "fmt"

// PenaltyMethod реализует внешний метод штрафных функций
// для задач с ограничениями c(x1, x2) <= 0.
type PenaltyMethod struct {
	Inequalities []Constraint2D
	Mu0          float64 // начальный коэффициент штрафа; если 0 — 1
	Beta         float64 // коэффициент роста штрафа; если <= 1 — 10
	MaxOuter     int     // максимум внешних итераций; если 0 — 30
}

func (PenaltyMethod) Name() string {
	return "Метод штрафных функций"
}

func (m PenaltyMethod) Minimize2D(f Func2, _ GradFunc2, x0 Vec2, eps float64) Result2 {
	mu := m.Mu0
	if mu <= 0 {
		mu = 1.0
	}

	beta := m.Beta
	if beta <= 1 {
		beta = 10.0
	}

	maxOuter := m.MaxOuter
	if maxOuter <= 0 {
		maxOuter = 30
	}

	x := x0
	trace := make([]Iteration2D, 0, maxOuter)
	totalInner := 0

	for k := 1; k <= maxOuter; k++ {
		penalty := func(x1, x2 float64) float64 {
			sum := 0.0
			for _, c := range m.Inequalities {
				v := c(x1, x2)
				if v > 0 {
					sum += v * v
				}
			}
			return sum
		}

		phi := func(x1, x2 float64) float64 {
			return f(x1, x2) + mu*penalty(x1, x2)
		}

		gradPhi := func(x1, x2 float64) (float64, float64) {
			const h = 1e-6
			g1 := (phi(x1+h, x2) - phi(x1-h, x2)) / (2 * h)
			g2 := (phi(x1, x2+h) - phi(x1, x2-h)) / (2 * h)
			return g1, g2
		}

		inner := GradientOptimal{}
		innerRes := inner.Minimize2D(phi, gradPhi, x, eps)
		x = innerRes.X
		totalInner += innerRes.Iterations

		viol := penalty(x.X1, x.X2)
		trace = append(trace, Iteration2D{
			K:     k,
			X1:    x.X1,
			X2:    x.X2,
			FX:    f(x.X1, x.X2),
			GNorm: viol,
			Step:  mu,
			Meta:  fmt.Sprintf("inner iters=%d", innerRes.Iterations),
		})

		if viol <= eps {
			break
		}
		mu *= beta
	}

	return Result2{
		X:          x,
		FMin:       f(x.X1, x.X2),
		Iterations: totalInner,
		Trace:      trace,
	}
}
