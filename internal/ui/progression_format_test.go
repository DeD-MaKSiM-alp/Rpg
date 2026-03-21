package ui

import (
	"testing"

	"mygame/internal/hero"
)

func TestCombatXPToNextBonusStep(t *testing.T) {
	steps := hero.CombatXPStepsPerBasicAttackBonus
	if CombatXPToNextBonusStep(0) != steps {
		t.Fatalf("from 0 want %d, got %d", steps, CombatXPToNextBonusStep(0))
	}
	if CombatXPToNextBonusStep(3) != 1 {
		t.Fatalf("from 3 want 1, got %d", CombatXPToNextBonusStep(3))
	}
	if CombatXPToNextBonusStep(4) != steps {
		t.Fatalf("from 4 want %d, got %d", steps, CombatXPToNextBonusStep(4))
	}
}
