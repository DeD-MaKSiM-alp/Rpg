package ui

import (
	"strings"
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

func TestFormationInspectCardModel_hasProgressionNoEnemyFlag(t *testing.T) {
	h := hero.DefaultHero()
	m := buildFormationInspectCardModel(&h, 0, 1, false, false, 0, nil, nil, 0, "")
	if m.IsEnemy {
		t.Fatal("formation ally")
	}
	if len(m.ProgressLines) == 0 {
		t.Fatal("want progression")
	}
}

func TestFlattenInspectCardText_noRawUnitIDInFormation(t *testing.T) {
	secret := "hidden_unit_id_xyz"
	h := hero.DefaultHero()
	h.UnitID = secret
	m := buildFormationInspectCardModel(&h, 0, 1, false, false, 0, nil, nil, 0, "")
	s := FlattenInspectCardText(m)
	if strings.Contains(s, secret) {
		t.Fatal("raw unit id should not appear in card text")
	}
}

func TestFormationInspectCard_mergedStatsAndNoExtraHealLine(t *testing.T) {
	h := hero.DefaultHero()
	m := buildFormationInspectCardModel(&h, 0, 1, false, false, 0, nil, nil, 0, "")
	if m.ExtraStatLine != "" {
		t.Fatalf("ExtraStatLine should be merged into StatsLine, got %q", m.ExtraStatLine)
	}
	if !strings.Contains(m.StatsLine, "Лечение") {
		t.Fatalf("heal should stay in StatsLine: %q", m.StatsLine)
	}
}

func TestInspectCard_progressionLineBudget(t *testing.T) {
	h := hero.DefaultHero()
	u := &battlepkg.CombatUnit{
		Side: battlepkg.TeamPlayer,
		Def:  battlepkg.CombatUnitDefinition{Base: battlepkg.UnitBaseStats{MaxHP: h.MaxHP}},
		State: battlepkg.CombatUnitState{
			HP:    h.CurrentHP,
			Alive: true,
		},
		Origin: battlepkg.CombatUnitOrigin{PartyActiveIndex: 0},
	}
	battleM := buildBattleInspectAllyModel(&h, u, 0, 1, 0, nil, nil, 0)
	if len(battleM.ProgressLines) > maxBattleInspectProgressLines {
		t.Fatalf("battle progression too long: %d lines %v", len(battleM.ProgressLines), battleM.ProgressLines)
	}
	formM := buildFormationInspectCardModel(&h, 0, 1, false, false, 0, nil, nil, 0, "")
	if len(formM.ProgressLines) > maxFormationInspectProgressLines {
		t.Fatalf("formation progression too long: %d lines %v", len(formM.ProgressLines), formM.ProgressLines)
	}
}

func TestInspectCard_abilitiesStillListed(t *testing.T) {
	h := hero.DefaultHero()
	m := buildFormationInspectCardModel(&h, 0, 1, false, false, 0, nil, nil, 0, "")
	flat := FlattenInspectCardText(m)
	if !strings.Contains(flat, "Базовый удар") {
		t.Fatalf("expected default ability label in card: %q", flat)
	}
}

func TestInspectPromotionLines_branchHero_compactTwoLines(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate(unitdata.EmpireWarriorSquire)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := hero.PromotionTargetUnitIDs(&h)
	if err != nil || len(targets) < 2 {
		t.Fatalf("need 2 branches for this test: targets=%v err=%v", targets, err)
	}
	costs := []int{1, 1}
	lines := inspectPromotionLines(&h, true, 99, targets, costs, 0)
	if len(lines) > 2 {
		t.Fatalf("want at most 2 lines for chosen branch, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "▸") {
		t.Fatalf("expected branch marker on selected line: %v", lines)
	}
}
