package bivariate

import (
	"math"
	"math/rand/v2"
)

var (
	discreteValuesX = []float64{1, 2, 3}
	discreteValuesY = []float64{0, 1}

	// Матрица вероятностей: строка — значение X, столбец — значение Y.
	discreteProbabilities = [][]float64{
		{0.10, 0.20},
		{0.20, 0.10},
		{0.25, 0.15},
	}
)

func simulateDiscrete(req DiscreteRequest, rng *rand.Rand) (DiscreteResult, error) {
	if err := validateSample(req.N); err != nil {
		return DiscreteResult{}, err
	}
	if err := validateAlpha(req.Alpha); err != nil {
		return DiscreteResult{}, err
	}

	rows := len(discreteValuesX)
	cols := len(discreteValuesY)
	counts := make([][]int, rows)
	for i := range counts {
		counts[i] = make([]int, cols)
	}

	// Разворачиваем матрицу построчно и накапливаем вероятности.
	cumulative := make([]float64, 0, rows*cols)
	total := 0.0
	for i := range rows {
		for j := range cols {
			total += discreteProbabilities[i][j]
			cumulative = append(cumulative, total)
		}
	}
	cumulative[len(cumulative)-1] = 1

	xs := make([]float64, req.N)
	ys := make([]float64, req.N)
	for k := range req.N {
		u := rng.Float64()
		index := 0
		for index < len(cumulative)-1 && cumulative[index] <= u {
			index++
		}
		i := index / cols
		j := index % cols
		counts[i][j]++
		xs[k] = discreteValuesX[i]
		ys[k] = discreteValuesY[j]
	}

	theory, marginalsX, marginalsY := discreteTheory()
	stats := summarize(xs, ys)
	intervals := confidenceIntervals(stats, req.N, req.Alpha)

	marginalCountsX := make([]int, rows)
	for i := range rows {
		for j := range cols {
			marginalCountsX[i] += counts[i][j]
		}
	}
	marginalCountsY := make([]int, cols)
	for j := range cols {
		for i := range rows {
			marginalCountsY[j] += counts[i][j]
		}
	}

	histogramX := make([]DiscreteCategory, rows)
	expectedX := make([]float64, rows)
	for i := range rows {
		histogramX[i] = DiscreteCategory{
			Value:         discreteValuesX[i],
			Count:         marginalCountsX[i],
			Frequency:     float64(marginalCountsX[i]) / float64(req.N),
			Probability:   marginalsX[i],
			ExpectedCount: marginalsX[i] * float64(req.N),
		}
		expectedX[i] = histogramX[i].ExpectedCount
	}
	histogramY := make([]DiscreteCategory, cols)
	expectedY := make([]float64, cols)
	for j := range cols {
		histogramY[j] = DiscreteCategory{
			Value:         discreteValuesY[j],
			Count:         marginalCountsY[j],
			Frequency:     float64(marginalCountsY[j]) / float64(req.N),
			Probability:   marginalsY[j],
			ExpectedCount: marginalsY[j] * float64(req.N),
		}
		expectedY[j] = histogramY[j].ExpectedCount
	}

	expected := make([][]float64, rows)
	for i := range rows {
		expected[i] = make([]float64, cols)
		for j := range cols {
			expected[i][j] = discreteProbabilities[i][j] * float64(req.N)
		}
	}

	conditionalXGivenY := make([]ConditionalRow, cols)
	for j := range cols {
		row := ConditionalRow{
			Fixed:       discreteValuesY[j],
			Values:      append([]float64(nil), discreteValuesX...),
			Theoretical: make([]float64, rows),
			Observed:    make([]float64, rows),
		}
		for i := range rows {
			row.Theoretical[i] = discreteProbabilities[i][j] / marginalsY[j]
			if marginalCountsY[j] > 0 {
				row.Observed[i] = float64(counts[i][j]) / float64(marginalCountsY[j])
			}
		}
		conditionalXGivenY[j] = row
	}

	conditionalYGivenX := make([]ConditionalRow, rows)
	for i := range rows {
		row := ConditionalRow{
			Fixed:       discreteValuesX[i],
			Values:      append([]float64(nil), discreteValuesY...),
			Theoretical: make([]float64, cols),
			Observed:    make([]float64, cols),
		}
		for j := range cols {
			row.Theoretical[j] = discreteProbabilities[i][j] / marginalsX[i]
			if marginalCountsX[i] > 0 {
				row.Observed[j] = float64(counts[i][j]) / float64(marginalCountsX[i])
			}
		}
		conditionalYGivenX[i] = row
	}

	result := DiscreteResult{
		SampleSize:         req.N,
		Alpha:              req.Alpha,
		ValuesX:            append([]float64(nil), discreteValuesX...),
		ValuesY:            append([]float64(nil), discreteValuesY...),
		Probabilities:      copyMatrix(discreteProbabilities),
		MarginalsX:         marginalsX,
		MarginalsY:         marginalsY,
		Preview:            previewPoints(xs, ys),
		Stats:              stats,
		Theoretical:        theory,
		Intervals:          intervals,
		HistogramX:         histogramX,
		HistogramY:         histogramY,
		Observed:           counts,
		Expected:           expected,
		ConditionalXGivenY: conditionalXGivenY,
		ConditionalYGivenX: conditionalYGivenX,
		Tests: []Test{
			goodnessTest("Согласие X с теоретическим распределением", marginalCountsX, expectedX, req.Alpha),
			goodnessTest("Согласие Y с теоретическим распределением", marginalCountsY, expectedY, req.Alpha),
			contingencyTest("Независимость X и Y", counts, req.Alpha),
			meanTest("Математическое ожидание X", stats.MeanX, stats.StdDevX, req.N, theory.MeanX, req.Alpha),
			meanTest("Математическое ожидание Y", stats.MeanY, stats.StdDevY, req.N, theory.MeanY, req.Alpha),
			varianceTest("Дисперсия X", stats.VarianceX, req.N, theory.VarianceX, req.Alpha),
			varianceTest("Дисперсия Y", stats.VarianceY, req.N, theory.VarianceY, req.Alpha),
			correlationTest("Корреляция", stats.Correlation, req.N, theory.Correlation, req.Alpha),
		},
	}

	if !finiteDiscrete(result) {
		return DiscreteResult{}, ErrNumeric
	}
	return result, nil
}

func discreteTheory() (TheoreticalStats, []float64, []float64) {
	rows := len(discreteValuesX)
	cols := len(discreteValuesY)

	marginalsX := make([]float64, rows)
	marginalsY := make([]float64, cols)
	for i := range rows {
		for j := range cols {
			marginalsX[i] += discreteProbabilities[i][j]
			marginalsY[j] += discreteProbabilities[i][j]
		}
	}

	var theory TheoreticalStats
	for i := range rows {
		theory.MeanX += discreteValuesX[i] * marginalsX[i]
	}
	for j := range cols {
		theory.MeanY += discreteValuesY[j] * marginalsY[j]
	}
	for i := range rows {
		for j := range cols {
			dx := discreteValuesX[i] - theory.MeanX
			dy := discreteValuesY[j] - theory.MeanY
			p := discreteProbabilities[i][j]
			theory.VarianceX += p * dx * dx
			theory.VarianceY += p * dy * dy
			theory.Covariance += p * dx * dy
		}
	}
	if theory.VarianceX > 0 && theory.VarianceY > 0 {
		theory.Correlation = theory.Covariance / math.Sqrt(theory.VarianceX*theory.VarianceY)
	}
	return theory, marginalsX, marginalsY
}

func copyMatrix(matrix [][]float64) [][]float64 {
	copied := make([][]float64, len(matrix))
	for i, row := range matrix {
		copied[i] = append([]float64(nil), row...)
	}
	return copied
}
