package randomvars

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
)

const tolerance = 1e-6

func mustDistribution(t *testing.T, law string, a, b, lambda float64) distribution {
	t.Helper()
	d, err := newDistribution(law, a, b, lambda)
	if err != nil {
		t.Fatalf("newDistribution(%q): %v", law, err)
	}
	return d
}

func TestInverseIsInverseOfCDF(t *testing.T) {
	cases := []struct {
		law          string
		a, b, lambda float64
	}{
		{"uniform", 2, 7, 0},
		{"exponential", 0, 0, 0.7},
	}

	for _, tc := range cases {
		d := mustDistribution(t, tc.law, tc.a, tc.b, tc.lambda)
		for _, x := range []float64{1e-6, 0.001, 0.1, 0.5, 0.9, 0.999, 1 - 1e-6} {
			y := d.inverse(x)
			if got := d.cdf(y); math.Abs(got-x) > 1e-9 {
				t.Errorf("%s: G(G^-1(%v)) = %v", d.law, x, got)
			}
		}
	}
}

func TestCDFBoundaries(t *testing.T) {
	uniform := mustDistribution(t, "uniform", 0, 2, 0)
	if uniform.cdf(-1) != 0 || uniform.cdf(3) != 1 {
		t.Error("uniform CDF does not clamp to [0, 1]")
	}

	exponential := mustDistribution(t, "exponential", 0, 0, 2)
	if exponential.cdf(-1) != 0 {
		t.Error("exponential CDF is non-zero for negative argument")
	}
}

func TestContinuousMoments(t *testing.T) {
	s := &RandomVariableService{}
	const n = 200_000

	cases := []struct {
		name string
		req  ContinuousRequest
	}{
		{"uniform", ContinuousRequest{Law: "uniform", A: -2, B: 4, N: n, Bins: 10, Alpha: 0.05}},
		{"exponential", ContinuousRequest{Law: "exponential", Lambda: 1.5, N: n, Bins: 10, Alpha: 0.05}},
	}

	for _, tc := range cases {
		got, err := s.SimulateContinuous(tc.req)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if math.Abs(got.Stats.Mean-got.TheoreticalMean) > 0.02*math.Max(1, math.Abs(got.TheoreticalMean)) {
			t.Errorf("%s: mean = %v, theory = %v", tc.name, got.Stats.Mean, got.TheoreticalMean)
		}
		if math.Abs(got.Stats.Variance-got.TheoreticalVariance) > 0.05*math.Max(1, got.TheoreticalVariance) {
			t.Errorf("%s: variance = %v, theory = %v", tc.name, got.Stats.Variance, got.TheoreticalVariance)
		}
	}
}

func TestContinuousResultsAreJSONSerializable(t *testing.T) {
	s := &RandomVariableService{}
	cases := []ContinuousRequest{
		{Law: "uniform", A: -1, B: 2, N: 1000, Bins: 8, Alpha: 0.05},
		{Law: "exponential", Lambda: 2, N: 1000, Bins: 8, Alpha: 0.05},
	}

	for _, req := range cases {
		got, err := s.SimulateContinuous(req)
		if err != nil {
			t.Fatalf("%s: %v", req.Law, err)
		}
		for _, bin := range got.Histogram {
			if !finite(bin.Lower) || !finite(bin.Upper) {
				t.Errorf("%s: bin bounds are not finite: %+v", req.Law, bin)
			}
		}
		if _, err := json.Marshal(got); err != nil {
			t.Errorf("%s: result is not JSON serializable: %v", req.Law, err)
		}
	}
}

func TestContinuousHistogramCoversSample(t *testing.T) {
	s := &RandomVariableService{}
	got, err := s.SimulateContinuous(ContinuousRequest{
		Law: "exponential", Lambda: 1, N: 50_000, Bins: 12, Alpha: 0.05,
	})
	if err != nil {
		t.Fatal(err)
	}

	total := 0
	expected := 0.0
	for _, bin := range got.Histogram {
		total += bin.Count
		expected += bin.ExpectedCount
	}
	if total != got.SampleSize {
		t.Errorf("histogram counts = %d, want %d", total, got.SampleSize)
	}
	if math.Abs(expected-float64(got.SampleSize)) > 1e-6 {
		t.Errorf("expected counts sum = %v, want %d", expected, got.SampleSize)
	}
	if len(got.Histogram) != 12 {
		t.Errorf("bins = %d, want 12", len(got.Histogram))
	}
	if got.Test.Valid && got.Test.DegreesOfFreedom != len(got.Histogram)-1 {
		t.Errorf("df = %d, want %d", got.Test.DegreesOfFreedom, len(got.Histogram)-1)
	}
}

func TestContinuousConfidenceIntervalsContainEstimates(t *testing.T) {
	s := &RandomVariableService{}
	got, err := s.SimulateContinuous(ContinuousRequest{
		Law: "uniform", A: 0, B: 1, N: 10_000, Bins: 10, Alpha: 0.05,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.MeanCI.Lower >= got.Stats.Mean || got.MeanCI.Upper <= got.Stats.Mean {
		t.Errorf("mean CI %+v does not contain mean %v", got.MeanCI, got.Stats.Mean)
	}
	if got.VarianceCI.Lower >= got.Stats.Variance || got.VarianceCI.Upper <= got.Stats.Variance {
		t.Errorf("variance CI %+v does not contain variance %v", got.VarianceCI, got.Stats.Variance)
	}
	if got.MeanCI.Lower >= got.MeanCI.Upper || got.VarianceCI.Lower >= got.VarianceCI.Upper {
		t.Error("confidence interval bounds are not ordered")
	}
	if math.Abs(got.MeanCI.Confidence-0.95) > 1e-12 {
		t.Errorf("confidence = %v, want 0.95", got.MeanCI.Confidence)
	}
}

func TestDiscreteFrequencies(t *testing.T) {
	s := &RandomVariableService{}
	values := []float64{1, 2, 5, 10}
	probabilities := []float64{0.1, 0.2, 0.3, 0.4}
	got, err := s.SimulateDiscrete(DiscreteRequest{
		Values: values, Probabilities: probabilities, N: 200_000, Alpha: 0.05,
	})
	if err != nil {
		t.Fatal(err)
	}

	total := 0
	for i, category := range got.Categories {
		if math.Abs(category.Frequency-probabilities[i]) > 0.01 {
			t.Errorf("value %v: frequency = %v, want %v", category.Value, category.Frequency, probabilities[i])
		}
		total += category.Count
	}
	if total != got.SampleSize {
		t.Errorf("counts = %d, want %d", total, got.SampleSize)
	}

	wantMean := 1*0.1 + 2*0.2 + 5*0.3 + 10*0.4
	if math.Abs(got.TheoreticalMean-wantMean) > 1e-12 {
		t.Errorf("theoretical mean = %v, want %v", got.TheoreticalMean, wantMean)
	}
	if math.Abs(got.Stats.Mean-wantMean) > 0.03 {
		t.Errorf("sample mean = %v, want ~%v", got.Stats.Mean, wantMean)
	}
	if !got.Test.Valid {
		t.Errorf("goodness-of-fit is not applicable: %+v", got.Test)
	}
	if got.Test.PValue < 0 || got.Test.PValue > 1 {
		t.Errorf("p-value out of range: %v", got.Test.PValue)
	}
}

func TestDiscreteNormalizesTinyRoundingError(t *testing.T) {
	s := &RandomVariableService{}
	got, err := s.SimulateDiscrete(DiscreteRequest{
		Values: []float64{1, 2, 3}, Probabilities: []float64{0.1, 0.2, 0.699999}, N: 10_000, Alpha: 0.05,
	})
	if err != nil {
		t.Fatal(err)
	}
	sum := 0.0
	for _, category := range got.Categories {
		sum += category.Probability
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Errorf("normalized probabilities sum to %v", sum)
	}
}

func TestDiscreteRejectsBadInput(t *testing.T) {
	s := &RandomVariableService{}

	cases := []struct {
		name string
		req  DiscreteRequest
		want error
	}{
		{"too few values", DiscreteRequest{Values: []float64{1}, Probabilities: []float64{1}, N: 100, Alpha: 0.05}, ErrValues},
		{"different lengths", DiscreteRequest{Values: []float64{1, 2}, Probabilities: []float64{1}, N: 100, Alpha: 0.05}, ErrProbabilities},
		{"duplicate values", DiscreteRequest{Values: []float64{1, 1}, Probabilities: []float64{0.5, 0.5}, N: 100, Alpha: 0.05}, ErrValues},
		{"bad probability", DiscreteRequest{Values: []float64{1, 2}, Probabilities: []float64{0, 1}, N: 100, Alpha: 0.05}, ErrProbabilities},
		{"sum is not one", DiscreteRequest{Values: []float64{1, 2}, Probabilities: []float64{0.4, 0.4}, N: 100, Alpha: 0.05}, ErrSum},
		{"small n", DiscreteRequest{Values: []float64{1, 2}, Probabilities: []float64{0.5, 0.5}, N: 1, Alpha: 0.05}, ErrSampleSize},
		{"bad alpha", DiscreteRequest{Values: []float64{1, 2}, Probabilities: []float64{0.5, 0.5}, N: 100, Alpha: 0.7}, ErrAlpha},
	}

	for _, tc := range cases {
		_, err := s.SimulateDiscrete(tc.req)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
}

func TestContinuousRejectsBadParameters(t *testing.T) {
	s := &RandomVariableService{}

	cases := []struct {
		name string
		req  ContinuousRequest
	}{
		{"unknown law", ContinuousRequest{Law: "normal", N: 100, Bins: 10, Alpha: 0.05}},
		{"removed rayleigh", ContinuousRequest{Law: "rayleigh", N: 100, Bins: 10, Alpha: 0.05}},
		{"uniform order", ContinuousRequest{Law: "uniform", A: 2, B: 1, N: 100, Bins: 10, Alpha: 0.05}},
		{"zero lambda", ContinuousRequest{Law: "exponential", Lambda: 0, N: 100, Bins: 10, Alpha: 0.05}},
		{"bad bins", ContinuousRequest{Law: "uniform", A: 0, B: 1, N: 100, Bins: 1, Alpha: 0.05}},
		{"overflowing width", ContinuousRequest{Law: "uniform", A: -1e308, B: 1e308, N: 100, Bins: 10, Alpha: 0.05}},
		{"underflowing lambda", ContinuousRequest{Law: "exponential", Lambda: 1e-310, N: 100, Bins: 10, Alpha: 0.05}},
		{"too small alpha", ContinuousRequest{Law: "uniform", A: 0, B: 1, N: 100, Bins: 10, Alpha: 1e-16}},
	}

	for _, tc := range cases {
		if _, err := s.SimulateContinuous(tc.req); err == nil {
			t.Errorf("%s: error expected", tc.name)
		}
	}
}

func TestConfidenceIntervalsMatchReference(t *testing.T) {
	stats := SampleStats{Mean: 10, Variance: 4, StdDev: 2}
	meanCI, varianceCI := confidenceIntervals(stats, 11, 0.05)

	// t_{0.975, 10} = 2.2281388519649385
	const tQuantile = 2.2281388519649385
	halfWidth := tQuantile * 2 / math.Sqrt(11)
	if got := meanCI.Lower; math.Abs(got-(10-halfWidth)) > 1e-9 {
		t.Errorf("mean CI lower = %v, want %v", got, 10-halfWidth)
	}
	if got := meanCI.Upper; math.Abs(got-(10+halfWidth)) > 1e-9 {
		t.Errorf("mean CI upper = %v, want %v", got, 10+halfWidth)
	}

	// χ²_{0.025, 10} и χ²_{0.975, 10}
	const chiLow, chiHigh = 3.2469727802368409, 20.483177350807399
	if got := varianceCI.Lower; math.Abs(got-40/chiHigh) > 1e-9 {
		t.Errorf("variance CI lower = %v, want %v", got, 40/chiHigh)
	}
	if got := varianceCI.Upper; math.Abs(got-40/chiLow) > 1e-9 {
		t.Errorf("variance CI upper = %v, want %v", got, 40/chiLow)
	}
}

func TestConfidenceIntervalsAtMaxSampleSize(t *testing.T) {
	stats := SampleStats{Mean: 1, Variance: 1, StdDev: 1}
	meanCI, varianceCI := confidenceIntervals(stats, MaxSamples, 0.05)

	if !finiteInterval(meanCI) || !finiteInterval(varianceCI) {
		t.Fatalf("intervals are not finite: %+v %+v", meanCI, varianceCI)
	}
	if meanCI.Lower >= meanCI.Upper || varianceCI.Lower >= varianceCI.Upper {
		t.Errorf("confidence interval bounds are not ordered: %+v %+v", meanCI, varianceCI)
	}
}

func TestGoodnessOfFitAcceptsAndRejects(t *testing.T) {
	accepted := chiSquareGoodnessOfFit(
		[]int{18, 31, 51},
		[]float64{20, 30, 50},
		0.05,
	)
	if !accepted.Valid || !accepted.Accepted {
		t.Errorf("matching sample rejected: %+v", accepted)
	}

	rejected := chiSquareGoodnessOfFit(
		[]int{0, 0, 100},
		[]float64{33.3, 33.3, 33.4},
		0.05,
	)
	if !rejected.Valid || rejected.Accepted {
		t.Errorf("strongly mismatching sample accepted: %+v", rejected)
	}
}

func TestGoodnessOfFitMergesRareBins(t *testing.T) {
	got := chiSquareGoodnessOfFit(
		[]int{3, 5, 10},
		[]float64{4, 4, 10},
		0.05,
	)
	if got.Groups != 2 {
		t.Errorf("groups = %d, want 2", got.Groups)
	}
	if got.DegreesOfFreedom != 1 {
		t.Errorf("df = %d, want 1", got.DegreesOfFreedom)
	}
}

func TestGoodnessOfFitReportsInvalidForOneGroup(t *testing.T) {
	got := chiSquareGoodnessOfFit([]int{2, 3}, []float64{1, 1}, 0.05)
	if got.Valid {
		t.Errorf("test with one group reported as valid: %+v", got)
	}
}

func TestDiscretePickCategoryBoundaries(t *testing.T) {
	cumulative := []float64{0.2, 0.5, 1}
	cases := map[float64]int{
		0:     0,
		0.199: 0,
		0.2:   1,
		0.499: 1,
		0.5:   2,
		0.999: 2,
	}
	for u, want := range cases {
		if got := pickCategory(u, cumulative); got != want {
			t.Errorf("u=%v: category = %d, want %d", u, got, want)
		}
	}
}
