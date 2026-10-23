package events

import "math/rand/v2"

func FullGroup(probabilities []float32) int {
	u := rand.Float32()

	accumulated := float32(0)
	for i, p := range probabilities {
		accumulated += p
		if u < accumulated {
			return i
		}
	}

	// сумма вероятностей меньше 1.0
	return len(probabilities) - 1
}
