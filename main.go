package main

import (
	"flag"
	"fmt"
	"math"

	"optimization/internal/fn"
	"optimization/internal/methods"
)

func main() {
	// Точность задаётся флагом -eps
	eps := flag.Float64("eps", fn.Eps, "Точность вычислений (например: 1e-6, 1e-9 и т.д.)")
	flag.Parse()

	if *eps <= 0 || math.IsNaN(*eps) || math.IsInf(*eps, 0) {
		fmt.Println("Ошибка: точность должна быть положительным конечным числом")
		return
	}

	// Все три метода реализуют один интерфейс methods.Optimizer
	optimizers := []methods.Optimizer{
		methods.GoldenSection{},
		methods.Tangent{},
		methods.Newton{},
	}

	fmt.Println("Лабораторная работа №1. Минимизация функции одной переменной")
	fmt.Println("Вариант 7: f(x) = -x³ + 3(1+x)[ln(x+1) - 1]")
	fmt.Printf("Отрезок: [%.1f; %.1f],  точность: %.0e\n\n", fn.A, fn.B, *eps)

	sep := func() {
		fmt.Println("─────────────────────────────────────────────────────────")
	}

	for _, opt := range optimizers {
		res := opt.Minimize(fn.F, fn.DF, fn.D2F, fn.A, fn.B, *eps)
		sep()
		fmt.Printf("Метод:      %s\n", opt.Name())
		fmt.Printf("x*        = %.10f\n", res.XMin)
		fmt.Printf("f(x*)     = %.10f\n", res.FMin)
		fmt.Printf("Итераций  = %d\n", res.Iterations)
	}
	sep()
}
