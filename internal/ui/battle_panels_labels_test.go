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
	if got := battleActionTargetLabelRU(pt, b, TierMedium); got != "себя" {
		t.Fatalf("self: got %q", got)
	}
	pt2 := &battlepkg.PlayerTurnState{SelectedTarget: battlepkg.NoTarget()}
	if got := battleActionTargetLabelRU(pt2, b, TierMedium); got != "нет" {
		t.Fatalf("none: got %q", got)
	}
}

// Контракт dedupe: на small tier подпись цели для юнита — только имя (без «(#id)»), т.к. колонка цели уже показывает имя.
func TestBattleActionTargetLabelRU_unitSmallOmitsIDSuffix(t *testing.T) {
	u := &battlepkg.CombatUnit{
		ID: 7,
		Def: battlepkg.CombatUnitDefinition{
			DisplayName: "Герой",
			Base:        battlepkg.UnitBaseStats{MaxHP: 10},
		},
		State: battlepkg.CombatUnitState{HP: 10, Alive: true},
	}
	b := &battlepkg.BattleContext{Units: map[battlepkg.UnitID]*battlepkg.CombatUnit{7: u}}
	pt := &battlepkg.PlayerTurnState{SelectedTarget: battlepkg.UnitTarget(7)}
	if med := battleActionTargetLabelRU(pt, b, TierMedium); !strings.Contains(med, "(#7)") {
		t.Fatalf("medium label should include id: %q", med)
	}
	if sm := battleActionTargetLabelRU(pt, b, TierSmall); sm != "Герой" || strings.Contains(sm, "#") {
		t.Fatalf("small label should be name only: %q", sm)
	}
}

func TestPlayerTurn_summaryStepLineRussian(t *testing.T) {
	p := battlepkg.PlayerTurnState{Phase: battlepkg.PlayerChooseTarget}
	line := fmt.Sprintf("Шаг: %s", p.PhaseLabelRU())
	if strings.Contains(line, "Step") || strings.Contains(line, "Choose") {
		t.Fatalf("unexpected English in %q", line)
	}
}
