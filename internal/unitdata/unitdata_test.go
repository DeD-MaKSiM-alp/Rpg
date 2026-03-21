package unitdata

import (
	"errors"
	"testing"
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
