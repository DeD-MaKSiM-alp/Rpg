package ui

import (
	"strings"
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
)

func TestInspectAndBattleUseSameAbilityLabels(t *testing.T) {
	h := hero.DefaultHero()
	m := buildFormationInspectCardModel(&h, 0, 1, false, false, 0, nil, nil, 0, "", "")
	flat := FlattenInspectCardText(m)
	label := battlepkg.PlayerAbilityLabelRU(battlepkg.AbilityBasicAttack)
	if !strings.Contains(flat, label) {
		t.Fatalf("inspect card should contain %q, got snippet of flat", label)
	}
	if label != "Базовый удар" {
		t.Fatal("canonical basic attack label")
	}
}
