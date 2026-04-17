package methods

import "math"

// Vec2 — вектор в ℝ².
type Vec2 struct{ X1, X2 float64 }

// Func2 — функция двух переменных.
type Func2 func(x1, x2 float64) float64

// GradFunc2 — градиент функции двух переменных; возвращает (∂f/∂x1, ∂f/∂x2).
type GradFunc2 func(x1, x2 float64) (float64, float64)

// Constraint2D задает ограничение в виде c(x1, x2) <= 0.
type Constraint2D func(x1, x2 float64) float64

// Iteration2D хранит одну строку итерационной таблицы для 2D-методов.
type Iteration2D struct {
	K     int
	X1    float64
	X2    float64
	FX    float64
	GNorm float64
	Step  float64
	Meta  string
}

// Result2 хранит результат минимизации функции двух переменных.
type Result2 struct {
	X          Vec2    // точка минимума (x1*, x2*)
	FMin       float64 // f(x1*, x2*)
	Iterations int     // количество шагов (обновлений точки)
	Trace      []Iteration2D
}

// Norm2 возвращает евклидову норму вектора (g1, g2).
func Norm2(g1, g2 float64) float64 {
	return math.Sqrt(g1*g1 + g2*g2)
}

// Optimizer2D — интерфейс для методов оптимизации функции двух переменных.
type Optimizer2D interface {
	Name() string
	Minimize2D(f Func2, grad GradFunc2, x0 Vec2, eps float64) Result2
}
