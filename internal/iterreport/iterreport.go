package iterreport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ml "optimization/internal/methods/linear"
	m1 "optimization/internal/methods/oneVariable"
	m2 "optimization/internal/methods/twoVariables"
)

// Save1D сохраняет итерационную таблицу метода одной переменной в txt.
func Save1D(filename, methodName string, res m1.Result) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(filename), err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create %q: %w", filename, err)
	}
	defer f.Close()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Метод: %s\n", methodName))
	b.WriteString(fmt.Sprintf("x* = %.10f\n", res.XMin))
	b.WriteString(fmt.Sprintf("f(x*) = %.10f\n", res.FMin))
	b.WriteString(fmt.Sprintf("Итераций = %d\n\n", res.Iterations))

	b.WriteString("k\ta\tb\tx\tf(x)\tmeta\n")
	for _, it := range res.Trace {
		b.WriteString(fmt.Sprintf("%d\t%.10f\t%.10f\t%.10f\t%.10f\t%s\n",
			it.K, it.A, it.B, it.X, it.FX, it.Meta,
		))
	}

	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("write %q: %w", filename, err)
	}
	return nil
}

// Save2D сохраняет итерационную таблицу метода двух переменных в txt.
func Save2D(filename, methodName string, res m2.Result2) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(filename), err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create %q: %w", filename, err)
	}
	defer f.Close()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Метод: %s\n", methodName))
	b.WriteString(fmt.Sprintf("x* = (%.10f, %.10f)\n", res.X.X1, res.X.X2))
	b.WriteString(fmt.Sprintf("f(x*) = %.10f\n", res.FMin))
	b.WriteString(fmt.Sprintf("Итераций = %d\n\n", res.Iterations))

	b.WriteString("k\tx1\tx2\tf(x)\t||grad||\tstep\tmeta\n")
	for _, it := range res.Trace {
		b.WriteString(fmt.Sprintf("%d\t%.10f\t%.10f\t%.10f\t%.10f\t%.10f\t%s\n",
			it.K, it.X1, it.X2, it.FX, it.GNorm, it.Step, it.Meta,
		))
	}

	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("write %q: %w", filename, err)
	}
	return nil
}

// SaveLP сохраняет результат решения задачи линейного программирования в txt.
func SaveLP(filename, methodName string, res ml.Result) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(filename), err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create %q: %w", filename, err)
	}
	defer f.Close()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Метод: %s\n", methodName))
	b.WriteString(fmt.Sprintf("Статус: %s\n", res.Status))
	b.WriteString(fmt.Sprintf("Итераций = %d\n", res.Iterations))
	b.WriteString(fmt.Sprintf("F* = %.10f\n", res.Objective))

	if len(res.X) > 0 {
		b.WriteString("\nРешение:\n")
		for i, v := range res.X {
			b.WriteString(fmt.Sprintf("x%d = %.10f\n", i+1, v))
		}
	}

	varCount := len(res.X)
	if varCount == 0 {
		for _, it := range res.Trace {
			if len(it.X) > varCount {
				varCount = len(it.X)
			}
		}
	}

	if len(res.Trace) > 0 {
		b.WriteString("\nИтерационная таблица:\n")
		b.WriteString("k\tphase\tenter\tleave\tF")
		for i := 0; i < varCount; i++ {
			b.WriteString(fmt.Sprintf("\tx%d", i+1))
		}
		b.WriteString("\n")

		for _, it := range res.Trace {
			enter := "-"
			leave := "-"
			if it.EnterVar > 0 {
				enter = fmt.Sprintf("x%d", it.EnterVar)
			}
			if it.LeaveVar > 0 {
				leave = fmt.Sprintf("x%d", it.LeaveVar)
			}

			b.WriteString(fmt.Sprintf("%d\t%s\t%s\t%s\t%.10f", it.K, it.Phase, enter, leave, it.Objective))
			for i := 0; i < varCount; i++ {
				v := 0.0
				if i < len(it.X) {
					v = it.X[i]
				}
				b.WriteString(fmt.Sprintf("\t%.10f", v))
			}
			b.WriteString("\n")
		}
	}

	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("write %q: %w", filename, err)
	}
	return nil
}
