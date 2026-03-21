package game

import (
	"testing"

	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

func TestPromotionCostFromTargetTier(t *testing.T) {
	if got := PromotionCostFromTargetTier(2); got != 2 {
		t.Fatalf("tier 2: %d", got)
	}
	if got := PromotionCostFromTargetTier(3); got != 3 {
		t.Fatalf("tier 3: %d", got)
	}
	if PromotionCostFromTargetTier(2) >= PromotionCostFromTargetTier(3) {
		t.Fatal("tier 3 should cost more than tier 2")
	}
	if got := PromotionCostFromTargetTier(0); got != 1 {
		t.Fatalf("clamp tier 0: %d", got)
	}
}

func TestPromotionTrainingMarkCostForHero_tier1ToTier2(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := PromotionTrainingMarkCostForHero(&h)
	if !ok || c != 2 {
		t.Fatalf("want cost 2 (target tier 2), got %d ok=%v", c, ok)
	}
}

func TestPromotionTrainingMarkCostForHero_noPathOrLegacy(t *testing.T) {
	h := hero.Hero{}
	if _, ok := PromotionTrainingMarkCostForHero(&h); ok {
		t.Fatal("legacy should not have cost")
	}
	h2, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := hero.TryPromoteHero(&h2); err != nil {
		t.Fatal(err)
	}
	if err := hero.TryPromoteHeroTo(&h2, unitdata.EmpireWarriorDD1); err != nil {
		t.Fatal(err)
	}
	if _, ok := PromotionTrainingMarkCostForHero(&h2); ok {
		t.Fatal("no next step — no cost")
	}
}

func TestPromotionTrainingMarkCostForHero_tier3Target(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := hero.TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	c, ok := PromotionTrainingMarkCostForHeroTarget(&h, unitdata.EmpireWarriorDD1)
	if !ok || c != 3 {
		t.Fatalf("want cost 3 (target tier 3), got %d ok=%v", c, ok)
	}
}

func TestEvaluatePromotionGate_usesTierCostInMessage(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	g := EvaluatePromotionGate(&h, true, 1, "")
	if g.Allowed {
		t.Fatal("expected insufficient")
	}
	// Сообщение должно содержать требуемую цену (2 для tier-2 цели).
	if g.Cost != 2 {
		t.Fatalf("gate.Cost=%d", g.Cost)
	}
}
