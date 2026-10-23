package wheel

import (
	"errors"

	"github.com/Dmitro44/IiSM/internal/events"
)

var (
	ErrGameName = errors.New("название игры не должно быть пустым")
	ErrAmount   = errors.New("сумма пожертвования должна быть положительной")
	ErrNoGames  = errors.New("колесо пусто: ни одного пожертвования")
)

type game struct {
	name   string
	amount int64
}

func draw(games []game, total int64) int {
	probabilities := make([]float32, len(games))
	for i, g := range games {
		probabilities[i] = float32(share(g.amount, total))
	}

	return events.FullGroup(probabilities)
}

func share(amount, total int64) float64 {
	if total == 0 {
		return 0
	}

	return float64(amount) / float64(total)
}
