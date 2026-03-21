package hero

import (
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/unitdata"
)

func TestNewHeroFromUnitTemplate_setsUnitIDAndStats(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireMilitiaSpearmanT1)
	if err != nil {
		t.Fatal(err)
	}
	if h.UnitID != unitdata.EmpireMilitiaSpearmanT1 {
		t.Fatalf("UnitID: got %q", h.UnitID)
	}
	if h.MaxHP != 10 || h.CurrentHP != 10 || h.Attack != 2 {
		t.Fatalf("unexpected stats: %+v", h)
	}
}

func TestNewHeroFromUnitTemplate_unknown(t *testing.T) {
	_, err := NewHeroFromUnitTemplate("___missing___")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRecruitHeroFromEarlyPool_rotatesIDs(t *testing.T) {
	a := RecruitHeroFromEarlyPool(1)
	b := RecruitHeroFromEarlyPool(2)
	c := RecruitHeroFromEarlyPool(5) // (5-1)%3 == (2-1)%3
	if a.UnitID == b.UnitID {
		t.Fatalf("expected different templates in pool step 1 vs 2, got %q", a.UnitID)
	}
	if b.UnitID != c.UnitID {
		t.Fatalf("pool index 2 and 5 should match for len=3, got %q vs %q", b.UnitID, c.UnitID)
	}
}

func TestCombatUnitSeed_identityFromTemplate(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireArcherRecruit)
	if err != nil {
		t.Fatal(err)
	}
	s := h.CombatUnitSeed()
	if s.Def.Base.MaxHP != h.MaxHP {
		t.Fatalf("seed hp mismatch")
	}
	if s.Def.Role != battlepkg.RoleArcher || !s.Def.IsRanged {
		t.Fatalf("archer seed: role=%v IsRanged=%v", s.Def.Role, s.Def.IsRanged)
	}
	if s.Def.TemplateUnitID != unitdata.EmpireArcherRecruit {
		t.Fatalf("TemplateUnitID=%q", s.Def.TemplateUnitID)
	}
	if s.Def.ArchetypeID != "ranged_generalist" {
		t.Fatalf("ArchetypeID=%q", s.Def.ArchetypeID)
	}
	if s.Def.FactionID != "empire" || s.Def.LineID != "archer" || s.Def.Tier != 1 {
		t.Fatalf("identity: %+v", s.Def)
	}
	if s.Def.IdentityAttackKind != battlepkg.TemplateAttackRanged {
		t.Fatalf("IdentityAttackKind=%v", s.Def.IdentityAttackKind)
	}
}

func TestCombatUnitSeed_groupHealerBranch(t *testing.T) {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireHealerGroup1)
	if err != nil {
		t.Fatal(err)
	}
	s := h.CombatUnitSeed()
	if len(s.Def.Loadout.Abilities) < 1 || s.Def.Loadout.Abilities[0] != battlepkg.AbilityGroupHeal {
		t.Fatalf("loadout: %+v", s.Def.Loadout.Abilities)
	}
	if s.Def.TemplateUnitID != unitdata.EmpireHealerGroup1 {
		t.Fatalf("TemplateUnitID=%q", s.Def.TemplateUnitID)
	}
}

func TestCombatUnitSeed_legacyEmptyUnitIDUsesAbilities(t *testing.T) {
	h := Hero{
		UnitID:    "",
		MaxHP:     10,
		CurrentHP: 10,
		Attack:    1,
		Abilities: []battlepkg.AbilityID{battlepkg.AbilityHeal, battlepkg.AbilityBasicAttack},
	}
	s := h.CombatUnitSeed()
	if s.Def.TemplateUnitID != "" {
		t.Fatalf("expected empty TemplateUnitID, got %q", s.Def.TemplateUnitID)
	}
	if s.Def.Role != battlepkg.RoleHealer || s.Def.IsRanged {
		t.Fatalf("legacy healer: role=%v IsRanged=%v", s.Def.Role, s.Def.IsRanged)
	}
	if s.Def.IdentityAttackKind != battlepkg.TemplateAttackUnknown {
		t.Fatalf("IdentityAttackKind=%v", s.Def.IdentityAttackKind)
	}
}

func TestRecruitFallback_noUnitID(t *testing.T) {
	h := recruitHeroFallbackNoTemplate()
	if h.UnitID != "" {
		t.Fatalf("fallback should leave UnitID empty, got %q", h.UnitID)
	}
	if len(h.Abilities) == 0 || h.Abilities[0] != battlepkg.AbilityBasicAttack {
		t.Fatalf("fallback abilities: %+v", h.Abilities)
	}
}
