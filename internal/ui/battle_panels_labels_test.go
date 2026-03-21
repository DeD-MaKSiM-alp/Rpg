package ui

import (
	"fmt"
	"strings"
	"testing"

	battlepkg "mygame/internal/battle"
)

func TestBattleActionTargetLabelRU(t *testing.T) {
	b := &battlepkg.BattleContext{Units: map[battlepkg.UnitID]*battlepkg.BattleUnit{}}
	pt := &battlepkg.PlayerTurnState{SelectedTarget: battlepkg.SelfTarget()}
	if got := battleActionTargetLabelRU(pt, b); got != "себя" {
		t.Fatalf("self: got %q", got)
	}
	pt2 := &battlepkg.PlayerTurnState{SelectedTarget: battlepkg.NoTarget()}
	if got := battleActionTargetLabelRU(pt2, b); got != "нет" {
		t.Fatalf("none: got %q", got)
	}
}

func TestPlayerTurn_summaryStepLineRussian(t *testing.T) {
	p := battlepkg.PlayerTurnState{Phase: battlepkg.PlayerChooseTarget}
	line := fmt.Sprintf("Шаг: %s", p.PhaseLabelRU())
	if strings.Contains(line, "Step") || strings.Contains(line, "Choose") {
		t.Fatalf("unexpected English in %q", line)
	}
}
