package ui

import (
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

func TestInspectRoleIconFromUnitTemplate_healerRanged(t *testing.T) {
	tpl, ok := unitdata.GetUnitTemplate(unitdata.EmpireHealerAcolyte)
	if !ok {
		t.Fatal("template")
	}
	if got := InspectRoleIconFromUnitTemplate(&tpl); got != InspectRoleIconHeal {
		t.Fatalf("healer want Heal, got %v", got)
	}
	tpl2, _ := unitdata.GetUnitTemplate(unitdata.EmpireArcherRecruit)
	if InspectRoleIconFromUnitTemplate(&tpl2) != InspectRoleIconRanged {
		t.Fatalf("archer want Ranged")
	}
	tpl3, _ := unitdata.GetUnitTemplate(unitdata.EmpireWarriorRecruit)
	if InspectRoleIconFromUnitTemplate(&tpl3) != InspectRoleIconMelee {
		t.Fatalf("warrior want Melee")
	}
}

func TestInspectRoleIconFromCombatUnit_legacyRanged(t *testing.T) {
	u := &battlepkg.CombatUnit{
		Def: battlepkg.CombatUnitDefinition{
			Role:     battlepkg.RoleFighter,
			IsRanged: true,
		},
	}
	if InspectRoleIconFromCombatUnit(u) != InspectRoleIconRanged {
		t.Fatal("expected ranged from IsRanged")
	}
}

func TestInspectRoleIconFromHero_matchesTemplate(t *testing.T) {
	h := hero.DefaultHero()
	if InspectRoleIconFromHero(&h) == InspectRoleIconUnknown {
		t.Fatal("default hero should resolve icon from template")
	}
}
