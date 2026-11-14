package randomvars

import (
	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// SampleStats — выборочные точечные оценки.
type SampleStats struct {
	Mean     float64 `json:"mean"`
	Variance float64 `json:"variance"`
	StdDev   float64 `json:"stdDev"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
}

// summarize считает несмещённую выборочную дисперсию (делитель n−1).
func summarize(samples []float64) SampleStats {
	if len(samples) == 0 {
		return SampleStats{}
	}

	mean, variance := stat.MeanVariance(samples, nil)
	if len(samples) < 2 {
		variance = 0
	}

	return SampleStats{
		Mean:     mean,
		Variance: variance,
		StdDev:   math.Sqrt(variance),
		Min:      floats.Min(samples),
		Max:      floats.Max(samples),
	}
}

// confidenceIntervals строит доверительные интервалы для математического
// ожидания и дисперсии. Для среднего используется распределение Стьюдента,
// для дисперсии — распределение χ².
func confidenceIntervals(stats SampleStats, n int, alpha float64) (Interval, Interval) {
	confidence := 1 - alpha
	if n < 2 {
		return Interval{Confidence: confidence}, Interval{Confidence: confidence}
	}

	df := float64(n - 1)
	t := distuv.StudentsT{Mu: 0, Sigma: 1, Nu: df}.Quantile(1 - alpha/2)
	halfWidth := t * stats.StdDev / math.Sqrt(float64(n))
	meanCI := Interval{
		Lower:      stats.Mean - halfWidth,
		Upper:      stats.Mean + halfWidth,
		Confidence: confidence,
	}

	chi := distuv.ChiSquared{K: df}
	chiLow := chi.Quantile(alpha / 2)
	chiHigh := chi.Quantile(1 - alpha/2)
	scaled := df * stats.Variance
	varianceCI := Interval{
		Lower:      scaled / chiHigh,
		Upper:      scaled / chiLow,
		Confidence: confidence,
	}

	return meanCI, varianceCI
}
