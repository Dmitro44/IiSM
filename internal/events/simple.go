package events

import (
	"math/rand/v2"
)

func Simple(p float32) bool {
	n := rand.Float32()

	return n < p
}
