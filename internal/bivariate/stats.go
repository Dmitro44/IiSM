package bivariate

import (
	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

func summarize(xs, ys []float64) SampleStats {
	stats := SampleStats{
		MeanX: stat.Mean(xs, nil),
		MeanY: stat.Mean(ys, nil),
		MinX:  floats.Min(xs),
		MaxX:  floats.Max(xs),
		MinY:  floats.Min(ys),
		MaxY:  floats.Max(ys),
	}
	stats.VarianceX = stat.Variance(xs, nil)
	stats.VarianceY = stat.Variance(ys, nil)
	stats.StdDevX = math.Sqrt(stats.VarianceX)
	stats.StdDevY = math.Sqrt(stats.VarianceY)
	stats.Covariance = stat.Covariance(xs, ys, nil)
	if stats.StdDevX > 0 && stats.StdDevY > 0 {
		stats.Correlation = stats.Covariance / (stats.StdDevX * stats.StdDevY)
	}
	return stats
}

func confidenceIntervals(stats SampleStats, n int, alpha float64) Intervals {
	return Intervals{
		MeanX:       meanInterval(stats.MeanX, stats.StdDevX, n, alpha),
		MeanY:       meanInterval(stats.MeanY, stats.StdDevY, n, alpha),
		VarianceX:   varianceInterval(stats.VarianceX, n, alpha),
		VarianceY:   varianceInterval(stats.VarianceY, n, alpha),
		Correlation: correlationInterval(stats.Correlation, n, alpha),
	}
}

func meanInterval(mean, stdDev float64, n int, alpha float64) Interval {
	confidence := 1 - alpha
	if n < 2 || stdDev <= 0 {
		return Interval{Confidence: confidence}
	}
	t := distuv.StudentsT{Mu: 0, Sigma: 1, Nu: float64(n - 1)}.Quantile(1 - alpha/2)
	halfWidth := t * stdDev / math.Sqrt(float64(n))
	return Interval{Lower: mean - halfWidth, Upper: mean + halfWidth, Confidence: confidence}
}

func varianceInterval(variance float64, n int, alpha float64) Interval {
	confidence := 1 - alpha
	if n < 2 {
		return Interval{Confidence: confidence}
	}
	df := float64(n - 1)
	chi := distuv.ChiSquared{K: df}
	scaled := df * variance
	return Interval{
		Lower:      scaled / chi.Quantile(1-alpha/2),
		Upper:      scaled / chi.Quantile(alpha/2),
		Confidence: confidence,
	}
}

func correlationInterval(r float64, n int, alpha float64) Interval {
	confidence := 1 - alpha
	if n < 4 || math.Abs(r) >= 1 {
		return Interval{Confidence: confidence}
	}
	z := math.Atanh(r)
	margin := distuv.UnitNormal.Quantile(1-alpha/2) / math.Sqrt(float64(n-3))
	return Interval{
		Lower:      math.Tanh(z - margin),
		Upper:      math.Tanh(z + margin),
		Confidence: confidence,
	}
}

// goodnessTest — критерий Пирсона с последовательным объединением редких
// интервалов; параметры распределения заданы, поэтому df = групп - 1.
func goodnessTest(name string, observed []int, expected []float64, alpha float64) Test {
	test := Test{Name: name, Alpha: alpha}

	type group struct {
		observed int
		expected float64
	}
	groups := make([]group, 0, len(observed))
	pending := group{}
	for i := range observed {
		pending.observed += observed[i]
		pending.expected += expected[i]
		if pending.expected >= 5 {
			groups = append(groups, pending)
			pending = group{}
		}
	}
	if pending.expected > 0 {
		if len(groups) == 0 {
			groups = append(groups, pending)
		} else {
			groups[len(groups)-1].observed += pending.observed
			groups[len(groups)-1].expected += pending.expected
		}
	}

	if len(groups) < 2 {
		test.Note = "менее двух групп с ожидаемой частотой не меньше 5"
		return test
	}

	for _, group := range groups {
		diff := float64(group.observed) - group.expected
		test.Statistic += diff * diff / group.expected
	}
	test.DegreesOfFreedom = len(groups) - 1
	test.PValue = distuv.ChiSquared{K: float64(test.DegreesOfFreedom)}.Survival(test.Statistic)
	test.Valid = true
	test.Accepted = test.PValue >= alpha
	test.Note = "группы с E < 5 объединяются при необходимости; df = групп - 1"
	return test
}

// contingencyTest — критерий независимости по таблице сопряжённости.
// Ожидаемые частоты берутся из выборочных маргинальных сумм.
func contingencyTest(name string, observed [][]int, alpha float64) Test {
	test := Test{Name: name, Alpha: alpha}
	rows := len(observed)
	cols := len(observed[0])

	rowSums := make([]int, rows)
	colSums := make([]int, cols)
	total := 0
	for i := range rows {
		for j := range cols {
			rowSums[i] += observed[i][j]
			colSums[j] += observed[i][j]
			total += observed[i][j]
		}
	}
	if total == 0 {
		test.Note = "таблица не содержит наблюдений"
		return test
	}

	minExpected := math.Inf(1)
	for i := range rows {
		for j := range cols {
			expected := float64(rowSums[i]) * float64(colSums[j]) / float64(total)
			minExpected = min(minExpected, expected)
			if expected > 0 {
				diff := float64(observed[i][j]) - expected
				test.Statistic += diff * diff / expected
			}
		}
	}

	test.DegreesOfFreedom = (rows - 1) * (cols - 1)
	test.PValue = distuv.ChiSquared{K: float64(test.DegreesOfFreedom)}.Survival(test.Statistic)
	test.Valid = minExpected >= 5
	if test.Valid {
		test.Note = "таблица сопряжённости; E_ij = n_i*n_j / N"
	} else {
		test.Note = "ожидаемая частота в некоторых ячейках меньше 5"
	}
	test.Accepted = test.Valid && test.PValue >= alpha
	return test
}

func meanTest(name string, mean, stdDev float64, n int, mu0, alpha float64) Test {
	test := Test{Name: name, Alpha: alpha}
	if n < 2 || stdDev <= 0 {
		test.Note = "недостаточно данных или нулевая выборочная дисперсия"
		return test
	}
	test.Statistic = (mean - mu0) / (stdDev / math.Sqrt(float64(n)))
	test.DegreesOfFreedom = n - 1
	cdf := distuv.StudentsT{Mu: 0, Sigma: 1, Nu: float64(n - 1)}.CDF(test.Statistic)
	test.PValue = min(1, 2*min(cdf, 1-cdf))
	test.Valid = true
	test.Accepted = test.PValue >= alpha
	test.Note = "двусторонний t-критерий Стьюдента"
	return test
}

func varianceTest(name string, variance float64, n int, sigma0Squared, alpha float64) Test {
	test := Test{Name: name, Alpha: alpha}
	if n < 2 || sigma0Squared <= 0 {
		test.Note = "недостаточно данных для проверки дисперсии"
		return test
	}
	test.Statistic = float64(n-1) * variance / sigma0Squared
	test.DegreesOfFreedom = n - 1
	cdf := distuv.ChiSquared{K: float64(n - 1)}.CDF(test.Statistic)
	test.PValue = min(1, 2*min(cdf, 1-cdf))
	test.Valid = true
	test.Accepted = test.PValue >= alpha
	test.Note = "χ²-критерий точен для нормального закона; при большом N — асимптотическое приближение"
	return test
}

// correlationTest — критерий Фишера для коэффициента корреляции.
func correlationTest(name string, r float64, n int, rho0, alpha float64) Test {
	test := Test{Name: name, Alpha: alpha}
	if n < 4 || math.Abs(r) >= 1 {
		test.Note = "нужно n >= 4 и |r| < 1"
		return test
	}
	z := math.Atanh(r)
	zeta := math.Atanh(rho0) + rho0/(2*float64(n-1))
	test.Statistic = (z - zeta) * math.Sqrt(float64(n-3))
	cdf := distuv.UnitNormal.CDF(test.Statistic)
	test.PValue = min(1, 2*min(cdf, 1-cdf))
	test.Valid = true
	test.Accepted = test.PValue >= alpha
	test.Note = "z-критерий Фишера (нормальное приближение); df не используется"
	return test
}
