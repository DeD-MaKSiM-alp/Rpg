package ui

import (
	"strings"
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

func TestBuildBattleInspectEnemyModel_hasProfileAndStats(t *testing.T) {
	u := &battlepkg.CombatUnit{
		Side: battlepkg.TeamEnemy,
		Def: battlepkg.CombatUnitDefinition{
			DisplayName:    "Тестовый враг",
			TemplateUnitID: unitdata.EmpireWarriorRecruit,
			Base: battlepkg.UnitBaseStats{
				MaxHP:      20,
				Attack:     3,
				Defense:    2,
				Initiative: 7,
			},
			Loadout: battlepkg.AbilityLoadout{Abilities: []battlepkg.AbilityID{battlepkg.AbilityBasicAttack}},
		},
		State: battlepkg.CombatUnitState{HP: 15, Alive: true},
	}
	m := buildBattleInspectEnemyModel(u)
	if m.Title == "" {
		t.Fatal("empty title")
	}
	if m.HPMax != 20 || m.HPCur != 15 {
		t.Fatalf("HP: got %d/%d", m.HPCur, m.HPMax)
	}
	if m.StatsLine == "" || !strings.Contains(m.StatsLine, "Атака") {
		t.Fatalf("stats: %q", m.StatsLine)
	}
	if len(m.ProfileLines) == 0 {
		t.Fatal("expected profile lines")
	}
	if len(m.ProgressLines) != 0 {
		t.Fatalf("enemy should have no progression, got %v", m.ProgressLines)
	}
}

func TestFlattenBattleInspectCardText_noRawTemplateIDLeak(t *testing.T) {
	secretID := "zzz_secret_template_id_for_test_only"
	u := &battlepkg.CombatUnit{
		Side: battlepkg.TeamEnemy,
		Def: battlepkg.CombatUnitDefinition{
			DisplayName:    "Гоблин",
			TemplateUnitID: secretID,
			Base: battlepkg.UnitBaseStats{
				MaxHP: 10,
			},
		},
		State: battlepkg.CombatUnitState{HP: 10, Alive: true},
	}
	m := buildBattleInspectEnemyModel(u)
	flat := FlattenInspectCardText(m)
	if strings.Contains(flat, secretID) {
		t.Fatalf("raw template id leaked into card text")
	}
}

func TestBuildBattleInspectAllyModel_hasProgressionSection(t *testing.T) {
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
	m := buildBattleInspectAllyModel(&h, u, 0, 1, 0, nil, nil, 0, "")
	if len(m.ProgressLines) == 0 {
		t.Fatal("ally should have progression lines")
	}
	joined := strings.Join(m.ProgressLines, " ")
	if !strings.Contains(strings.ToLower(joined), "опыт") {
		t.Fatalf("expected XP line, got %v", m.ProgressLines)
	}
}
