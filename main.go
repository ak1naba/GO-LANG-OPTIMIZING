package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	fn1 "optimization/internal/fn/oneVariable"
	fn2 "optimization/internal/fn/twoVariables"
	m1 "optimization/internal/methods/oneVariable"
	m2 "optimization/internal/methods/twoVariables"
	"optimization/internal/plotter"
)

func main() {
	// Точность задаётся флагом -eps
	eps := flag.Float64("eps", fn1.Eps, "Точность вычислений (например: 1e-6, 1e-9 и т.д.)")
	flag.Parse()

	if *eps <= 0 || math.IsNaN(*eps) || math.IsInf(*eps, 0) {
		fmt.Println("Ошибка: точность должна быть положительным конечным числом")
		return
	}

	sep := func() {
		fmt.Println("─────────────────────────────────────────────────────────")
	}

	// Лабораторная работа №1
	fmt.Println("Лабораторная работа №1. Минимизация функции одной переменной")
	fmt.Println("Вариант 7: f(x) = -x³ + 3(1+x)[ln(x+1) - 1]")
	fmt.Printf("Отрезок: [%.1f; %.1f],  точность: %.0e\n\n", fn1.A, fn1.B, *eps)

	optimizers1 := []m1.Optimizer{
		m1.GoldenSection{},
		m1.Tangent{},
		m1.Newton{},
	}

	for _, opt := range optimizers1 {
		res := opt.Minimize(fn1.F, fn1.DF, fn1.D2F, fn1.A, fn1.B, *eps)
		sep()
		fmt.Printf("Метод:      %s\n", opt.Name())
		fmt.Printf("x*        = %.7f\n", res.XMin)
		fmt.Printf("f(x*)     = %.7f\n", res.FMin)
		fmt.Printf("Итераций  = %d\n", res.Iterations)
	}
	sep()

	// Графики функции одной переменной
	err := plotter.PlotFuncs(
		"Лаб. №1 · f(x), f'(x), f''(x)  на [-0.5; 0.5]",
		fn1.A, fn1.B, 500,
		"output/fn1_plot.png",
		plotter.FuncSeries{F: fn1.F, Label: "f(x)  = -x³ + 3(1+x)[ln(x+1) − 1]"},
		plotter.FuncSeries{F: fn1.DF, Label: "f'(x) = -3x² + 3·ln(x+1)"},
		plotter.FuncSeries{F: fn1.D2F, Label: "f''(x) = -6x + 3/(x+1)"},
	)
	if err != nil {
		log.Printf("Ошибка построения графика: %v", err)
	} else {
		fmt.Println("График сохранён: output/fn1_plot.png")
	}

	// Лабораторная работа №2
	fmt.Println()
	fmt.Println("Лабораторная работа №2. Минимизация функции двух переменных")
	fmt.Println("Вариант 7: f(x1,x2) = x1² + 3x2² + cos(x1+x2)")
	fmt.Printf("Начальная точка: x̄₀ = (%.1f; %.1f),  точность: %.0e\n\n", fn2.X01, fn2.X02, *eps)

	optimizers2 := []m2.Optimizer2D{
		m2.GradientConst{Alpha0: 0.5},
	}

	x0 := m2.Vec2{X1: fn2.X01, X2: fn2.X02}
	for _, opt := range optimizers2 {
		res := opt.Minimize2D(fn2.F2, fn2.GradF2, x0, *eps)
		sep()
		fmt.Printf("Метод:      %s\n", opt.Name())
		fmt.Printf("x*        = (%.7f; %.7f)\n", res.X.X1, res.X.X2)
		fmt.Printf("f(x*)     = %.7f\n", res.FMin)
		fmt.Printf("Итераций  = %d\n", res.Iterations)
	}
	sep()

	// Контурный график функции двух переменных
	err2 := plotter.PlotContour(
		"Лаб. №2 · f(x₁,x₂) = x₁² + 3x₂² + cos(x₁+x₂)",
		-2.5, 2.5, -1.5, 1.5,
		200, 20,
		"output/fn2_contour.png",
		fn2.F2,
	)
	if err2 != nil {
		log.Printf("Ошибка построения контурного графика: %v", err2)
	} else {
		fmt.Println("Контурный график сохранён: output/fn2_contour.png")
	}

}
