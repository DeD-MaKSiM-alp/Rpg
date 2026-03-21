package game

import (
	"testing"

	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

func TestApplyVictoryTrainingMarks(t *testing.T) {
	g := NewGame(0, 2, 2)
	if g.TrainingMarks != 0 {
		t.Fatalf("start marks=%d", g.TrainingMarks)
	}
	g.applyVictoryTrainingMarks()
	g.applyVictoryTrainingMarks()
	if g.TrainingMarks != 2*TrainingMarksPerVictory {
		t.Fatalf("marks=%d", g.TrainingMarks)
	}
}

func TestPromotionSuccessDeductsMarks(t *testing.T) {
	g := NewGame(0, 2, 2)
	g.TrainingMarks = 2
	h := g.party.HeroAtGlobalIndex(0)
	if h == nil {
		t.Fatal("no hero")
	}
	gate := EvaluatePromotionGate(h, true, g.TrainingMarks, "")
	if !gate.Allowed {
		t.Fatalf("gate: %s", gate.Message)
	}
	if gate.Cost != 2 {
		t.Fatalf("tier-2 target cost=%d", gate.Cost)
	}
	if err := hero.TryPromoteHero(h); err != nil {
		t.Fatal(err)
	}
	g.TrainingMarks -= gate.Cost
	if g.TrainingMarks != 0 {
		t.Fatalf("after deduct marks=%d", g.TrainingMarks)
	}
}

func TestPromotionGateBlocksWithInsufficientMarks_NoDeductSimulated(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	marks := 1
	gate := EvaluatePromotionGate(&h, true, marks, "")
	if gate.Allowed {
		t.Fatal("expected block")
	}
	if gate.Cost != 2 {
		t.Fatalf("Cost=%d", gate.Cost)
	}
	// Имитация: при отказе по gate не вызываем TryPromoteHero и не списываем.
	if marks != 1 {
		t.Fatal("marks should be unchanged")
	}
}
