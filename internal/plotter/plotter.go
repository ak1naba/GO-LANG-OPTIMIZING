package plotter

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	ml "optimization/internal/methods/linear"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
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

// PlotConstrainedContour строит контурный график функции f(x1, x2)
// и наносит линейные ограничения в виде a1*x1 + a2*x2 (<=|>=|=) b,
// а также допустимую область и найденную точку x*.
func PlotConstrainedContour(
	title, filename string,
	x1min, x1max, x2min, x2max float64,
	n, nLevels int,
	f Func2,
	A [][]float64,
	B []float64,
	sense []ml.ConstraintSense,
	xStarX, xStarY float64,
) error {
	if len(A) != len(B) {
		return fmt.Errorf("A rows mismatch with B: %d vs %d", len(A), len(B))
	}
	if len(A) == 0 {
		return fmt.Errorf("empty constraints")
	}

	if n < 2 {
		n = 100
	}
	if nLevels < 2 {
		nLevels = 15
	}

	consSense := make([]ml.ConstraintSense, len(B))
	if len(sense) == 0 {
		for i := range consSense {
			consSense[i] = ml.SenseLE
		}
	} else {
		if len(sense) != len(B) {
			return fmt.Errorf("Sense len mismatch: %d vs %d", len(sense), len(B))
		}
		copy(consSense, sense)
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

	const eps = 1e-8
	isFeasible := func(x, y float64) bool {
		for i := range A {
			if len(A[i]) != 2 {
				return false
			}
			lhs := A[i][0]*x + A[i][1]*y
			s := consSense[i]
			switch s {
			case ml.SenseLE:
				if lhs-B[i] > eps {
					return false
				}
			case ml.SenseGE:
				if B[i]-lhs > eps {
					return false
				}
			case ml.SenseEQ:
				if math.Abs(lhs-B[i]) > eps {
					return false
				}
			default:
				return false
			}
		}
		return true
	}

	type line2 struct {
		a, b, c float64
	}
	intersect := func(l1, l2 line2) (plotter.XY, bool) {
		det := l1.a*l2.b - l2.a*l1.b
		if math.Abs(det) <= eps {
			return plotter.XY{}, false
		}
		x := (l1.c*l2.b - l2.c*l1.b) / det
		y := (l1.a*l2.c - l2.a*l1.c) / det
		if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
			return plotter.XY{}, false
		}
		return plotter.XY{X: x, Y: y}, true
	}

	lines := make([]line2, 0, len(A))
	for i := range A {
		if len(A[i]) != 2 {
			return fmt.Errorf("A[%d] must have 2 columns, got %d", i, len(A[i]))
		}
		lines = append(lines, line2{a: A[i][0], b: A[i][1], c: B[i]})
	}

	unique := make([]plotter.XY, 0)
	addUnique := func(p plotter.XY) {
		for _, q := range unique {
			if math.Abs(p.X-q.X) <= 1e-6 && math.Abs(p.Y-q.Y) <= 1e-6 {
				return
			}
		}
		unique = append(unique, p)
	}

	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			p, ok := intersect(lines[i], lines[j])
			if !ok {
				continue
			}
			if isFeasible(p.X, p.Y) {
				addUnique(p)
			}
		}
	}

	cp := plot.New()
	cp.Title.Text = title
	cp.X.Label.Text = "x1"
	cp.Y.Label.Text = "x2"
	cp.X.Min, cp.X.Max = x1min, x1max
	cp.Y.Min, cp.Y.Max = x2min, x2max
	cp.Add(plotter.NewGrid())

	contour := plotter.NewContour(g, levels, heatPalette{nLevels})
	cp.Add(contour)

	if len(unique) >= 3 {
		feasibleHull := convexHull(unique)
		if len(feasibleHull) >= 3 {
			poly, err := plotter.NewPolygon(feasibleHull)
			if err != nil {
				return fmt.Errorf("NewPolygon(feasible): %w", err)
			}
			poly.Color = color.RGBA{R: 138, G: 201, B: 120, A: 90}
			poly.LineStyle.Color = color.RGBA{R: 38, G: 70, B: 83, A: 220}
			poly.LineStyle.Width = vg.Points(1.1)
			cp.Add(poly)
			cp.Legend.Add("допустимая область", poly)
		}
	}

	for i := range A {
		linePts, ok := constraintSegment(A[i][0], A[i][1], B[i], x1min, x1max, x2min, x2max)
		if !ok {
			continue
		}
		ln, err := plotter.NewLine(linePts)
		if err != nil {
			return fmt.Errorf("NewLine(constraint %d): %w", i+1, err)
		}
		ln.LineStyle.Width = vg.Points(1.6)
		ln.LineStyle.Color = lineColors[i%len(lineColors)]
		ln.LineStyle.Dashes = dashStyles[(i+1)%len(dashStyles)]
		cp.Add(ln)
		cp.Legend.Add(fmt.Sprintf("огр. %d", i+1), ln)
	}

	optPts := plotter.XYs{{X: xStarX, Y: xStarY}}
	sc, err := plotter.NewScatter(optPts)
	if err != nil {
		return fmt.Errorf("NewScatter(opt): %w", err)
	}
	sc.Shape = draw.CrossGlyph{}
	sc.Radius = vg.Points(5)
	sc.Color = color.RGBA{R: 200, G: 40, B: 40, A: 255}
	cp.Add(sc)
	cp.Legend.Add("x* (штрафной метод)", sc)

	if err := cp.Save(11*vg.Inch, 8*vg.Inch, filename); err != nil {
		return fmt.Errorf("Save(%q): %w", filename, err)
	}
	return nil
}

// PlotLP2D строит график задачи ЛП в 2D:
// границы ограничений, допустимую область, оптимальную точку и линию уровня цели.
func PlotLP2D(title, filename string, prob ml.Problem, res *ml.Result) error {
	if len(prob.C) != 2 {
		return fmt.Errorf("PlotLP2D supports only 2 variables, got %d", len(prob.C))
	}
	if len(prob.A) != len(prob.B) {
		return fmt.Errorf("A rows mismatch with B: %d vs %d", len(prob.A), len(prob.B))
	}

	sense := make([]ml.ConstraintSense, len(prob.B))
	if len(prob.Sense) == 0 {
		for i := range sense {
			sense[i] = ml.SenseLE
		}
	} else {
		if len(prob.Sense) != len(prob.B) {
			return fmt.Errorf("Sense len mismatch: %d vs %d", len(prob.Sense), len(prob.B))
		}
		copy(sense, prob.Sense)
	}

	type line2 struct {
		a, b, c float64
	}

	lines := make([]line2, 0, len(prob.A)+2)
	for i := range prob.A {
		if len(prob.A[i]) != 2 {
			return fmt.Errorf("A[%d] must have 2 columns, got %d", i, len(prob.A[i]))
		}
		lines = append(lines, line2{a: prob.A[i][0], b: prob.A[i][1], c: prob.B[i]})
	}
	// x1 = 0, x2 = 0
	lines = append(lines, line2{a: 1, b: 0, c: 0}, line2{a: 0, b: 1, c: 0})

	const eps = 1e-8
	isFeasible := func(x, y float64) bool {
		if x < -eps || y < -eps {
			return false
		}
		for i := range prob.A {
			lhs := prob.A[i][0]*x + prob.A[i][1]*y
			s := sense[i]
			switch s {
			case ml.SenseLE:
				if lhs-prob.B[i] > eps {
					return false
				}
			case ml.SenseGE:
				if prob.B[i]-lhs > eps {
					return false
				}
			case ml.SenseEQ:
				if math.Abs(lhs-prob.B[i]) > eps {
					return false
				}
			default:
				return false
			}
		}
		return true
	}

	intersect := func(l1, l2 line2) (plotter.XY, bool) {
		det := l1.a*l2.b - l2.a*l1.b
		if math.Abs(det) <= eps {
			return plotter.XY{}, false
		}
		x := (l1.c*l2.b - l2.c*l1.b) / det
		y := (l1.a*l2.c - l2.a*l1.c) / det
		if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
			return plotter.XY{}, false
		}
		return plotter.XY{X: x, Y: y}, true
	}

	unique := make([]plotter.XY, 0)
	addUnique := func(p plotter.XY) {
		for _, q := range unique {
			if math.Abs(p.X-q.X) <= 1e-6 && math.Abs(p.Y-q.Y) <= 1e-6 {
				return
			}
		}
		unique = append(unique, p)
	}

	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			p, ok := intersect(lines[i], lines[j])
			if !ok {
				continue
			}
			if isFeasible(p.X, p.Y) {
				addUnique(p)
			}
		}
	}

	if len(unique) == 0 {
		return fmt.Errorf("no feasible points for plotting")
	}

	feasibleHull := convexHull(unique)
	if len(feasibleHull) < 3 {
		return fmt.Errorf("feasible set is degenerate and cannot be filled")
	}

	xMin, xMax := feasibleHull[0].X, feasibleHull[0].X
	yMin, yMax := feasibleHull[0].Y, feasibleHull[0].Y
	for _, p := range feasibleHull {
		if p.X < xMin {
			xMin = p.X
		}
		if p.X > xMax {
			xMax = p.X
		}
		if p.Y < yMin {
			yMin = p.Y
		}
		if p.Y > yMax {
			yMax = p.Y
		}
	}

	if res != nil && len(res.X) >= 2 {
		if res.X[0] > xMax {
			xMax = res.X[0]
		}
		if res.X[1] > yMax {
			yMax = res.X[1]
		}
	}

	xPad := 0.12 * math.Max(1, xMax-xMin)
	yPad := 0.12 * math.Max(1, yMax-yMin)
	xMin = math.Max(0, xMin-xPad)
	yMin = math.Max(0, yMin-yPad)
	xMax += xPad
	yMax += yPad

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "x1"
	p.Y.Label.Text = "x2"
	p.Add(plotter.NewGrid())
	p.X.Min, p.X.Max = xMin, xMax
	p.Y.Min, p.Y.Max = yMin, yMax

	polyPts := make(plotter.XYs, len(feasibleHull))
	copy(polyPts, feasibleHull)
	poly, err := plotter.NewPolygon(polyPts)
	if err != nil {
		return fmt.Errorf("NewPolygon(feasible): %w", err)
	}
	poly.Color = color.RGBA{R: 153, G: 196, B: 255, A: 110}
	poly.LineStyle.Color = color.RGBA{R: 35, G: 78, B: 140, A: 220}
	poly.LineStyle.Width = vg.Points(1.2)
	p.Add(poly)
	p.Legend.Add("допустимая область", poly)

	for i := range prob.A {
		linePts, ok := constraintSegment(prob.A[i][0], prob.A[i][1], prob.B[i], xMin, xMax, yMin, yMax)
		if !ok {
			continue
		}
		ln, err := plotter.NewLine(linePts)
		if err != nil {
			return fmt.Errorf("NewLine(constraint %d): %w", i+1, err)
		}
		ln.LineStyle.Width = vg.Points(1.6)
		ln.LineStyle.Color = lineColors[i%len(lineColors)]
		ln.LineStyle.Dashes = dashStyles[(i+1)%len(dashStyles)]
		p.Add(ln)
		p.Legend.Add(fmt.Sprintf("огр. %d", i+1), ln)
	}

	if res != nil && len(res.X) >= 2 {
		xStar := res.X[0]
		yStar := res.X[1]
		optPts := plotter.XYs{{X: xStar, Y: yStar}}
		sc, err := plotter.NewScatter(optPts)
		if err != nil {
			return fmt.Errorf("NewScatter(opt): %w", err)
		}
		sc.Shape = draw.CrossGlyph{}
		sc.Radius = vg.Points(5)
		sc.Color = color.RGBA{R: 200, G: 40, B: 40, A: 255}
		p.Add(sc)
		p.Legend.Add("оптимум", sc)

		objPts, ok := objectiveSegment(prob.C[0], prob.C[1], res.Objective, xMin, xMax, yMin, yMax)
		if ok {
			objLine, err := plotter.NewLine(objPts)
			if err != nil {
				return fmt.Errorf("NewLine(objective): %w", err)
			}
			objLine.LineStyle.Width = vg.Points(2)
			objLine.LineStyle.Color = color.RGBA{R: 200, G: 40, B: 40, A: 255}
			objLine.LineStyle.Dashes = []vg.Length{vg.Points(7), vg.Points(4)}
			p.Add(objLine)
			p.Legend.Add("F = const (в оптимуме)", objLine)
		}
	}

	if err := p.Save(11*vg.Inch, 8*vg.Inch, filename); err != nil {
		return fmt.Errorf("Save(%q): %w", filename, err)
	}
	return nil
}

func constraintSegment(a, b, c, xMin, xMax, yMin, yMax float64) (plotter.XYs, bool) {
	const eps = 1e-10
	pts := make([]plotter.XY, 0, 4)
	add := func(x, y float64) {
		if x < xMin-1e-8 || x > xMax+1e-8 || y < yMin-1e-8 || y > yMax+1e-8 {
			return
		}
		for _, p := range pts {
			if math.Abs(p.X-x) <= 1e-6 && math.Abs(p.Y-y) <= 1e-6 {
				return
			}
		}
		pts = append(pts, plotter.XY{X: x, Y: y})
	}

	if math.Abs(b) > eps {
		add(xMin, (c-a*xMin)/b)
		add(xMax, (c-a*xMax)/b)
	}
	if math.Abs(a) > eps {
		add((c-b*yMin)/a, yMin)
		add((c-b*yMax)/a, yMax)
	}

	if len(pts) < 2 {
		return nil, false
	}
	if len(pts) > 2 {
		maxD := -1.0
		bestI, bestJ := 0, 1
		for i := 0; i < len(pts); i++ {
			for j := i + 1; j < len(pts); j++ {
				dx := pts[i].X - pts[j].X
				dy := pts[i].Y - pts[j].Y
				d := dx*dx + dy*dy
				if d > maxD {
					maxD = d
					bestI, bestJ = i, j
				}
			}
		}
		return plotter.XYs{pts[bestI], pts[bestJ]}, true
	}
	return plotter.XYs{pts[0], pts[1]}, true
}

func objectiveSegment(c1, c2, val, xMin, xMax, yMin, yMax float64) (plotter.XYs, bool) {
	return constraintSegment(c1, c2, val, xMin, xMax, yMin, yMax)
}

func convexHull(points []plotter.XY) plotter.XYs {
	if len(points) <= 1 {
		return append(plotter.XYs(nil), points...)
	}

	pts := append([]plotter.XY(nil), points...)
	sort.Slice(pts, func(i, j int) bool {
		if pts[i].X == pts[j].X {
			return pts[i].Y < pts[j].Y
		}
		return pts[i].X < pts[j].X
	})

	cross := func(o, a, b plotter.XY) float64 {
		return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
	}

	lower := make([]plotter.XY, 0, len(pts))
	for _, p := range pts {
		for len(lower) >= 2 && cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}

	upper := make([]plotter.XY, 0, len(pts))
	for i := len(pts) - 1; i >= 0; i-- {
		p := pts[i]
		for len(upper) >= 2 && cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}

	hull := append(lower[:len(lower)-1], upper[:len(upper)-1]...)
	return plotter.XYs(hull)
}
