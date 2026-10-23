package events

import "math/rand/v2"

func Dependent(pA, pBA float32) int {
	aHappened := rand.Float32() < pA

	pB := pBA

	// P(B|¬A) = 1 - P(B|A)
	if !aHappened {
		pB = 1 - pBA
	}
	bHappened := rand.Float32() < pB

	switch {
	case aHappened && bHappened:
		return 0
	case aHappened:
		return 1
	case bHappened:
		return 2
	default:
		return 3
	}
}
