package game

import (
	"errors"
	"testing"

	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

func TestEvaluatePromotionGate_notAtCampBlocks(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	g := EvaluatePromotionGate(&h, false, 99, "")
	if g.Allowed {
		t.Fatal("expected block off camp")
	}
	if g.Message == "" {
		t.Fatal("expected message")
	}
	if g.Cost != 2 {
		t.Fatalf("tier2 target cost: got Cost=%d", g.Cost)
	}
}

func TestEvaluatePromotionGate_atCampAllowsDomain(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	g := EvaluatePromotionGate(&h, true, 2, "")
	if !g.Allowed {
		t.Fatalf("expected allow, got %q", g.Message)
	}
	if g.Cost != 2 {
		t.Fatalf("Cost=%d want 2 (tier 2 target)", g.Cost)
	}
}

func TestEvaluatePromotionGate_insufficientMarks(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	g := EvaluatePromotionGate(&h, true, 1, "")
	if g.Allowed {
		t.Fatal("expected block: not enough marks")
	}
	if g.Message == "" {
		t.Fatal("expected message")
	}
	if g.Cost != 2 {
		t.Fatalf("Cost=%d", g.Cost)
	}
}

func TestEvaluatePromotionGate_noPathEvenAtCamp(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := hero.TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	if err := hero.TryPromoteHeroTo(&h, unitdata.EmpireWarriorDD1); err != nil {
		t.Fatal(err)
	}
	g := EvaluatePromotionGate(&h, true, 99, "")
	if g.Allowed {
		t.Fatal("tier3 should have no upgrade path")
	}
	if g.Message == "" {
		t.Fatal("expected message")
	}
}

func TestEvaluatePromotionGate_legacyNoUnitID(t *testing.T) {
	h := hero.Hero{}
	g := EvaluatePromotionGate(&h, true, 99, "")
	if g.Allowed {
		t.Fatal("legacy")
	}
	if !errors.Is(hero.ValidatePromotionDomain(&h), hero.ErrPromotionNoUnitID) {
		t.Fatal("domain")
	}
}

func TestEvaluatePromotionGate_twoBranchesRequiresSelection(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := hero.TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	g := EvaluatePromotionGate(&h, true, 99, "")
	if g.Allowed {
		t.Fatal("must choose branch first")
	}
	if g.Cost != 0 {
		t.Fatalf("Cost=%d want 0 until branch chosen", g.Cost)
	}
	g2 := EvaluatePromotionGate(&h, true, 99, unitdata.EmpireWarriorTank1)
	if !g2.Allowed {
		t.Fatalf("expected allow with branch: %q", g2.Message)
	}
	if g2.Cost != 3 {
		t.Fatalf("tier3 cost: %d", g2.Cost)
	}
}
