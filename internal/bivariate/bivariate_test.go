package bivariate

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

func near(got, want, tolerance float64) bool {
	return math.Abs(got-want) <= tolerance
}

func findTest(t *testing.T, tests []Test, name string) Test {
	t.Helper()
	for _, test := range tests {
		if test.Name == name {
			return test
		}
	}
	t.Fatalf("критерий %q не найден", name)
	return Test{}
}

func TestMarginalDensityMatchesJointIntegral(t *testing.T) {
	for _, x := range []float64{0, 1, -1.5, 2} {
		a := math.Sqrt(9 - x*x)
		integral := simpson(func(y float64) float64 {
			return jointDensity(x, y)
		}, -a, a, 512)
		if got := marginalDensity(x); !near(got, integral, 1e-6) {
			t.Errorf("f_X(%v) = %.12f, интеграл совместной плотности = %.12f", x, got, integral)
		}
	}
}

func TestConditionalDensityNormalizes(t *testing.T) {
	for _, x0 := range []float64{0, 1, 2} {
		a := math.Sqrt(9 - x0*x0)
		integral := simpson(func(y float64) float64 {
			return conditionalDensity(x0, y)
		}, -a, a, 512)
		if !near(integral, 1, 1e-6) {
			t.Errorf("интеграл f(y|x=%v) = %.12f, ожидалось 1", x0, integral)
		}
	}
}

func TestContinuousSampler(t *testing.T) {
	const n = 200_000
	got, err := simulateContinuous(ContinuousRequest{N: n, Bins: 20, Alpha: 0.05}, rand.New(rand.NewPCG(1, 2)))
	if err != nil {
		t.Fatalf("simulateContinuous: %v", err)
	}

	if math.Abs(got.Stats.MeanX) >= 0.02 {
		t.Errorf("meanX = %v", got.Stats.MeanX)
	}
	if math.Abs(got.Stats.MeanY) >= 0.02 {
		t.Errorf("meanY = %v", got.Stats.MeanY)
	}
	variance := 297.0 / 140
	if math.Abs(got.Stats.VarianceX-variance) >= 0.02*variance {
		t.Errorf("varianceX = %v, теория = %v", got.Stats.VarianceX, variance)
	}
	if math.Abs(got.Stats.VarianceY-variance) >= 0.02*variance {
		t.Errorf("varianceY = %v, теория = %v", got.Stats.VarianceY, variance)
	}
	if math.Abs(got.Stats.Correlation) >= 0.02 {
		t.Errorf("correlation = %v", got.Stats.Correlation)
	}

	sum := 0
	for _, bin := range got.HistogramX {
		sum += bin.Count
	}
	if sum != n {
		t.Errorf("сумма частот гистограммы X = %d, ожидалось %d", sum, n)
	}

	rng := rand.New(rand.NewPCG(1, 2))
	for range n {
		point := sampleDisk(rng)
		if math.Hypot(point.X, point.Y) > diskRadius+1e-12 {
			t.Fatalf("точка вне круга: (%v, %v)", point.X, point.Y)
		}
	}
}

func TestContinuousHypothesisTests(t *testing.T) {
	got, err := simulateContinuous(ContinuousRequest{N: 100_000, Bins: 20, Alpha: 0.001}, rand.New(rand.NewPCG(3, 4)))
	if err != nil {
		t.Fatalf("simulateContinuous: %v", err)
	}

	for _, name := range []string{
		"Согласие X с теоретической плотностью",
		"Согласие Y с теоретической плотностью",
		"Корреляция",
	} {
		test := findTest(t, got.Tests, name)
		if !test.Valid || !test.Accepted {
			t.Errorf("%q: valid = %v, accepted = %v, p = %v", name, test.Valid, test.Accepted, test.PValue)
		}
	}

	independence := findTest(t, got.Tests, "Независимость X и Y")
	if !independence.Valid || independence.Accepted {
		t.Errorf("независимость: valid = %v, accepted = %v, p = %v",
			independence.Valid, independence.Accepted, independence.PValue)
	}
}

func TestDiscreteSampler(t *testing.T) {
	got, err := simulateDiscrete(DiscreteRequest{N: 200_000, Alpha: 0.01}, rand.New(rand.NewPCG(5, 6)))
	if err != nil {
		t.Fatalf("simulateDiscrete: %v", err)
	}
	theory := got.Theoretical

	if math.Abs(got.Stats.MeanX-theory.MeanX) >= 0.01 {
		t.Errorf("meanX = %v, теория = %v", got.Stats.MeanX, theory.MeanX)
	}
	if math.Abs(got.Stats.MeanY-theory.MeanY) >= 0.01 {
		t.Errorf("meanY = %v, теория = %v", got.Stats.MeanY, theory.MeanY)
	}
	if math.Abs(got.Stats.VarianceX-theory.VarianceX) >= 0.01*theory.VarianceX {
		t.Errorf("varianceX = %v, теория = %v", got.Stats.VarianceX, theory.VarianceX)
	}
	if math.Abs(got.Stats.VarianceY-theory.VarianceY) >= 0.01*theory.VarianceY {
		t.Errorf("varianceY = %v, теория = %v", got.Stats.VarianceY, theory.VarianceY)
	}
	if math.Abs(got.Stats.Correlation-theory.Correlation) >= 0.01 {
		t.Errorf("correlation = %v, теория = %v", got.Stats.Correlation, theory.Correlation)
	}

	for _, row := range append(append([]ConditionalRow(nil), got.ConditionalXGivenY...), got.ConditionalYGivenX...) {
		theoretical := 0.0
		observed := 0.0
		for i := range row.Values {
			theoretical += row.Theoretical[i]
			observed += row.Observed[i]
		}
		if !near(theoretical, 1, 1e-9) {
			t.Errorf("условные вероятности при %v в сумме дают %v", row.Fixed, theoretical)
		}
		if !near(observed, 1, 1e-9) {
			t.Errorf("условные частоты при %v в сумме дают %v", row.Fixed, observed)
		}
	}

	if test := findTest(t, got.Tests, "Независимость X и Y"); !test.Valid || test.Accepted {
		t.Errorf("независимость: valid = %v, accepted = %v, p = %v", test.Valid, test.Accepted, test.PValue)
	}
	if test := findTest(t, got.Tests, "Корреляция"); !test.Valid || !test.Accepted {
		t.Errorf("корреляция: valid = %v, accepted = %v, p = %v", test.Valid, test.Accepted, test.PValue)
	}
}

func TestResultsAreFinite(t *testing.T) {
	continuous, err := simulateContinuous(
		ContinuousRequest{N: 1000, Bins: 10, Alpha: 0.05},
		rand.New(rand.NewPCG(7, 8)),
	)
	if err != nil {
		t.Fatalf("simulateContinuous: %v", err)
	}
	if !finiteContinuous(continuous) {
		t.Fatal("результат непрерывной имитации содержит нечисловые значения")
	}
	if len(continuous.Tests) != 8 {
		t.Errorf("непрерывных критериев %d, ожидалось 8", len(continuous.Tests))
	}
	checkTests(t, continuous.Tests, 0.05)
	checkIntervals(t, continuous.Intervals, 0.95)
	if len(continuous.Preview) != 20 {
		t.Errorf("preview содержит %d точек, ожидалось 20", len(continuous.Preview))
	}

	discrete, err := simulateDiscrete(
		DiscreteRequest{N: 1000, Alpha: 0.05},
		rand.New(rand.NewPCG(9, 10)),
	)
	if err != nil {
		t.Fatalf("simulateDiscrete: %v", err)
	}
	if !finiteDiscrete(discrete) {
		t.Fatal("результат дискретной имитации содержит нечисловые значения")
	}
	if len(discrete.Tests) != 8 {
		t.Errorf("дискретных критериев %d, ожидалось 8", len(discrete.Tests))
	}
	checkTests(t, discrete.Tests, 0.05)
	checkIntervals(t, discrete.Intervals, 0.95)
	if len(discrete.Preview) != 20 {
		t.Errorf("preview содержит %d точек, ожидалось 20", len(discrete.Preview))
	}
}

func checkTests(t *testing.T, tests []Test, alpha float64) {
	t.Helper()
	for _, test := range tests {
		if test.Name == "" {
			t.Error("критерий без названия")
		}
		if test.Alpha != alpha {
			t.Errorf("%q: alpha = %v, ожидалось %v", test.Name, test.Alpha, alpha)
		}
		if test.PValue < 0 || test.PValue > 1 {
			t.Errorf("%q: p = %v вне [0, 1]", test.Name, test.PValue)
		}
		if test.DegreesOfFreedom < 0 {
			t.Errorf("%q: df = %d", test.Name, test.DegreesOfFreedom)
		}
		if test.Accepted && !test.Valid {
			t.Errorf("%q: принят при valid = false", test.Name)
		}
	}
}

func checkIntervals(t *testing.T, intervals Intervals, confidence float64) {
	t.Helper()
	for _, interval := range []Interval{
		intervals.MeanX, intervals.MeanY,
		intervals.VarianceX, intervals.VarianceY,
		intervals.Correlation,
	} {
		if !finiteInterval(interval) {
			t.Errorf("интервал содержит нечисловые значения: %+v", interval)
		}
		if interval.Confidence != confidence {
			t.Errorf("доверительная вероятность %v, ожидалось %v", interval.Confidence, confidence)
		}
	}
}

func TestValidation(t *testing.T) {
	service := NewBivariateService()

	continuousCases := []struct {
		req  ContinuousRequest
		want error
	}{
		{ContinuousRequest{N: 5, Bins: 10, Alpha: 0.05}, ErrSampleSize},
		{ContinuousRequest{N: 1_000_001, Bins: 10, Alpha: 0.05}, ErrSampleSize},
		{ContinuousRequest{N: 1000, Bins: 1, Alpha: 0.05}, ErrBins},
		{ContinuousRequest{N: 1000, Bins: 51, Alpha: 0.05}, ErrBins},
		{ContinuousRequest{N: 1000, Bins: 10, Alpha: 0}, ErrAlpha},
		{ContinuousRequest{N: 1000, Bins: 10, Alpha: 0.5}, ErrAlpha},
		{ContinuousRequest{N: 1000, Bins: 10, Alpha: 1e-7}, ErrAlpha},
	}
	for _, tc := range continuousCases {
		if _, err := service.SimulateContinuous(tc.req); !errors.Is(err, tc.want) {
			t.Errorf("SimulateContinuous(%+v) = %v, ожидалась ошибка %v", tc.req, err, tc.want)
		}
	}

	discreteCases := []struct {
		req  DiscreteRequest
		want error
	}{
		{DiscreteRequest{N: 5, Alpha: 0.05}, ErrSampleSize},
		{DiscreteRequest{N: 1_000_001, Alpha: 0.05}, ErrSampleSize},
		{DiscreteRequest{N: 1000, Alpha: 0}, ErrAlpha},
		{DiscreteRequest{N: 1000, Alpha: 0.5}, ErrAlpha},
	}
	for _, tc := range discreteCases {
		if _, err := service.SimulateDiscrete(tc.req); !errors.Is(err, tc.want) {
			t.Errorf("SimulateDiscrete(%+v) = %v, ожидалась ошибка %v", tc.req, err, tc.want)
		}
	}
}
