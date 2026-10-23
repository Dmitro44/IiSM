package events

import (
	"errors"
	"math"
	"testing"
)

const tol = 0.01

func TestSimpleConverges(t *testing.T) {
	s := &EventService{}
	for _, p := range []float32{0.1, 0.5, 0.9} {
		got, err := s.SimulateSimple(p, 200000)
		if err != nil {
			t.Fatalf("p=%v: %v", p, err)
		}
		if math.Abs(got.Outcomes[0].Frequency-float64(p)) > tol {
			t.Errorf("p=%v empirical=%v", p, got.Outcomes[0].Frequency)
		}
	}
}

func TestComplexConverges(t *testing.T) {
	s := &EventService{}
	ps := []float32{0.3, 0.5, 0.7}
	got, err := s.SimulateComplex(ps, 200000)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Outcomes) != 1<<len(ps) {
		t.Fatalf("outcomes = %d, want %d", len(got.Outcomes), 1<<len(ps))
	}

	// Совместные частоты сходятся к произведениям -- это и есть проверка
	// независимости: подмена k независимых розыгрышей одним общим числом
	// оставила бы маргинали верными, но испортила бы эти частоты.
	total := 0
	probabilities := 0.0
	for i, o := range got.Outcomes {
		if math.Abs(o.Frequency-o.Probability) > tol {
			t.Errorf("outcome %d: freq=%v theory=%v", i, o.Frequency, o.Probability)
		}
		total += o.Count
		probabilities += o.Probability
	}

	// Исходы несовместны и покрывают всё пространство: в каждом испытании
	// выпадает ровно один, поэтому сумма счётчиков обязана равняться N.
	if total != got.SampleSize {
		t.Errorf("counts sum to %d, want %d", total, got.SampleSize)
	}
	if math.Abs(probabilities-1) > 1e-6 {
		t.Errorf("probabilities sum to %v, want 1", probabilities)
	}
}

func TestComplexProbabilitiesAreProducts(t *testing.T) {
	s := &EventService{}
	ps := []float32{0.3, 0.5, 0.7}
	got, err := s.SimulateComplex(ps, 1000)
	if err != nil {
		t.Fatal(err)
	}

	// Вероятность исхода -- произведение вероятностей произошедших событий и их
	// дополнений. Проверяется без оглядки на частоты: сумма произведений всегда
	// равна единице, и ошибиться в самой формуле можно, не нарушив сумму.
	for mask, o := range got.Outcomes {
		want := 1.0
		for i, p := range ps {
			if mask&(1<<i) != 0 {
				want *= float64(p)
			} else {
				want *= 1 - float64(p)
			}
		}
		if math.Abs(o.Probability-want) > 1e-6 {
			t.Errorf("mask %03b: probability=%v want %v", mask, o.Probability, want)
		}
	}
}

func TestComplexRejectsLongList(t *testing.T) {
	s := &EventService{}
	_, err := s.SimulateComplex(make([]float32, MaxComplexEvents+1), 1000)
	if !errors.Is(err, ErrTooManyEvents) {
		t.Errorf("k=%d: err=%v, want ErrTooManyEvents", MaxComplexEvents+1, err)
	}
}

func TestDependentConverges(t *testing.T) {
	s := &EventService{}
	got, err := s.SimulateDependent(0.5, 0.8, 200000)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, o := range got.Outcomes {
		if math.Abs(o.Frequency-o.Probability) > tol {
			t.Errorf("empirical=%v theory=%v", o.Frequency, o.Probability)
		}
		total += o.Count
	}
	if total != got.SampleSize {
		t.Errorf("counts sum to %d, want %d", total, got.SampleSize)
	}
	// 0.8 has no exact binary form, so P(AB) is 0.40000000596
	if math.Abs(got.Outcomes[0].Probability-0.4) > 1e-6 {
		t.Errorf("P(AB)=%v want ~0.4", got.Outcomes[0].Probability)
	}
}

func TestFullGroupConverges(t *testing.T) {
	s := &EventService{}
	ps := []float32{0.2, 0.3, 0.5}
	got, err := s.SimulateFullGroup(ps, 200000)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for i, o := range got.Outcomes {
		if math.Abs(o.Frequency-float64(ps[i])) > tol {
			t.Errorf("p[%d]=%v empirical=%v", i, ps[i], o.Frequency)
		}
		total += o.Count
	}
	if total != got.SampleSize {
		t.Errorf("counts sum to %d, want %d", total, got.SampleSize)
	}
}

func TestRejectsBadInput(t *testing.T) {
	s := &EventService{}
	if _, err := s.SimulateSimple(0.5, 0); err == nil {
		t.Error("n=0 accepted")
	}
	if _, err := s.SimulateSimple(0.5, MaxTrials+1); err == nil {
		t.Error("n too large accepted")
	}
	if _, err := s.SimulateSimple(1.5, 100); err == nil {
		t.Error("p=1.5 accepted")
	}
	if _, err := s.SimulateFullGroup(nil, 100); err == nil {
		t.Error("empty list accepted")
	}
}

func TestNoNaNAtBoundary(t *testing.T) {
	s := &EventService{}
	got, err := s.SimulateFullGroup([]float32{0.5, 0.5}, MinTrials)
	if err != nil {
		t.Fatal(err)
	}
	for i, o := range got.Outcomes {
		if math.IsNaN(o.Frequency) {
			t.Errorf("outcome %d is NaN", i)
		}
	}
}
