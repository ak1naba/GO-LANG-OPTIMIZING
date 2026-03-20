package fn

import "math"

// Функция: f(x) = -x³ + 3(1+x)[ln(x+1) - 1],  x ∈ [-0.5; 0.5]
// f'(x)  = -3x² + 3·ln(x+1)
// f''(x) = -6x  + 3/(x+1)

// A и B — границы отрезка поиска. Eps — точность по умолчанию (1e-4).
const (
	A   = -0.5
	B   = 0.5
	Eps = 1e-4
)

// F вычисляет значение целевой функции f(x).
func F(x float64) float64 {
	return -x*x*x + 3*(1+x)*(math.Log(x+1)-1)
}

// DF вычисляет первую производную f'(x).
func DF(x float64) float64 {
	return -3*x*x + 3*math.Log(x+1)
}

// D2F вычисляет вторую производную f”(x).
func D2F(x float64) float64 {
	return -6*x + 3/(x+1)
}
