package progression

import (
	"strings"
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
	"mygame/world/entity"
)

func TestBuildVictoryProgressionSummary_twoSurvivors(t *testing.T) {
	enc := battlepkg.Encounter{
		Enemies: []battlepkg.EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}},
	}
	a := hero.DefaultHero()
	b := hero.DefaultHero()
	roster := party.Party{
		Active: []hero.Hero{a, b},
	}
	seeds := roster.PlayerCombatSeeds()
	ctx := battlepkg.BuildBattleContextFromEncounter(enc, seeds, 0)
	if ctx == nil {
		t.Fatal("ctx")
	}
	sum := BuildVictoryProgressionSummary(ctx, &roster, 1)
	if len(sum.Lines) < 2 {
		t.Fatalf("lines: %#v", sum.Lines)
	}
	if !strings.Contains(sum.Lines[0], "Боевой опыт") || !strings.Contains(sum.Lines[0], "+1") {
		t.Fatalf("first line: %q", sum.Lines[0])
	}
}

func TestBuildVictoryProgressionSummary_deadNoXP(t *testing.T) {
	enc := battlepkg.Encounter{
		Enemies: []battlepkg.EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}},
	}
	roster := party.Party{Active: []hero.Hero{hero.DefaultHero(), hero.DefaultHero()}}
	seeds := roster.PlayerCombatSeeds()
	ctx := battlepkg.BuildBattleContextFromEncounter(enc, seeds, 0)
	var deadID battlepkg.UnitID
	for id, u := range ctx.Units {
		if u != nil && u.Side == battlepkg.TeamPlayer && u.Origin.PartyActiveIndex == 1 {
			u.State.HP = 0
			u.State.Alive = false
			deadID = id
			break
		}
	}
	if deadID == 0 {
		t.Fatal("no second ally")
	}
	sum := BuildVictoryProgressionSummary(ctx, &roster, 1)
	found := false
	for _, ln := range sum.Lines {
		if strings.Contains(ln, "Поверженные") && strings.Contains(ln, "не получили") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected fallen line: %#v", sum.Lines)
	}
}

func TestBuildVictoryProgressionSummary_reserveLine(t *testing.T) {
	enc := battlepkg.Encounter{
		Enemies: []battlepkg.EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}},
	}
	roster := party.Party{
		Active:  []hero.Hero{hero.DefaultHero()},
		Reserve: []hero.Hero{hero.DefaultHero()},
	}
	seeds := roster.PlayerCombatSeeds()
	ctx := battlepkg.BuildBattleContextFromEncounter(enc, seeds, 0)
	sum := BuildVictoryProgressionSummary(ctx, &roster, 1)
	found := false
	for _, ln := range sum.Lines {
		if strings.Contains(ln, "Резерв") && strings.Contains(ln, "не начисляется") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected reserve line: %#v", sum.Lines)
	}
}

func TestHeroShortLabel_usesTemplateName(t *testing.T) {
	h, err := hero.NewHeroFromUnitTemplate("empire_warrior_recruit")
	if err != nil {
		t.Fatal(err)
	}
	if s := HeroShortLabel(&h, 0); s == "" || s == "—" {
		t.Fatalf("label: %q", s)
	}
}
