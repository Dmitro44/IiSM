package wheel

import (
	"math"
	"testing"
)

const tol = 0.01

// donate вносит пожертвование и останавливает тест при ошибке.
func donate(t *testing.T, s *WheelService, name string, amount float64) {
	t.Helper()
	if _, err := s.Donate(name, amount); err != nil {
		t.Fatal(err)
	}
}

func TestDonationsAccumulate(t *testing.T) {
	s := NewWheelService()
	donate(t, s, "Witcher 3", 250)
	donate(t, s, "Cyberpunk 2077", 100)
	donate(t, s, "Witcher 3", 50)

	state := s.Snapshot()
	if len(state.Games) != 2 {
		t.Fatalf("games = %d, want 2", len(state.Games))
	}
	if state.Games[0].Name != "Witcher 3" || state.Games[0].Amount != 300 {
		t.Errorf("first game = %+v, want Witcher 3 / 300", state.Games[0])
	}
	if state.Games[1].Name != "Cyberpunk 2077" || state.Games[1].Amount != 100 {
		t.Errorf("second game = %+v, want Cyberpunk 2077 / 100", state.Games[1])
	}
	if state.Total != 400 {
		t.Errorf("total = %v, want 400", state.Total)
	}
}

func TestSharesAreDonationFractions(t *testing.T) {
	s := NewWheelService()
	for _, d := range []struct {
		name   string
		amount float64
	}{
		{"Witcher 3", 250},
		{"Cyberpunk 2077", 100},
		{"Hades", 50},
	} {
		donate(t, s, d.name, d.amount)
	}

	want := []float64{0.625, 0.25, 0.125}
	sum := 0.0
	for i, game := range s.Snapshot().Games {
		if math.Abs(game.Probability-want[i]) > 1e-9 {
			t.Errorf("%s share = %v, want %v", game.Name, game.Probability, want[i])
		}
		sum += game.Probability
	}
	if math.Abs(sum-1) > 1e-9 {
		t.Errorf("shares sum to %v, want 1", sum)
	}
}

func TestSpinFollowsDonationShares(t *testing.T) {
	s := NewWheelService()
	for _, d := range []struct {
		name   string
		amount float64
	}{
		{"Witcher 3", 300},
		{"Cyberpunk 2077", 100},
		{"Hades", 100},
	} {
		donate(t, s, d.name, d.amount)
	}

	const n = 200000
	counts := make(map[string]int)
	for range n {
		result, err := s.Spin()
		if err != nil {
			t.Fatal(err)
		}
		counts[result.Winner]++
	}

	total := 0
	for _, game := range s.Snapshot().Games {
		frequency := float64(counts[game.Name]) / n
		if math.Abs(frequency-game.Probability) > tol {
			t.Errorf("%s empirical=%v theory=%v", game.Name, frequency, game.Probability)
		}
		total += counts[game.Name]
	}
	if total != n {
		t.Errorf("counts sum to %d, want %d", total, n)
	}
}

// Пожертвования, пришедшие уже во время розыгрыша, меняют вероятности:
// победитель определяется долей на момент вращения.
func TestSpinUsesDonationsAtSpinTime(t *testing.T) {
	s := NewWheelService()
	donate(t, s, "Witcher 3", 999)
	donate(t, s, "Hades", 1)

	result, err := s.Spin()
	if err != nil {
		t.Fatal(err)
	}
	if result.Winner != "Witcher 3" {
		t.Errorf("winner = %s, want Witcher 3", result.Winner)
	}

	donate(t, s, "Hades", 100000)
	games := s.Snapshot().Games
	if games[1].Probability <= games[0].Probability {
		t.Errorf("Hades share %v is not the largest after the new donation", games[1].Probability)
	}
	if math.Abs(games[1].Probability-100000.0/100999.0) > 1e-6 {
		t.Errorf("Hades share = %v, want %v", games[1].Probability, 100000.0/100999.0)
	}
}

func TestRejectsBadDonation(t *testing.T) {
	s := NewWheelService()

	if _, err := s.Donate("  ", 100); err == nil {
		t.Error("blank name accepted")
	}
	if _, err := s.Donate("Hades", 0); err == nil {
		t.Error("zero amount accepted")
	}
	if _, err := s.Donate("Hades", -100); err == nil {
		t.Error("negative amount accepted")
	}
	if state := s.Snapshot(); len(state.Games) != 0 || state.Total != 0 {
		t.Errorf("rejected donations changed the wheel: %+v", state)
	}
}

func TestTrimsDonationName(t *testing.T) {
	s := NewWheelService()
	for _, name := range []string{"Witcher 3", " Witcher 3 "} {
		donate(t, s, name, 100)
	}

	state := s.Snapshot()
	if len(state.Games) != 1 {
		t.Fatalf("games = %d, want 1", len(state.Games))
	}
	if state.Games[0].Name != "Witcher 3" {
		t.Errorf("name = %q, want %q", state.Games[0].Name, "Witcher 3")
	}
}

func TestSpinOnEmptyWheelFails(t *testing.T) {
	if _, err := NewWheelService().Spin(); err == nil {
		t.Error("empty wheel produced a winner")
	}
}

func TestDonateConvertsRublesToKopecks(t *testing.T) {
	s := NewWheelService()

	state, err := s.Donate("Witcher 3", 250)
	if err != nil {
		t.Fatal(err)
	}
	if state.Total != 250 {
		t.Errorf("total = %v, want 250", state.Total)
	}
	if len(state.Games) != 1 || state.Games[0].Amount != 250 {
		t.Fatalf("games = %+v, want one game with 250", state.Games)
	}

	// Дробная часть до копейки доезжает через округление, а не теряется.
	state, err = s.Donate("Witcher 3", 0.005)
	if err != nil {
		t.Fatal(err)
	}
	if state.Total != 250.01 {
		t.Errorf("total = %v, want 250.01", state.Total)
	}
}

func TestSpinCountsWinsAndFollowsShares(t *testing.T) {
	s := NewWheelService()
	donate(t, s, "Witcher 3", 750)
	donate(t, s, "Hades", 250)

	const n = 20000
	for range n {
		result, err := s.Spin()
		if err != nil {
			t.Fatal(err)
		}
		if result.Winner != "Witcher 3" && result.Winner != "Hades" {
			t.Fatalf("winner = %q, want one of the donated games", result.Winner)
		}
		if result.State.Spins < 1 || result.State.Spins > n {
			t.Fatalf("spin counter = %d, want 1..%d", result.State.Spins, n)
		}
	}

	state := s.Snapshot()
	if state.Spins != n {
		t.Errorf("spins = %d, want %d", state.Spins, n)
	}

	wins := 0
	for _, game := range state.Games {
		wins += game.Wins
		frequency := float64(game.Wins) / n
		if math.Abs(frequency-game.Probability) > 0.02 {
			t.Errorf("%s empirical=%v theory=%v", game.Name, frequency, game.Probability)
		}
	}
	if wins != n {
		t.Errorf("wins sum to %d, want %d", wins, n)
	}
}

func TestRejectsBadAmount(t *testing.T) {
	s := NewWheelService()

	for _, amount := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, err := s.Donate("Hades", amount); err == nil {
			t.Errorf("amount %v accepted", amount)
		}
	}
	if state := s.Snapshot(); len(state.Games) != 0 || state.Total != 0 {
		t.Errorf("rejected donations changed the wheel: %+v", state)
	}
}

func TestResetClearsWheelAndStats(t *testing.T) {
	s := NewWheelService()
	donate(t, s, "Witcher 3", 250)
	if _, err := s.Spin(); err != nil {
		t.Fatal(err)
	}

	state := s.Reset()
	if len(state.Games) != 0 || state.Total != 0 || state.Spins != 0 {
		t.Errorf("state after reset = %+v, want empty", state)
	}
	if _, err := s.Spin(); err == nil {
		t.Error("spin succeeded on the reset wheel")
	}
}
