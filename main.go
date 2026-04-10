package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	fn3 "optimization/internal/fn/linear"
	fn1 "optimization/internal/fn/oneVariable"
	fn2 "optimization/internal/fn/twoVariables"
	"optimization/internal/iterreport"
	ml "optimization/internal/methods/linear"
	m1 "optimization/internal/methods/oneVariable"
	m2 "optimization/internal/methods/twoVariables"
	"optimization/internal/plotter"
)

func main() {
	// Точность задаётся флагом -eps
	eps := flag.Float64("eps", fn1.Eps, "Точность вычислений (например: 1e-6, 1e-9 и т.д.)")
	flag.Parse()

	graphicsDir := "output/graphics"
	tablesDir := "output/tables"
	if err := os.MkdirAll(graphicsDir, 0o755); err != nil {
		log.Fatalf("Ошибка создания каталога графиков: %v", err)
	}
	if err := os.MkdirAll(tablesDir, 0o755); err != nil {
		log.Fatalf("Ошибка создания каталога таблиц: %v", err)
	}

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

	for i, opt := range optimizers1 {
		res := opt.Minimize(fn1.F, fn1.DF, fn1.D2F, fn1.A, fn1.B, *eps)
		txtName := fmt.Sprintf("%s/iter_lab1_method_%d.txt", tablesDir, i+1)
		if err := iterreport.Save1D(txtName, opt.Name(), res); err != nil {
			log.Printf("Ошибка сохранения таблицы итераций: %v", err)
		} else {
			fmt.Printf("Таблица итераций сохранена: %s\n", txtName)
		}
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
		graphicsDir+"/fn1_plot.png",
		plotter.FuncSeries{F: fn1.F, Label: "f(x)  = -x³ + 3(1+x)[ln(x+1) − 1]"},
		plotter.FuncSeries{F: fn1.DF, Label: "f'(x) = -3x² + 3·ln(x+1)"},
		plotter.FuncSeries{F: fn1.D2F, Label: "f''(x) = -6x + 3/(x+1)"},
	)
	if err != nil {
		log.Printf("Ошибка построения графика: %v", err)
	} else {
		fmt.Println("График сохранён: output/graphics/fn1_plot.png")
	}

	// Лабораторная работа №2
	fmt.Println()
	fmt.Println("Лабораторная работа №2. Минимизация функции двух переменных")
	fmt.Println("Вариант 7: f(x1,x2) = x1² + 3x2² + cos(x1+x2)")
	fmt.Printf("Начальная точка: x̄₀ = (%.1f; %.1f),  точность: %.0e\n\n", fn2.X01, fn2.X02, *eps)

	optimizers2 := []m2.Optimizer2D{
		m2.GradientConst{Alpha0: 0.5},
		m2.GradientOptimal{},
		m2.HookeJeeves{Step0: 0.559, Reduction: 0.5},
	}

	x0 := m2.Vec2{X1: fn2.X01, X2: fn2.X02}
	for i, opt := range optimizers2 {
		res := opt.Minimize2D(fn2.F2, fn2.GradF2, x0, *eps)
		txtName := fmt.Sprintf("%s/iter_lab2_method_%d.txt", tablesDir, i+1)
		if err := iterreport.Save2D(txtName, opt.Name(), res); err != nil {
			log.Printf("Ошибка сохранения таблицы итераций: %v", err)
		} else {
			fmt.Printf("Таблица итераций сохранена: %s\n", txtName)
		}
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
		graphicsDir+"/fn2_contour.png",
		fn2.F2,
	)
	if err2 != nil {
		log.Printf("Ошибка построения контурного графика: %v", err2)
	} else {
		fmt.Println("Контурный график сохранён: output/graphics/fn2_contour.png")
	}

	// 3D-график поверхности (wireframe с изометрической проекцией)
	err3 := plotter.PlotSurface(
		"Лаб. №2 · f(x₁,x₂) = x₁² + 3x₂² + cos(x₁+x₂)",
		-2.5, 2.5, -1.5, 1.5,
		35,
		graphicsDir+"/fn2_surface.png",
		fn2.F2,
	)
	if err3 != nil {
		log.Printf("Ошибка построения 3D-графика: %v", err3)
	} else {
		fmt.Println("3D-график сохранён: output/graphics/fn2_surface.png")
	}

	// Лабораторная работа №3
	fmt.Println()
	fmt.Println("Лабораторная работа №3. Линейное программирование")
	fmt.Println("Вариант: F = 3x1 - 2x2 -> max")
	fmt.Println("Ограничения:")
	fmt.Println("  2x1 + x2 <= 11")
	fmt.Println("  -3x1 + 2x2 <= 10")
	fmt.Println("  3x1 + 4x2 >= 20")
	fmt.Println("  x1, x2 >= 0")

	lp := fn3.LPVariant7()
	lpRes, errLP := ml.SolveSimplex(lp, *eps)
	if errLP != nil {
		log.Printf("Ошибка решения ЛП симплекс-методом: %v", errLP)
	} else {
		txtName := fmt.Sprintf("%s/iter_lab3_simplex.txt", tablesDir)
		if err := iterreport.SaveLP(txtName, "Двухфазный симплекс-метод", lpRes); err != nil {
			log.Printf("Ошибка сохранения таблицы ЛП: %v", err)
		} else {
			fmt.Printf("Таблица результата ЛП сохранена: %s\n", txtName)
		}

		sep()
		fmt.Printf("Метод:      %s\n", "Двухфазный симплекс-метод")
		fmt.Printf("Статус     = %s\n", lpRes.Status)
		if len(lpRes.X) > 0 {
			fmt.Printf("x*        = (%.7f; %.7f)\n", lpRes.X[0], lpRes.X[1])
		}
		fmt.Printf("F(x*)     = %.7f\n", lpRes.Objective)
		fmt.Printf("Итераций  = %d\n", lpRes.Iterations)
	}

	var lpResPtr *ml.Result
	if errLP == nil {
		lpResPtr = &lpRes
	}
	errLPPlot := plotter.PlotLP2D(
		"Лаб. №3 · Допустимая область и оптимум (симплекс)",
		graphicsDir+"/lp_lab3_feasible.png",
		lp,
		lpResPtr,
	)
	if errLPPlot != nil {
		log.Printf("Ошибка построения графика ЛП: %v", errLPPlot)
	} else {
		fmt.Println("График ЛП сохранён: output/graphics/lp_lab3_feasible.png")
	}

}
