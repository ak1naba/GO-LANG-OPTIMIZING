package fn

import "math"

// Функция двух переменных (Вариант 7):
//   f(x1, x2) = x1² + 3x2² + cos(x1 + x2)
//
// Градиент:
//   ∂f/∂x1 = 2x1 − sin(x1 + x2)
//   ∂f/∂x2 = 6x2 − sin(x1 + x2)

// X01, X02 — начальная точка x̄₀ = (1; 1). Eps — точность по умолчанию.
const (
	X01 = 1.0
	X02 = 1.0
	Eps = 1e-4
)

// F2 вычисляет значение f(x1, x2).
func F2(x1, x2 float64) float64 {
	return x1*x1 + 3*x2*x2 + math.Cos(x1+x2)
}

// GradF2 возвращает (∂f/∂x1, ∂f/∂x2).
func GradF2(x1, x2 float64) (float64, float64) {
	s := math.Sin(x1 + x2)
	return 2*x1 - s, 6*x2 - s
}
