package fn

import "optimization/internal/methods/linear"

// LPVariant7 возвращает задачу ЛП (вариант 7):
//
//	F = 3x1 - 2x2 -> max
//
// при ограничениях:
//
//	2x1 + x2 <= 11
//	-3x1 + 2x2 <= 10
//	3x1 + 4x2 >= 20
//	x1, x2 >= 0
func LPVariant7() linear.Problem {
	return linear.Problem{
		C: []float64{3, -2},
		A: [][]float64{
			{2, 1},
			{-3, 2},
			{3, 4},
		},
		B: []float64{11, 10, 20},
		Sense: []linear.ConstraintSense{
			linear.SenseLE,
			linear.SenseLE,
			linear.SenseGE,
		},
	}
}
