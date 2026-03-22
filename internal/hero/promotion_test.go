package hero

import (
	"errors"
	"strings"
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/unitdata"
)

func TestPreserveHPRatioOnPromotion(t *testing.T) {
	if got := preserveHPRatioOnPromotion(5, 10, 20); got != 10 {
		t.Fatalf("5/10 of 20: got %d", got)
	}
	if got := preserveHPRatioOnPromotion(0, 10, 20); got != 0 {
		t.Fatalf("dead: got %d", got)
	}
	if got := preserveHPRatioOnPromotion(10, 10, 12); got != 12 {
		t.Fatalf("full: got %d", got)
	}
}

func TestTryPromoteHero_WarriorRecruitToSquire(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	h.CombatExperience = 7
	h.BasicAttackBonus = 1
	h.RecruitLabel = "Новобранец 1"
	h.CurrentHP = 5
	h.MaxHP = 9

	err = TryPromoteHero(&h)
	if err != nil {
		t.Fatal(err)
	}
	if h.UnitID != unitdata.EmpireWarriorSquire {
		t.Fatalf("UnitID=%q", h.UnitID)
	}
	if h.CombatExperience != 7 || h.BasicAttackBonus != 1 {
		t.Fatalf("progression not kept: xp=%d bonus=%d", h.CombatExperience, h.BasicAttackBonus)
	}
	if h.RecruitLabel != "Новобранец 1" {
		t.Fatalf("label: %q", h.RecruitLabel)
	}
	if h.MaxHP != 12 {
		t.Fatalf("MaxHP=%d", h.MaxHP)
	}
	// (5*12 + 4) / 9 = 7
	if h.CurrentHP != 7 {
		t.Fatalf("CurrentHP=%d want 7", h.CurrentHP)
	}
	if len(h.Abilities) != 2 || h.Abilities[0] != battlepkg.AbilityPowerStrike || h.Abilities[1] != battlepkg.AbilityBasicAttack {
		t.Fatalf("abilities: %+v", h.Abilities)
	}
}

func TestTryPromoteHero_LegacyNoUnitID(t *testing.T) {
	h := recruitHeroFallbackNoTemplate()
	err := TryPromoteHero(&h)
	if !errors.Is(err, ErrPromotionNoUnitID) {
		t.Fatalf("got %v", err)
	}
}

func TestTryPromoteHero_ThirdPromotionNoPath(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); !errors.Is(err, ErrPromotionBranchChoiceRequired) {
		t.Fatalf("squire has two branches: %v", err)
	}
	if err := TryPromoteHeroTo(&h, unitdata.EmpireWarriorDD1); err != nil {
		t.Fatal(err)
	}
	if h.UnitID != unitdata.EmpireWarriorDD1 {
		t.Fatalf("want tier3 warrior, got %q", h.UnitID)
	}
	err = TryPromoteHero(&h)
	if !errors.Is(err, ErrPromotionNoPath) {
		t.Fatalf("promote after tier3: %v", err)
	}
}

func TestPromotionUILine_campAndDomain(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	sOff := PromotionUILine(&h, false)
	if sOff == "" || !strings.Contains(sOff, "лагер") {
		t.Fatalf("off camp: %q", sOff)
	}
	sOn := PromotionUILine(&h, true)
	if sOn == "" || !strings.Contains(sOn, "P") {
		t.Fatalf("on camp: %q", sOn)
	}
}

func TestCombatUnitSeed_AfterPromotion(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireHealerNovice)
	if err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	s := h.CombatUnitSeed()
	if s.Def.TemplateUnitID != unitdata.EmpireHealerAcolyte {
		t.Fatalf("seed TemplateUnitID=%q", s.Def.TemplateUnitID)
	}
	if s.Def.Tier != 2 {
		t.Fatalf("tier %d", s.Def.Tier)
	}
}

func TestTryPromoteHero_twoBranchesRequiresTo(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); !errors.Is(err, ErrPromotionBranchChoiceRequired) {
		t.Fatalf("expected branch choice: %v", err)
	}
}

func TestTryPromoteHeroTo_tankBranch(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHeroTo(&h, unitdata.EmpireWarriorTank1); err != nil {
		t.Fatal(err)
	}
	if h.UnitID != unitdata.EmpireWarriorTank1 {
		t.Fatalf("got %q", h.UnitID)
	}
	s := h.CombatUnitSeed()
	if s.Def.TemplateUnitID != unitdata.EmpireWarriorTank1 {
		t.Fatalf("seed TemplateUnitID=%q", s.Def.TemplateUnitID)
	}
}

func TestTryPromoteHeroTo_rejectsWrongTarget(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHeroTo(&h, unitdata.EmpireArcherPure1); err == nil || !errors.Is(err, ErrPromotionTargetNotAllowed) {
		t.Fatalf("expected ErrPromotionTargetNotAllowed, got %v", err)
	}
}

func TestCombatUnitSeed_AfterTier3Promotion(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireHealerNovice)
	if err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHero(&h); err != nil {
		t.Fatal(err)
	}
	if err := TryPromoteHeroTo(&h, unitdata.EmpireHealerSingle1); err != nil {
		t.Fatal(err)
	}
	s := h.CombatUnitSeed()
	if s.Def.TemplateUnitID != unitdata.EmpireHealerSingle1 {
		t.Fatalf("seed TemplateUnitID=%q", s.Def.TemplateUnitID)
	}
	if s.Def.Tier != 3 {
		t.Fatalf("tier %d", s.Def.Tier)
	}
}
