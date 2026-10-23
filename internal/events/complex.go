package events

import "math/rand/v2"

const MaxComplexEvents = 10

func Complex(pList []float32) []bool {
	result := make([]bool, len(pList))
	for i, p := range pList {
		if rand.Float32() < p {
			result[i] = true
		}
	}

	return result
}

func jointProbability(pList []float32, mask int) float64 {
	probability := 1.0
	for i, p := range pList {
		if mask&(1<<i) != 0 {
			probability *= float64(p)
		} else {
			probability *= 1 - float64(p)
		}
	}

	return probability
}
