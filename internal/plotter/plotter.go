package plotter

import (
	"fmt"
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Func — тип функции одной переменной (совместим с fn1.F, fn1.DF, fn1.D2F).
type Func func(float64) float64

// FuncSeries описывает одну кривую: функцию и подпись в легенде.
type FuncSeries struct {
	F     Func
	Label string
}

// цвета линий (matplotlib-стиль)
var lineColors = []color.RGBA{
	{R: 31, G: 119, B: 180, A: 255},  // синий
	{R: 214, G: 39, B: 40, A: 255},   // красный
	{R: 44, G: 160, B: 44, A: 255},   // зелёный
	{R: 148, G: 103, B: 189, A: 255}, // фиолетовый
	{R: 255, G: 127, B: 14, A: 255},  // оранжевый
}

// стили штриховки: сплошная, пунктир, мелкий пунктир, штрихпунктир
var dashStyles = [][]vg.Length{
	nil,
	{vg.Points(6), vg.Points(3)},
	{vg.Points(3), vg.Points(3)},
	{vg.Points(8), vg.Points(3), vg.Points(2), vg.Points(3)},
}

// PlotFuncs строит PNG-график для произвольного набора функций на отрезке [a, b].
// n — количество точек выборки; filename — путь к выходному PNG-файлу.
//
// Пример:
//
//	plotter.PlotFuncs("f(x)", fn1.A, fn1.B, 500, "output/plot.png",
//	    plotter.FuncSeries{F: fn1.F,   Label: "f(x)"},
//	    plotter.FuncSeries{F: fn1.DF,  Label: "f'(x)"},
//	    plotter.FuncSeries{F: fn1.D2F, Label: "f''(x)"},
//	)
func PlotFuncs(title string, a, b float64, n int, filename string, series ...FuncSeries) error {
	if n < 2 {
		n = 500
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"
	p.Add(plotter.NewGrid())
	p.Legend.Top = true

	step := (b - a) / float64(n-1)

	for i, s := range series {
		pts := make(plotter.XYs, n)
		for j := range pts {
			x := a + float64(j)*step
			pts[j].X = x
			pts[j].Y = s.F(x)
		}

		line, err := plotter.NewLine(pts)
		if err != nil {
			return fmt.Errorf("NewLine(%q): %w", s.Label, err)
		}

		line.LineStyle.Color = lineColors[i%len(lineColors)]
		line.LineStyle.Width = vg.Points(2)
		line.LineStyle.Dashes = dashStyles[i%len(dashStyles)]

		p.Add(line)
		p.Legend.Add(s.Label, line)
	}

	if err := p.Save(12*vg.Inch, 7*vg.Inch, filename); err != nil {
		return fmt.Errorf("Save(%q): %w", filename, err)
	}
	return nil
}

// Func2 — тип функции двух переменных (совместим с fn2.F2).
type Func2 func(x1, x2 float64) float64

// funcGrid реализует интерфейс plotter.XYZGrid для построения контурного графика.
type funcGrid struct {
	f        Func2
	x1s, x2s []float64
}

func (g funcGrid) Dims() (int, int)   { return len(g.x1s), len(g.x2s) }
func (g funcGrid) Z(c, r int) float64 { return g.f(g.x1s[c], g.x2s[r]) }
func (g funcGrid) X(c int) float64    { return g.x1s[c] }
func (g funcGrid) Y(r int) float64    { return g.x2s[r] }

// heatPalette — простая тепловая палитра от синего к красному.
type heatPalette struct{ n int }

func (p heatPalette) Colors() []color.Color {
	c := make([]color.Color, p.n)
	for i := range c {
		t := float64(i) / float64(p.n-1)
		c[i] = color.RGBA{
			R: uint8(255 * t),
			G: uint8(255 * (1 - math.Abs(2*t-1))),
			B: uint8(255 * (1 - t)),
			A: 255,
		}
	}
	return c
}

// PlotContour строит контурный PNG-график функции двух переменных f(x1, x2)
// на области [x1min, x1max] × [x2min, x2max].
// n — число точек по каждой оси; nLevels — число уровней изолиний.
func PlotContour(title string, x1min, x1max, x2min, x2max float64, n, nLevels int, filename string, f Func2) error {
	if n < 2 {
		n = 100
	}
	if nLevels < 2 {
		nLevels = 15
	}

	x1s := make([]float64, n)
	x2s := make([]float64, n)
	dx1 := (x1max - x1min) / float64(n-1)
	dx2 := (x2max - x2min) / float64(n-1)
	for i := range x1s {
		x1s[i] = x1min + float64(i)*dx1
		x2s[i] = x2min + float64(i)*dx2
	}

	g := funcGrid{f: f, x1s: x1s, x2s: x2s}

	// диапазон значений функции для расстановки уровней
	zMin, zMax := math.Inf(1), math.Inf(-1)
	for _, x1 := range x1s {
		for _, x2 := range x2s {
			z := f(x1, x2)
			if z < zMin {
				zMin = z
			}
			if z > zMax {
				zMax = z
			}
		}
	}

	levels := make([]float64, nLevels)
	for i := range levels {
		levels[i] = zMin + float64(i)*(zMax-zMin)/float64(nLevels-1)
	}

	cp := plot.New()
	cp.Title.Text = title
	cp.X.Label.Text = "x₁"
	cp.Y.Label.Text = "x₂"

	contour := plotter.NewContour(g, levels, heatPalette{nLevels})
	cp.Add(contour)

	if err := cp.Save(10*vg.Inch, 8*vg.Inch, filename); err != nil {
		return fmt.Errorf("Save(%q): %w", filename, err)
	}
	return nil
}

// PlotSurface строит PNG-график 3D-поверхности f(x1,x2) как изометрическую
// проекцию wireframe сетки. n — число узлов по каждой оси (30–40 достаточно).
func PlotSurface(title string, x1min, x1max, x2min, x2max float64, n int, filename string, f Func2) error {
	if n < 2 {
		n = 35
	}

	// вычисляем сетку значений
	x1s := make([]float64, n)
	x2s := make([]float64, n)
	dx1 := (x1max - x1min) / float64(n-1)
	dx2 := (x2max - x2min) / float64(n-1)
	for i := range x1s {
		x1s[i] = x1min + float64(i)*dx1
		x2s[i] = x2min + float64(i)*dx2
	}

	zs := make([][]float64, n)
	zMin, zMax := math.Inf(1), math.Inf(-1)
	for i := range x1s {
		zs[i] = make([]float64, n)
		for j := range x2s {
			z := f(x1s[i], x2s[j])
			zs[i][j] = z
			if z < zMin {
				zMin = z
			}
			if z > zMax {
				zMax = z
			}
		}
	}

	rangeX1 := x1max - x1min
	rangeX2 := x2max - x2min
	rangeZ := zMax - zMin
	if rangeZ == 0 {
		rangeZ = 1
	}

	// изометрическая проекция: азимут  225°, элевация 30°
	az := 225.0 * math.Pi / 180.0
	el := 30.0 * math.Pi / 180.0
	cosAz, sinAz := math.Cos(az), math.Sin(az)
	cosEl, sinEl := math.Cos(el), math.Sin(el)

	project := func(x1, x2, z float64) (float64, float64) {
		// нормализация в [0,1]
		nx := (x1 - x1min) / rangeX1
		ny := (x2 - x2min) / rangeX2
		nz := (z - zMin) / rangeZ
		// поворот вокруг вертикали (Z)
		rx := nx*cosAz - ny*sinAz
		ry := nx*sinAz + ny*cosAz
		// применение элевации (X)
		screenX := rx
		screenY := ry*sinEl + nz*cosEl
		return screenX, screenY
	}

	// цвет линии по z
	zColor := func(z float64) color.RGBA {
		t := (z - zMin) / rangeZ
		return color.RGBA{
			R: uint8(255 * t),
			G: uint8(255 * (1 - math.Abs(2*t-1))),
			B: uint8(255 * (1 - t)),
			A: 200,
		}
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = ""
	p.Y.Label.Text = ""
	p.X.Tick.Marker = plot.ConstantTicks{}
	p.Y.Tick.Marker = plot.ConstantTicks{}

	// линии вдоль x1 (фиксируем x2)
	for j := range x2s {
		pts := make(plotter.XYs, n)
		for i := range x1s {
			sx, sy := project(x1s[i], x2s[j], zs[i][j])
			pts[i].X = sx
			pts[i].Y = sy
		}
		line, err := plotter.NewLine(pts)
		if err != nil {
			return err
		}
		avgZ := 0.0
		for i := range x1s {
			avgZ += zs[i][j]
		}
		avgZ /= float64(n)
		line.LineStyle.Color = zColor(avgZ)
		line.LineStyle.Width = vg.Points(0.6)
		p.Add(line)
	}

	// линии вдоль x2 (фиксируем x1)
	for i := range x1s {
		pts := make(plotter.XYs, n)
		for j := range x2s {
			sx, sy := project(x1s[i], x2s[j], zs[i][j])
			pts[j].X = sx
			pts[j].Y = sy
		}
		line, err := plotter.NewLine(pts)
		if err != nil {
			return err
		}
		avgZ := 0.0
		for j := range x2s {
			avgZ += zs[i][j]
		}
		avgZ /= float64(n)
		line.LineStyle.Color = zColor(avgZ)
		line.LineStyle.Width = vg.Points(0.6)
		p.Add(line)
	}

	if err := p.Save(10*vg.Inch, 9*vg.Inch, filename); err != nil {
		return fmt.Errorf("Save(%q): %w", filename, err)
	}
	return nil
}
