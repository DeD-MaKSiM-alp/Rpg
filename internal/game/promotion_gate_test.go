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
	g := EvaluatePromotionGate(&h, false)
	if g.Allowed {
		t.Fatal("expected block off camp")
	}
	if g.Message == "" {
		t.Fatal("expected message")
	}
}

func TestEvaluatePromotionGate_atCampAllowsDomain(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	g := EvaluatePromotionGate(&h, true)
	if !g.Allowed {
		t.Fatalf("expected allow, got %q", g.Message)
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
	g := EvaluatePromotionGate(&h, true)
	if g.Allowed {
		t.Fatal("tier2 should have no upgrade path")
	}
	if g.Message == "" {
		t.Fatal("expected message")
	}
}

func TestEvaluatePromotionGate_legacyNoUnitID(t *testing.T) {
	h := hero.Hero{}
	g := EvaluatePromotionGate(&h, true)
	if g.Allowed {
		t.Fatal("legacy")
	}
	if !errors.Is(hero.ValidatePromotionDomain(&h), hero.ErrPromotionNoUnitID) {
		t.Fatal("domain")
	}
}
