package battle

import "testing"

func TestBattleContext_PhaseLabelRU(t *testing.T) {
	c := &BattleContext{Phase: PhaseAwaitAction}
	if got := c.PhaseLabelRU(); got != "действие" {
		t.Fatalf("got %q", got)
	}
}

func TestPlayerTurnState_PhaseLabelRU(t *testing.T) {
	p := PlayerTurnState{Phase: PlayerChooseAbility}
	if got := p.PhaseLabelRU(); got != "способность" {
		t.Fatalf("got %q", got)
	}
}
