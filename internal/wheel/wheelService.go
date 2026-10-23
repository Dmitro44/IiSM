package wheel

import (
	"fmt"
	"math"
	"strings"
)

type WheelService struct {
	games []game
	index map[string]int // название игры -> позиция в games
	total int64          // сумма всех пожертвований в копейках
	wins  map[string]int // сколько раз выпала каждая игра
	spins int
}

func NewWheelService() *WheelService {
	return &WheelService{index: make(map[string]int), wins: make(map[string]int)}
}

// Game — игра на колесе в белорусских рублях.
type Game struct {
	Name        string
	Amount      float64
	Probability float64
	Wins        int
}

// State — состояние колеса для интерфейса.
type State struct {
	Games []Game
	Total float64
	Spins int
}

type Spin struct {
	Winner string
	State  State
}

func (s *WheelService) Donate(name string, amount float64) (State, error) {
	kopecks, err := toKopecks(amount)
	if err != nil {
		return State{}, err
	}
	if err := s.donate(name, kopecks); err != nil {
		return State{}, err
	}

	return s.state(), nil
}

func (s *WheelService) Spin() (Spin, error) {
	if len(s.games) == 0 {
		return Spin{}, ErrNoGames
	}

	winner := s.games[draw(s.games, s.total)]
	s.spins++
	s.wins[winner.name]++

	return Spin{Winner: winner.name, State: s.state()}, nil
}

func (s *WheelService) Snapshot() State {
	return s.state()
}

func (s *WheelService) Reset() State {
	s.games = nil
	s.index = make(map[string]int)
	s.total = 0
	s.wins = make(map[string]int)
	s.spins = 0

	return s.state()
}

func (s *WheelService) donate(name string, amount int64) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrGameName
	}
	if amount <= 0 {
		return fmt.Errorf("%w: получено %d", ErrAmount, amount)
	}

	if i, ok := s.index[name]; ok {
		s.games[i].amount += amount
	} else {
		s.index[name] = len(s.games)
		s.games = append(s.games, game{name: name, amount: amount})
	}
	s.total += amount

	return nil
}

func (s *WheelService) state() State {
	games := make([]Game, len(s.games))
	for i, g := range s.games {
		games[i] = Game{
			Name:        g.name,
			Amount:      float64(g.amount) / 100,
			Probability: share(g.amount, s.total),
			Wins:        s.wins[g.name],
		}
	}

	return State{
		Games: games,
		Total: float64(s.total) / 100,
		Spins: s.spins,
	}
}

func toKopecks(amount float64) (int64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("%w: получено %v", ErrAmount, amount)
	}
	if amount > float64(math.MaxInt64)/100 {
		return 0, fmt.Errorf("%w: сумма не помещается в 64 бита копеек", ErrAmount)
	}

	return int64(math.Round(amount * 100)), nil
}
