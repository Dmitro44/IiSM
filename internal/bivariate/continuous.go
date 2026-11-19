package bivariate

import (
	"math"
	"math/rand/v2"
)

const (
	jointQuadrature = 24
	marginalSteps   = 128
	conditionalXs   = 5  // x0 = -2, -1, 0, 1, 2
	conditionalN    = 40 // 40 отрезков, 41 точка
	formulaText     = "f(x, y) = (9 - √(x² + y²)) / (63π) при x² + y² ≤ 9, иначе 0"
	methodText      = "Метод отбрасывания: точка равномерно в круге радиуса 3, принимается с вероятностью (9-r)/9"
)

// jointDensity — плотность f(x, y) из варианта 3.
func jointDensity(x, y float64) float64 {
	r := math.Hypot(x, y)
	if r > diskRadius {
		return 0
	}
	return (9 - r) / (63 * math.Pi)
}

// marginalDensity — аналитическая маргинальная плотность f_X(x).
// Формула симметрична по x, поэтому f_Y получается той же функцией.
func marginalDensity(x float64) float64 {
	a := math.Sqrt(9 - x*x)
	if a <= 0 || math.IsNaN(a) {
		return 0
	}
	if a >= diskRadius {
		// При x = 0 множитель x^2/2 умножается на бесконечность, поэтому
		// берём предел аналитически.
		return 5 / (7 * math.Pi)
	}
	// ln((3+a)/(3-a)) через log1p устойчивее при a -> 0.
	logRatio := math.Log1p(2 * a / (diskRadius - a))
	return (15*a - x*x/2*logRatio) / (63 * math.Pi)
}

// conditionalDensity — плотность Y при условии X = x0.
func conditionalDensity(x0, y float64) float64 {
	if x0 == 0 {
		if math.Abs(y) > diskRadius {
			return 0
		}
		return (9 - math.Abs(y)) / 45
	}
	densityX := marginalDensity(x0)
	if densityX <= 0 {
		return 0
	}
	return jointDensity(x0, y) / densityX
}

// sampleDisk разыгрывает точку с плотностью (9-r)/(63*pi). В полярных
// координатах плотность точки равна f(r cos t, r sin t)*r = r(9-r)/(63*pi),
// то есть радиальная плотность 2r(9-r)/63 и равномерный угол. Равномерная
// точка диска имеет радиальную плотность 2r/9, а отбор с вероятностью (9-r)/9
// и нормировкой на вероятность приёма 7/9 даёт ровно 2r(9-r)/63.
func sampleDisk(rng *rand.Rand) Point {
	for {
		r := diskRadius * math.Sqrt(rng.Float64())
		theta := 2 * math.Pi * rng.Float64()
		if rng.Float64() <= (9-r)/9 {
			return Point{X: r * math.Cos(theta), Y: r * math.Sin(theta)}
		}
	}
}

// simpson интегрирует f на [lo, hi] методом Симпсона.
func simpson(f func(float64) float64, lo, hi float64, steps int) float64 {
	if steps%2 == 1 {
		steps++
	}
	h := (hi - lo) / float64(steps)
	sum := f(lo) + f(hi)
	for i := 1; i < steps; i++ {
		weight := 2.0
		if i%2 == 1 {
			weight = 4
		}
		sum += weight * f(lo+float64(i)*h)
	}
	return sum * h / 3
}

func simulateContinuous(req ContinuousRequest, rng *rand.Rand) (ContinuousResult, error) {
	if err := validateSample(req.N); err != nil {
		return ContinuousResult{}, err
	}
	if err := validateBins(req.Bins); err != nil {
		return ContinuousResult{}, err
	}
	if err := validateAlpha(req.Alpha); err != nil {
		return ContinuousResult{}, err
	}

	xs := make([]float64, req.N)
	ys := make([]float64, req.N)
	for i := range req.N {
		point := sampleDisk(rng)
		xs[i] = point.X
		ys[i] = point.Y
	}

	stats := summarize(xs, ys)
	intervals := confidenceIntervals(stats, req.N, req.Alpha)
	histogramX := marginalHistogram(xs, req.Bins)
	histogramY := marginalHistogram(ys, req.Bins)
	joint := jointCells(xs, ys)
	conditional := conditionalCurves()

	result := ContinuousResult{
		SampleSize:    req.N,
		Alpha:         req.Alpha,
		Formula:       formulaText,
		Method:        methodText,
		Preview:       previewPoints(xs, ys),
		Stats:         stats,
		Theoretical:   continuousTheory(),
		Intervals:     intervals,
		HistogramX:    histogramX,
		HistogramY:    histogramY,
		Conditional:   conditional,
		Joint:         joint,
		JointGridSize: jointGridSize,
		Tests: []Test{
			goodnessTest("Согласие X с теоретической плотностью", counts(histogramX), expectedCounts(histogramX), req.Alpha),
			goodnessTest("Согласие Y с теоретической плотностью", counts(histogramY), expectedCounts(histogramY), req.Alpha),
			contingencyTest("Независимость X и Y", jointMatrix(joint), req.Alpha),
			meanTest("Математическое ожидание X", stats.MeanX, stats.StdDevX, req.N, 0, req.Alpha),
			meanTest("Математическое ожидание Y", stats.MeanY, stats.StdDevY, req.N, 0, req.Alpha),
			varianceTest("Дисперсия X", stats.VarianceX, req.N, 297.0/140, req.Alpha),
			varianceTest("Дисперсия Y", stats.VarianceY, req.N, 297.0/140, req.Alpha),
			correlationTest("Корреляция", stats.Correlation, req.N, 0, req.Alpha),
		},
	}

	if !finiteContinuous(result) {
		return ContinuousResult{}, ErrNumeric
	}
	return result, nil
}

func continuousTheory() TheoreticalStats {
	// E[X] = E[Y] = 0 и Cov = 0 из-за симметрии круга;
	// D[X] = D[Y] = E[R^2]/2 = (297/70)/2 = 297/140.
	return TheoreticalStats{
		MeanX:       0,
		MeanY:       0,
		VarianceX:   297.0 / 140,
		VarianceY:   297.0 / 140,
		Covariance:  0,
		Correlation: 0,
	}
}

func previewPoints(xs, ys []float64) []Point {
	n := min(len(xs), previewLimit)
	points := make([]Point, n)
	for i := range n {
		points[i] = Point{X: xs[i], Y: ys[i]}
	}
	return points
}

func marginalHistogram(values []float64, bins int) []HistogramBin {
	width := 2 * diskRadius / float64(bins)
	histogram := make([]HistogramBin, bins)
	for i := range histogram {
		lower := -diskRadius + float64(i)*width
		upper := lower + width
		center := lower + width/2
		probability := simpson(marginalDensity, lower, upper, marginalSteps)
		histogram[i] = HistogramBin{
			Lower:              lower,
			Upper:              upper,
			TheoreticalDensity: marginalDensity(center),
			ExpectedCount:      probability * float64(len(values)),
		}
	}

	histogram[0].Lower = -diskRadius
	histogram[bins-1].Upper = diskRadius

	total := float64(len(values))
	for _, value := range values {
		index := int((value + diskRadius) / width)
		if index < 0 {
			index = 0
		}
		if index >= bins {
			index = bins - 1
		}
		histogram[index].Count++
	}
	for i := range histogram {
		histogram[i].Frequency = float64(histogram[i].Count) / total
		histogram[i].Density = float64(histogram[i].Count) / (total * width)
	}
	return histogram
}

func counts(histogram []HistogramBin) []int {
	observed := make([]int, len(histogram))
	for i, bin := range histogram {
		observed[i] = bin.Count
	}
	return observed
}

func expectedCounts(histogram []HistogramBin) []float64 {
	expected := make([]float64, len(histogram))
	for i, bin := range histogram {
		expected[i] = bin.ExpectedCount
	}
	return expected
}

func conditionalCurves() []ConditionalCurve {
	curves := make([]ConditionalCurve, 0, conditionalXs)
	for i := range conditionalXs {
		x0 := float64(i) - 2
		a := math.Sqrt(9 - x0*x0)
		points := make([]CurvePoint, conditionalN+1)
		for j := range conditionalN + 1 {
			y := -a + 2*a*float64(j)/conditionalN
			points[j] = CurvePoint{Value: y, Density: conditionalDensity(x0, y)}
		}
		curves = append(curves, ConditionalCurve{
			FixedX:          x0,
			MarginalDensity: marginalDensity(x0),
			Points:          points,
		})
	}
	return curves
}

func jointCells(xs, ys []float64) []JointCell {
	width := 2 * diskRadius / float64(jointGridSize)
	cells := make([]JointCell, jointGridSize*jointGridSize)
	for i := range jointGridSize {
		for j := range jointGridSize {
			lowerX := -diskRadius + float64(i)*width
			lowerY := -diskRadius + float64(j)*width
			cells[i*jointGridSize+j] = JointCell{
				CenterX:       lowerX + width/2,
				CenterY:       lowerY + width/2,
				ExpectedCount: cellProbability(lowerX, lowerY, width) * float64(len(xs)),
			}
		}
	}

	for k := range xs {
		i := int((xs[k] + diskRadius) / width)
		j := int((ys[k] + diskRadius) / width)
		if i < 0 {
			i = 0
		}
		if i >= jointGridSize {
			i = jointGridSize - 1
		}
		if j < 0 {
			j = 0
		}
		if j >= jointGridSize {
			j = jointGridSize - 1
		}
		cells[i*jointGridSize+j].Count++
	}
	return cells
}

// cellProbability оценивает вероятность попадания в ячейку средней точкой
// (квадратура 24x24 на ячейку).
func cellProbability(lowerX, lowerY, width float64) float64 {
	step := width / jointQuadrature
	sum := 0.0
	for i := range jointQuadrature {
		x := lowerX + (float64(i)+0.5)*step
		for j := range jointQuadrature {
			y := lowerY + (float64(j)+0.5)*step
			sum += jointDensity(x, y)
		}
	}
	return sum * step * step
}

func jointMatrix(cells []JointCell) [][]int {
	matrix := make([][]int, jointGridSize)
	for i := range jointGridSize {
		matrix[i] = make([]int, jointGridSize)
		for j := range jointGridSize {
			matrix[i][j] = cells[i*jointGridSize+j].Count
		}
	}
	return matrix
}
