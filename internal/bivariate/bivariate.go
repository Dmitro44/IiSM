// Package bivariate имитирует двумерные случайные величины для лабораторной
// работы №3: непрерывную варианта 3 и дискретную с заданной матрицей.
package bivariate

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
)

const (
	MinSamples = 10
	MaxSamples = 1_000_000
	MinBins    = 2
	MaxBins    = 50
	MinAlpha   = 1e-6

	diskRadius    = 3.0
	jointGridSize = 5
	previewLimit  = 20
)

var (
	ErrSampleSize = errors.New("N должно быть от 10 до 1000000")
	ErrBins       = errors.New("число интервалов должно быть от 2 до 50")
	ErrAlpha      = errors.New("уровень значимости alpha должен быть в интервале [1e-6, 0.5)")
	ErrNumeric    = errors.New("результат содержит нечисловые значения")
)

// ContinuousRequest — параметры имитации непрерывной двумерной величины.
type ContinuousRequest struct {
	N     int     `json:"n"`
	Bins  int     `json:"bins"`
	Alpha float64 `json:"alpha"`
}

// DiscreteRequest — параметры имитации дискретной двумерной величины.
type DiscreteRequest struct {
	N     int     `json:"n"`
	Alpha float64 `json:"alpha"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// HistogramBin — интервал гистограммы одной компоненты.
type HistogramBin struct {
	Lower              float64 `json:"lower"`
	Upper              float64 `json:"upper"`
	Count              int     `json:"count"`
	Frequency          float64 `json:"frequency"`
	Density            float64 `json:"density"`
	TheoreticalDensity float64 `json:"theoreticalDensity"`
	ExpectedCount      float64 `json:"expectedCount"`
}

type CurvePoint struct {
	Value   float64 `json:"value"`
	Density float64 `json:"density"`
}

type ConditionalCurve struct {
	FixedX          float64      `json:"fixedX"`
	MarginalDensity float64      `json:"marginalDensity"`
	Points          []CurvePoint `json:"points"`
}

type DiscreteCategory struct {
	Value         float64 `json:"value"`
	Count         int     `json:"count"`
	Frequency     float64 `json:"frequency"`
	Probability   float64 `json:"probability"`
	ExpectedCount float64 `json:"expectedCount"`
}

type ConditionalRow struct {
	Fixed       float64   `json:"fixed"`
	Values      []float64 `json:"values"`
	Theoretical []float64 `json:"theoretical"`
	Observed    []float64 `json:"observed"`
}

type SampleStats struct {
	MeanX       float64 `json:"meanX"`
	MeanY       float64 `json:"meanY"`
	VarianceX   float64 `json:"varianceX"`
	VarianceY   float64 `json:"varianceY"`
	StdDevX     float64 `json:"stdDevX"`
	StdDevY     float64 `json:"stdDevY"`
	Covariance  float64 `json:"covariance"`
	Correlation float64 `json:"correlation"`
	MinX        float64 `json:"minX"`
	MaxX        float64 `json:"maxX"`
	MinY        float64 `json:"minY"`
	MaxY        float64 `json:"maxY"`
}

type TheoreticalStats struct {
	MeanX       float64 `json:"meanX"`
	MeanY       float64 `json:"meanY"`
	VarianceX   float64 `json:"varianceX"`
	VarianceY   float64 `json:"varianceY"`
	Covariance  float64 `json:"covariance"`
	Correlation float64 `json:"correlation"`
}

type Interval struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Confidence float64 `json:"confidence"`
}

type Intervals struct {
	MeanX       Interval `json:"meanX"`
	MeanY       Interval `json:"meanY"`
	VarianceX   Interval `json:"varianceX"`
	VarianceY   Interval `json:"varianceY"`
	Correlation Interval `json:"correlation"`
}

type Test struct {
	Name             string  `json:"name"`
	Statistic        float64 `json:"statistic"`
	DegreesOfFreedom int     `json:"degreesOfFreedom"`
	PValue           float64 `json:"pValue"`
	Alpha            float64 `json:"alpha"`
	Accepted         bool    `json:"accepted"`
	Valid            bool    `json:"valid"`
	Note             string  `json:"note"`
}

type JointCell struct {
	CenterX       float64 `json:"centerX"`
	CenterY       float64 `json:"centerY"`
	Count         int     `json:"count"`
	ExpectedCount float64 `json:"expectedCount"`
}

type ContinuousResult struct {
	SampleSize    int                `json:"sampleSize"`
	Alpha         float64            `json:"alpha"`
	Formula       string             `json:"formula"`
	Method        string             `json:"method"`
	Preview       []Point            `json:"preview"`
	Stats         SampleStats        `json:"stats"`
	Theoretical   TheoreticalStats   `json:"theoretical"`
	Intervals     Intervals          `json:"intervals"`
	HistogramX    []HistogramBin     `json:"histogramX"`
	HistogramY    []HistogramBin     `json:"histogramY"`
	Conditional   []ConditionalCurve `json:"conditional"`
	Joint         []JointCell        `json:"joint"`
	JointGridSize int                `json:"jointGridSize"`
	Tests         []Test             `json:"tests"`
}

type DiscreteResult struct {
	SampleSize         int                `json:"sampleSize"`
	Alpha              float64            `json:"alpha"`
	ValuesX            []float64          `json:"valuesX"`
	ValuesY            []float64          `json:"valuesY"`
	Probabilities      [][]float64        `json:"probabilities"`
	MarginalsX         []float64          `json:"marginalsX"`
	MarginalsY         []float64          `json:"marginalsY"`
	Preview            []Point            `json:"preview"`
	Stats              SampleStats        `json:"stats"`
	Theoretical        TheoreticalStats   `json:"theoretical"`
	Intervals          Intervals          `json:"intervals"`
	HistogramX         []DiscreteCategory `json:"histogramX"`
	HistogramY         []DiscreteCategory `json:"histogramY"`
	Observed           [][]int            `json:"observed"`
	Expected           [][]float64        `json:"expected"`
	ConditionalXGivenY []ConditionalRow   `json:"conditionalXGivenY"`
	ConditionalYGivenX []ConditionalRow   `json:"conditionalYGivenX"`
	Tests              []Test             `json:"tests"`
}

type BivariateService struct{}

func NewBivariateService() *BivariateService {
	return &BivariateService{}
}

func (s *BivariateService) SimulateContinuous(req ContinuousRequest) (ContinuousResult, error) {
	return simulateContinuous(req, newRNG())
}

func (s *BivariateService) SimulateDiscrete(req DiscreteRequest) (DiscreteResult, error) {
	return simulateDiscrete(req, newRNG())
}

func newRNG() *rand.Rand {
	return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
}

func validateSample(n int) error {
	if n < MinSamples || n > MaxSamples {
		return fmt.Errorf("%w: получено %d", ErrSampleSize, n)
	}
	return nil
}

func validateBins(bins int) error {
	if bins < MinBins || bins > MaxBins {
		return fmt.Errorf("%w: получено %d", ErrBins, bins)
	}
	return nil
}

func validateAlpha(alpha float64) error {
	if !finite(alpha) || alpha < MinAlpha || alpha >= 0.5 {
		return fmt.Errorf("%w: получено %v", ErrAlpha, alpha)
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func finiteStats(stats SampleStats) bool {
	return finite(stats.MeanX) && finite(stats.MeanY) &&
		finite(stats.VarianceX) && finite(stats.VarianceY) &&
		finite(stats.StdDevX) && finite(stats.StdDevY) &&
		finite(stats.Covariance) && finite(stats.Correlation) &&
		finite(stats.MinX) && finite(stats.MaxX) &&
		finite(stats.MinY) && finite(stats.MaxY)
}

func finiteTheory(stats TheoreticalStats) bool {
	return finite(stats.MeanX) && finite(stats.MeanY) &&
		finite(stats.VarianceX) && finite(stats.VarianceY) &&
		finite(stats.Covariance) && finite(stats.Correlation)
}

func finiteInterval(interval Interval) bool {
	return finite(interval.Lower) && finite(interval.Upper) && finite(interval.Confidence)
}

func finiteIntervals(intervals Intervals) bool {
	return finiteInterval(intervals.MeanX) && finiteInterval(intervals.MeanY) &&
		finiteInterval(intervals.VarianceX) && finiteInterval(intervals.VarianceY) &&
		finiteInterval(intervals.Correlation)
}

func finiteTests(tests []Test) bool {
	for _, test := range tests {
		if !finite(test.Statistic) || !finite(test.PValue) || !finite(test.Alpha) {
			return false
		}
	}
	return true
}

func finiteContinuous(result ContinuousResult) bool {
	if !finite(result.Alpha) || !finiteStats(result.Stats) ||
		!finiteTheory(result.Theoretical) || !finiteIntervals(result.Intervals) ||
		!finiteTests(result.Tests) {
		return false
	}
	for _, point := range result.Preview {
		if !finite(point.X) || !finite(point.Y) {
			return false
		}
	}
	for _, bin := range append(append([]HistogramBin(nil), result.HistogramX...), result.HistogramY...) {
		if !finite(bin.Lower) || !finite(bin.Upper) || !finite(bin.Frequency) ||
			!finite(bin.Density) || !finite(bin.TheoreticalDensity) || !finite(bin.ExpectedCount) {
			return false
		}
	}
	for _, curve := range result.Conditional {
		if !finite(curve.FixedX) || !finite(curve.MarginalDensity) {
			return false
		}
		for _, point := range curve.Points {
			if !finite(point.Value) || !finite(point.Density) {
				return false
			}
		}
	}
	for _, cell := range result.Joint {
		if !finite(cell.CenterX) || !finite(cell.CenterY) || !finite(cell.ExpectedCount) {
			return false
		}
	}
	return true
}

func finiteDiscrete(result DiscreteResult) bool {
	if !finite(result.Alpha) || !finiteStats(result.Stats) ||
		!finiteTheory(result.Theoretical) || !finiteIntervals(result.Intervals) ||
		!finiteTests(result.Tests) {
		return false
	}
	for _, values := range [][]float64{result.ValuesX, result.ValuesY, result.MarginalsX, result.MarginalsY} {
		for _, value := range values {
			if !finite(value) {
				return false
			}
		}
	}
	for _, row := range result.Probabilities {
		for _, value := range row {
			if !finite(value) {
				return false
			}
		}
	}
	for _, row := range result.Expected {
		for _, value := range row {
			if !finite(value) {
				return false
			}
		}
	}
	for _, point := range result.Preview {
		if !finite(point.X) || !finite(point.Y) {
			return false
		}
	}
	for _, category := range append(append([]DiscreteCategory(nil), result.HistogramX...), result.HistogramY...) {
		if !finite(category.Value) || !finite(category.Frequency) ||
			!finite(category.Probability) || !finite(category.ExpectedCount) {
			return false
		}
	}
	for _, row := range append(append([]ConditionalRow(nil), result.ConditionalXGivenY...), result.ConditionalYGivenX...) {
		if !finite(row.Fixed) {
			return false
		}
		for _, values := range [][]float64{row.Values, row.Theoretical, row.Observed} {
			for _, value := range values {
				if !finite(value) {
					return false
				}
			}
		}
	}
	return true
}
