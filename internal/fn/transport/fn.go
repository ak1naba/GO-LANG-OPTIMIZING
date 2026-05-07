package fn

import mt "optimization/internal/methods/transport"

// TransportVariant7 возвращает сбалансированную транспортную задачу из варианта 7.
//
// Матрица стоимостей:
//
//	2  4  5  1
//	2  3  9  4
//	3  4  2  5
//
// Запасы: 60, 70, 20
// Потребности: 40, 30, 30, 50
func TransportVariant7() mt.Problem {
	return mt.Problem{
		Costs: [][]float64{
			{2, 4, 5, 1},
			{2, 3, 9, 4},
			{3, 4, 2, 5},
		},
		Supply: []float64{60, 70, 20},
		Demand: []float64{40, 30, 30, 50},
	}
}
