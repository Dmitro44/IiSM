package events

import (
	"errors"
	"fmt"
)

const (
	MinTrials = 1
	MaxTrials = 1_000_000
)

var (
	ErrTrials        = errors.New("n должно быть от 1 до 1000000")
	ErrProbability   = errors.New("вероятность должна быть в интервале [0, 1]")
	ErrNoEvents      = errors.New("нужен хотя бы один список вероятностей")
	ErrTooManyEvents = errors.New("событий должно быть не больше 8: исходов 2^k")
)

type EventService struct{}

type Outcome struct {
	Count       int
	Frequency   float64
	Probability float64
}

type Result struct {
	SampleSize int
	Outcomes   []Outcome
}

func validate(n int, probabilities ...float32) error {
	if n < MinTrials || n > MaxTrials {
		return fmt.Errorf("%w: получено %d", ErrTrials, n)
	}
	if len(probabilities) == 0 {
		return ErrNoEvents
	}
	for _, p := range probabilities {
		if p < 0 || p > 1 {
			return fmt.Errorf("%w: получено %v", ErrProbability, p)
		}
	}
	return nil
}

func newOutcome(count, n int, probability float64) Outcome {
	return Outcome{
		Count:       count,
		Frequency:   float64(count) / float64(n),
		Probability: probability,
	}
}

func (e *EventService) SimulateSimple(p float32, n int) (Result, error) {
	if err := validate(n, p); err != nil {
		return Result{}, err
	}

	count := 0
	for range n {
		if Simple(p) {
			count++
		}
	}

	return Result{
		SampleSize: n,
		Outcomes: []Outcome{
			newOutcome(count, n, float64(p)),
			newOutcome(n-count, n, 1-float64(p)),
		},
	}, nil
}

func (e *EventService) SimulateComplex(pList []float32, n int) (Result, error) {
	if err := validate(n, pList...); err != nil {
		return Result{}, err
	}
	if len(pList) > MaxComplexEvents {
		return Result{}, fmt.Errorf("%w: получено %d, максимум %d",
			ErrTooManyEvents, len(pList), MaxComplexEvents)
	}

	counts := make([]int, 1<<len(pList))
	for range n {
		mask := 0
		for i, occurred := range Complex(pList) {
			if occurred {
				mask |= 1 << i
			}
		}
		counts[mask]++
	}

	outcomes := make([]Outcome, len(counts))
	for mask, count := range counts {
		outcomes[mask] = newOutcome(count, n, jointProbability(pList, mask))
	}

	return Result{SampleSize: n, Outcomes: outcomes}, nil
}

func (e *EventService) SimulateDependent(pA, pBA float32, n int) (Result, error) {
	if err := validate(n, pA, pBA); err != nil {
		return Result{}, err
	}

	pNotA := 1 - float64(pA)
	pNotBA := 1 - float64(pBA) // P(¬B|A)

	pBNotA := 1 - float64(pBA) // P(B|Ā)  = 1 − P(B|A) по заданию.

	counts := [4]int{}
	for range n {
		counts[Dependent(pA, pBA)]++
	}

	return Result{
		SampleSize: n,
		Outcomes: []Outcome{
			newOutcome(counts[0], n, float64(pA)*float64(pBA)),
			newOutcome(counts[1], n, float64(pA)*pNotBA),
			newOutcome(counts[2], n, pNotA*pBNotA),
			newOutcome(counts[3], n, pNotA*(1-pBNotA)),
		},
	}, nil
}

func (e *EventService) SimulateFullGroup(probabilities []float32, n int) (Result, error) {
	if err := validate(n, probabilities...); err != nil {
		return Result{}, err
	}

	counts := make([]int, len(probabilities))
	for range n {
		counts[FullGroup(probabilities)]++
	}

	outcomes := make([]Outcome, len(probabilities))
	for i, p := range probabilities {
		outcomes[i] = newOutcome(counts[i], n, float64(p))
	}

	return Result{SampleSize: n, Outcomes: outcomes}, nil
}
