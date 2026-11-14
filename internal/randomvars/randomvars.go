package randomvars

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"

	"gonum.org/v1/gonum/stat/distuv"
)

const (
	MinSamples    = 10
	MaxSamples    = 1_000_000
	MinBins       = 2
	MaxBins       = 50
	MaxCategories = 100
	MinAlpha      = 1e-6
)

var (
	ErrSampleSize    = errors.New("N должно быть от 10 до 1000000")
	ErrBins          = errors.New("число интервалов должно быть от 2 до 50")
	ErrAlpha         = errors.New("уровень значимости alpha должен быть в интервале [1e-6, 0.5)")
	ErrLaw           = errors.New("неизвестный закон распределения")
	ErrParameter     = errors.New("некорректный параметр закона")
	ErrValues        = errors.New("нужно не меньше двух различных значений y_i")
	ErrProbabilities = errors.New("вероятности должны быть положительными и соответствовать значениям")
	ErrSum           = errors.New("сумма вероятностей должна быть равна 1")
	ErrNumeric       = errors.New("параметры слишком велики: результат не помещается в float64")
)

// ContinuousRequest — параметры имитации непрерывной случайной величины.
type ContinuousRequest struct {
	Law    string  `json:"law"`
	A      float64 `json:"a"`
	B      float64 `json:"b"`
	Lambda float64 `json:"lambda"`
	N      int     `json:"n"`
	Bins   int     `json:"bins"`
	Alpha  float64 `json:"alpha"`
}

// DiscreteRequest — значения и вероятности дискретной случайной величины.
type DiscreteRequest struct {
	Values        []float64 `json:"values"`
	Probabilities []float64 `json:"probabilities"`
	N             int       `json:"n"`
	Alpha         float64   `json:"alpha"`
}

type Interval struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Confidence float64 `json:"confidence"`
}

// ContinuousBin — интервал гистограммы непрерывной величины. Границы могут
// быть бесконечными: тогда крайний интервал включает соответствующий хвост.
type ContinuousBin struct {
	Lower          float64 `json:"lower"`
	Upper          float64 `json:"upper"`
	LowerUnbounded bool    `json:"lowerUnbounded"`
	UpperUnbounded bool    `json:"upperUnbounded"`
	Count          int     `json:"count"`
	Frequency      float64 `json:"frequency"`
	Probability    float64 `json:"probability"`
	ExpectedCount  float64 `json:"expectedCount"`
}

// DiscreteCategory — одна пара y_i, p_i и её выборочная частота.
type DiscreteCategory struct {
	Value         float64 `json:"value"`
	Count         int     `json:"count"`
	Frequency     float64 `json:"frequency"`
	Probability   float64 `json:"probability"`
	ExpectedCount float64 `json:"expectedCount"`
}

type GoodnessOfFit struct {
	Name             string  `json:"name"`
	Statistic        float64 `json:"statistic"`
	DegreesOfFreedom int     `json:"degreesOfFreedom"`
	PValue           float64 `json:"pValue"`
	Alpha            float64 `json:"alpha"`
	Accepted         bool    `json:"accepted"`
	Valid            bool    `json:"valid"`
	Groups           int     `json:"groups"`
}

type ContinuousResult struct {
	Law                 string          `json:"law"`
	Name                string          `json:"name"`
	Formula             string          `json:"formula"`
	SampleSize          int             `json:"sampleSize"`
	Bins                int             `json:"bins"`
	Preview             []float64       `json:"preview"`
	Stats               SampleStats     `json:"stats"`
	TheoreticalMean     float64         `json:"theoreticalMean"`
	TheoreticalVariance float64         `json:"theoreticalVariance"`
	MeanCI              Interval        `json:"meanCi"`
	VarianceCI          Interval        `json:"varianceCi"`
	Histogram           []ContinuousBin `json:"histogram"`
	Test                GoodnessOfFit   `json:"test"`
}

type DiscreteResult struct {
	SampleSize          int                `json:"sampleSize"`
	Preview             []float64          `json:"preview"`
	Stats               SampleStats        `json:"stats"`
	TheoreticalMean     float64            `json:"theoreticalMean"`
	TheoreticalVariance float64            `json:"theoreticalVariance"`
	MeanCI              Interval           `json:"meanCi"`
	VarianceCI          Interval           `json:"varianceCi"`
	Categories          []DiscreteCategory `json:"categories"`
	Test                GoodnessOfFit      `json:"test"`
}

type RandomVariableService struct{}

func NewRandomVariableService() *RandomVariableService {
	return &RandomVariableService{}
}

func (s *RandomVariableService) SimulateContinuous(req ContinuousRequest) (ContinuousResult, error) {
	if err := validateSample(req.N); err != nil {
		return ContinuousResult{}, err
	}
	if err := validateAlpha(req.Alpha); err != nil {
		return ContinuousResult{}, err
	}
	if req.Bins < MinBins || req.Bins > MaxBins {
		return ContinuousResult{}, fmt.Errorf("%w: получено %d", ErrBins, req.Bins)
	}

	d, err := newDistribution(req.Law, req.A, req.B, req.Lambda)
	if err != nil {
		return ContinuousResult{}, err
	}

	samples := make([]float64, req.N)
	for i := range samples {
		samples[i] = d.inverse(rand.Float64())
	}

	histogram := continuousHistogram(d, samples, req.Bins)
	observed := make([]int, len(histogram))
	expected := make([]float64, len(histogram))
	for i, bin := range histogram {
		observed[i] = bin.Count
		expected[i] = bin.ExpectedCount
	}

	stats := summarize(samples)
	meanCI, varianceCI := confidenceIntervals(stats, req.N, req.Alpha)
	if !finiteStats(stats) || !finiteInterval(meanCI) || !finiteInterval(varianceCI) {
		return ContinuousResult{}, ErrNumeric
	}
	for _, bin := range histogram {
		if !finite(bin.Probability) || !finite(bin.ExpectedCount) {
			return ContinuousResult{}, ErrNumeric
		}
	}

	return ContinuousResult{
		Law:                 string(d.law),
		Name:                d.name(),
		Formula:             d.formula(),
		SampleSize:          req.N,
		Bins:                req.Bins,
		Preview:             preview(samples),
		Stats:               stats,
		TheoreticalMean:     d.mean(),
		TheoreticalVariance: d.variance(),
		MeanCI:              meanCI,
		VarianceCI:          varianceCI,
		Histogram:           histogram,
		Test:                chiSquareGoodnessOfFit(observed, expected, req.Alpha),
	}, nil
}

func (s *RandomVariableService) SimulateDiscrete(req DiscreteRequest) (DiscreteResult, error) {
	if err := validateSample(req.N); err != nil {
		return DiscreteResult{}, err
	}
	if err := validateAlpha(req.Alpha); err != nil {
		return DiscreteResult{}, err
	}

	values, probabilities, err := validateDiscrete(req.Values, req.Probabilities)
	if err != nil {
		return DiscreteResult{}, err
	}

	cumulative := make([]float64, len(probabilities))
	total := 0.0
	for i, p := range probabilities {
		total += p
		cumulative[i] = total
	}
	cumulative[len(cumulative)-1] = 1

	counts := make([]int, len(values))
	samples := make([]float64, req.N)
	for i := range samples {
		index := pickCategory(rand.Float64(), cumulative)
		counts[index]++
		samples[i] = values[index]
	}

	categories := make([]DiscreteCategory, len(values))
	expectedCounts := make([]float64, len(values))
	theoreticalMean := 0.0
	theoreticalVariance := 0.0
	for i, value := range values {
		categories[i] = DiscreteCategory{
			Value:         value,
			Count:         counts[i],
			Frequency:     float64(counts[i]) / float64(req.N),
			Probability:   probabilities[i],
			ExpectedCount: probabilities[i] * float64(req.N),
		}
		expectedCounts[i] = categories[i].ExpectedCount
		theoreticalMean += value * probabilities[i]
	}
	for i, value := range values {
		diff := value - theoreticalMean
		theoreticalVariance += diff * diff * probabilities[i]
	}

	stats := summarize(samples)
	meanCI, varianceCI := confidenceIntervals(stats, req.N, req.Alpha)
	if !finite(theoreticalMean) || !finite(theoreticalVariance) ||
		!finiteStats(stats) || !finiteInterval(meanCI) || !finiteInterval(varianceCI) {
		return DiscreteResult{}, ErrNumeric
	}
	for _, category := range categories {
		if !finite(category.ExpectedCount) {
			return DiscreteResult{}, ErrNumeric
		}
	}

	return DiscreteResult{
		SampleSize:          req.N,
		Preview:             preview(samples),
		Stats:               stats,
		TheoreticalMean:     theoreticalMean,
		TheoreticalVariance: theoreticalVariance,
		MeanCI:              meanCI,
		VarianceCI:          varianceCI,
		Categories:          categories,
		Test:                chiSquareGoodnessOfFit(counts, expectedCounts, req.Alpha),
	}, nil
}

func validateSample(n int) error {
	if n < MinSamples || n > MaxSamples {
		return fmt.Errorf("%w: получено %d", ErrSampleSize, n)
	}
	return nil
}

func finiteStats(stats SampleStats) bool {
	return finite(stats.Mean) && finite(stats.Variance) &&
		finite(stats.StdDev) && finite(stats.Min) && finite(stats.Max)
}

func finiteInterval(interval Interval) bool {
	return finite(interval.Lower) && finite(interval.Upper) &&
		finite(interval.Confidence)
}

func validateAlpha(alpha float64) error {
	if !finite(alpha) || alpha < MinAlpha || alpha >= 0.5 {
		return fmt.Errorf("%w: получено %v", ErrAlpha, alpha)
	}
	return nil
}

func validateDiscrete(values, probabilities []float64) ([]float64, []float64, error) {
	if len(values) != len(probabilities) {
		return nil, nil, fmt.Errorf("%w: значений %d, вероятностей %d",
			ErrProbabilities, len(values), len(probabilities))
	}
	if len(values) < 2 {
		return nil, nil, ErrValues
	}
	if len(values) > MaxCategories {
		return nil, nil, fmt.Errorf("%w: максимум %d", ErrValues, MaxCategories)
	}

	seen := make([]float64, 0, len(values))
	sum := 0.0
	for i := range values {
		value := values[i]
		p := probabilities[i]
		if !finite(value) {
			return nil, nil, fmt.Errorf("%w: y_%d не является конечным числом", ErrValues, i+1)
		}
		for _, other := range seen {
			if math.Abs(value-other) <= 1e-12 {
				return nil, nil, fmt.Errorf("%w: значение %v повторяется", ErrValues, value)
			}
		}
		seen = append(seen, value)

		if !finite(p) || p <= 0 || p > 1 {
			return nil, nil, fmt.Errorf("%w: p_%d = %v", ErrProbabilities, i+1, p)
		}
		sum += p
	}

	if math.Abs(sum-1) > 1e-6 {
		return nil, nil, fmt.Errorf("%w: получено %v", ErrSum, sum)
	}

	// После проверки нормируем: это убирает накопленную ошибку сложения и
	// гарантирует, что последний отрезок действительно доходит до 1.
	normalized := make([]float64, len(probabilities))
	for i, p := range probabilities {
		normalized[i] = p / sum
	}

	return append([]float64(nil), values...), normalized, nil
}

func pickCategory(u float64, cumulative []float64) int {
	index := sort.Search(len(cumulative), func(i int) bool {
		return cumulative[i] > u
	})
	if index >= len(cumulative) {
		return len(cumulative) - 1
	}
	return index
}

func preview(samples []float64) []float64 {
	const limit = 50
	if len(samples) <= limit {
		return append([]float64(nil), samples...)
	}
	return append([]float64(nil), samples[:limit]...)
}

// continuousHistogram строит интервалы равной ширины. Крайние интервалы
// включают хвосты распределения, поэтому сумма ожидаемых частот равна N и
// критерий χ² можно применять без отбрасывания части выборки.
func continuousHistogram(d distribution, samples []float64, bins int) []ContinuousBin {
	lo, hi := d.support()
	const tail = 1e-3

	start := lo
	if math.IsInf(lo, -1) {
		start = d.inverse(tail)
	}
	end := hi
	if math.IsInf(hi, 1) {
		end = d.inverse(1 - tail)
	}

	width := (end - start) / float64(bins)
	edges := make([]float64, bins+1)
	for i := range edges {
		edges[i] = start + float64(i)*width
	}

	histogram := make([]ContinuousBin, bins)
	for i := range histogram {
		lower := edges[i]
		upper := edges[i+1]
		probabilityLower := lower
		probabilityUpper := upper
		lowerUnbounded := false
		upperUnbounded := false

		if i == 0 {
			lower = lo
			probabilityLower = lo
			lowerUnbounded = math.IsInf(lo, -1)
			if lowerUnbounded {
				// В JSON нельзя передать -Inf, поэтому для отображения
				// оставляем конечную границу, а флаг говорит о хвосте.
				lower = start
			}
		}
		if i == bins-1 {
			upper = hi
			probabilityUpper = hi
			upperUnbounded = math.IsInf(hi, 1)
			if upperUnbounded {
				upper = end
			}
		}

		histogram[i] = ContinuousBin{
			Lower:          lower,
			Upper:          upper,
			LowerUnbounded: lowerUnbounded,
			UpperUnbounded: upperUnbounded,
			Probability:    d.cdf(probabilityUpper) - d.cdf(probabilityLower),
		}
	}

	for _, sample := range samples {
		index := sort.Search(len(edges), func(i int) bool {
			return edges[i] > sample
		}) - 1
		if index < 0 {
			index = 0
		}
		if index >= bins {
			index = bins - 1
		}
		histogram[index].Count++
	}

	for i := range histogram {
		histogram[i].Frequency = float64(histogram[i].Count) / float64(len(samples))
		histogram[i].ExpectedCount = histogram[i].Probability * float64(len(samples))
	}

	return histogram
}

// chiSquareGoodnessOfFit считает критерий Пирсона. Редкие интервалы
// (ожидаемая частота меньше 5) объединяются с соседними, иначе приближение χ²
// работает плохо. Параметры распределения заданы, а не оценены по выборке,
// поэтому число степеней свободы равно числу групп минус один.
func chiSquareGoodnessOfFit(observed []int, expected []float64, alpha float64) GoodnessOfFit {
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

	name := "Критерий согласия χ² Пирсона"
	if len(groups) < 2 {
		return GoodnessOfFit{
			Name:     name,
			Alpha:    alpha,
			Accepted: false,
			Valid:    false,
			Groups:   len(groups),
		}
	}

	statistic := 0.0
	for _, group := range groups {
		diff := float64(group.observed) - group.expected
		statistic += diff * diff / group.expected
	}

	df := len(groups) - 1
	pValue := distuv.ChiSquared{K: float64(df)}.Survival(statistic)

	return GoodnessOfFit{
		Name:             fmt.Sprintf("%s, групп: %d", name, len(groups)),
		Statistic:        statistic,
		DegreesOfFreedom: df,
		PValue:           pValue,
		Alpha:            alpha,
		Accepted:         pValue >= alpha,
		Valid:            true,
		Groups:           len(groups),
	}
}
