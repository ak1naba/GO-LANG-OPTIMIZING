package methods

// Func — тип функции одной переменной.
type Func func(float64) float64

// Result хранит результат минимизации.
type Result struct {
	XMin       float64 // точка минимума x*
	FMin       float64 // значение функции f(x*) в точке минимума
	Iterations int     // количество итераций (шагов)
}

// Optimizer — интерфейс для всех методов оптимизации.
// Каждый метод обязан реализовать:
//   - Name()     — название метода
//   - Minimize() — нахождение минимума функции f на отрезке [a, b] с точностью eps
type Optimizer interface {
	Name() string
	Minimize(f, df, d2f Func, a, b, eps float64) Result
}
