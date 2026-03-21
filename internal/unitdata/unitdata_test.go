package unitdata

import (
	"errors"
	"testing"

	battlepkg "mygame/internal/battle"
)

func TestGetUnitTemplate_known(t *testing.T) {
	tpl, ok := GetUnitTemplate(EmpireWarriorRecruit)
	if !ok || tpl.UnitID != EmpireWarriorRecruit {
		t.Fatalf("expected warrior recruit, got ok=%v tpl=%+v", ok, tpl)
	}
}

func TestGetUnitTemplate_emptyAndUnknown(t *testing.T) {
	if _, ok := GetUnitTemplate(""); ok {
		t.Fatal("empty id should miss")
	}
	if _, ok := GetUnitTemplate("no_such_unit"); ok {
		t.Fatal("unknown id should miss")
	}
}

func TestMustGetUnitTemplate_error(t *testing.T) {
	_, err := MustGetUnitTemplate("bad")
	var eu ErrUnknownUnit
	if !errors.As(err, &eu) {
		t.Fatalf("expected ErrUnknownUnit, got %v", err)
	}
}

func TestTier2Templates_registered(t *testing.T) {
	for _, id := range []string{EmpireWarriorSquire, EmpireArcherMarksmanBase, EmpireHealerAcolyte} {
		if _, ok := GetUnitTemplate(id); !ok {
			t.Fatalf("missing %q", id)
		}
	}
}

func TestTier3Templates_registeredAndUpgradePath(t *testing.T) {
	for _, id := range []string{EmpireWarriorDD1, EmpireWarriorTank1, EmpireArcherPure1, EmpireHealerSingle1, EmpireHealerGroup1} {
		tpl, ok := GetUnitTemplate(id)
		if !ok || tpl.Tier != 3 {
			t.Fatalf("tier3 %q: ok=%v tier=%d", id, ok, tpl.Tier)
		}
		if tpl.UpgradeToUnitID != "" {
			t.Fatalf("%q: tier3 top should have empty upgrade", id)
		}
	}
	ws, ok := GetUnitTemplate(EmpireWarriorSquire)
	if !ok || len(ws.UpgradeOptions) != 2 || ws.UpgradeOptions[0] != EmpireWarriorTank1 || ws.UpgradeOptions[1] != EmpireWarriorDD1 {
		t.Fatalf("warrior squire branches: %+v", ws)
	}
	if ws.UpgradeToUnitID != "" {
		t.Fatal("warrior squire should use UpgradeOptions, not UpgradeToUnitID")
	}
	am, ok := GetUnitTemplate(EmpireArcherMarksmanBase)
	if !ok || am.UpgradeToUnitID != EmpireArcherPure1 {
		t.Fatalf("archer upgrade: %+v", am)
	}
	ha, ok := GetUnitTemplate(EmpireHealerAcolyte)
	if !ok || len(ha.UpgradeOptions) != 2 || ha.UpgradeOptions[0] != EmpireHealerSingle1 || ha.UpgradeOptions[1] != EmpireHealerGroup1 {
		t.Fatalf("healer branches: %+v", ha)
	}
}

func TestHealerBranchTemplates_groupVsSingle(t *testing.T) {
	sg, ok := GetUnitTemplate(EmpireHealerSingle1)
	if !ok {
		t.Fatal("single")
	}
	gr, ok := GetUnitTemplate(EmpireHealerGroup1)
	if !ok {
		t.Fatal("group")
	}
	if len(sg.Abilities) == 0 || sg.Abilities[0] != battlepkg.AbilityHeal {
		t.Fatalf("single abilities: %+v", sg.Abilities)
	}
	if len(gr.Abilities) == 0 || gr.Abilities[0] != battlepkg.AbilityGroupHeal {
		t.Fatalf("group abilities: %+v", gr.Abilities)
	}
	if gr.HealPower != 0 || sg.HealPower != 1 {
		t.Fatalf("HealPower gr=%d sg=%d", gr.HealPower, sg.HealPower)
	}
}

func TestPromotionTargetUnitIDs_warriorSquire(t *testing.T) {
	ws, ok := GetUnitTemplate(EmpireWarriorSquire)
	if !ok {
		t.Fatal("missing squire")
	}
	ids := PromotionTargetUnitIDs(ws)
	if len(ids) != 2 || ids[0] != EmpireWarriorTank1 || ids[1] != EmpireWarriorDD1 {
		t.Fatalf("got %#v", ids)
	}
}

func TestUpgradeToUnitID_onWarriorRecruit(t *testing.T) {
	tpl, ok := GetUnitTemplate(EmpireWarriorRecruit)
	if !ok || tpl.UpgradeToUnitID != "empire_warrior_squire" {
		t.Fatalf("UpgradeToUnitID: %+v ok=%v", tpl, ok)
	}
}

func TestEarlyRecruitUnitIDs_nonempty(t *testing.T) {
	ids := EarlyRecruitUnitIDs()
	if len(ids) < 2 {
		t.Fatalf("expected pool size >= 2, got %d", len(ids))
	}
	for _, id := range ids {
		if _, ok := GetUnitTemplate(id); !ok {
			t.Fatalf("pool id %q not in registry", id)
		}
	}
}
