package battle

import (
	"testing"

	"mygame/world/entity"
)

func TestEmpireHealerGroup1Template_hasGroupHealAbility(t *testing.T) {
	// Данные шаблона (через seed как в партии): group — AbilityGroupHeal, single — AbilityHeal.
	gh := BuildPlayerCombatSeed(11, 1, 0, 2, []AbilityID{AbilityGroupHeal, AbilityBasicAttack}, 0, 0)
	sh := BuildPlayerCombatSeed(10, 1, 0, 2, []AbilityID{AbilityHeal, AbilityBasicAttack}, 1, 0)
	if gh.Def.Base.HealPower != 0 {
		t.Fatalf("group seed HealPower bonus: %d", gh.Def.Base.HealPower)
	}
	if sh.Def.Base.HealPower != 1 {
		t.Fatalf("single seed HealPower bonus: %d", sh.Def.Base.HealPower)
	}
	uG := &CombatUnit{Def: gh.Def, State: CombatUnitState{HP: 11, Alive: true}}
	uS := &CombatUnit{Def: sh.Def, State: CombatUnitState{HP: 10, Alive: true}}
	if uG.GroupHealPower() != 1 || uG.HealPower() != 2 {
		t.Fatalf("group: GroupHealPower=%d HealPower=%d", uG.GroupHealPower(), uG.HealPower())
	}
	if uS.HealPower() != 3 {
		t.Fatalf("single HealPower want 3, got %d", uS.HealPower())
	}
}

func TestResolveAbility_groupHeal_healsAllAlliesNotEnemies(t *testing.T) {
	enc := Encounter{
		Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}},
	}
	healerSeed := BuildPlayerCombatSeed(10, 1, 0, 5, []AbilityID{AbilityGroupHeal, AbilityBasicAttack}, 0, 0)
	allySeed := BuildPlayerCombatSeed(10, 2, 0, 3, []AbilityID{AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{healerSeed, allySeed}, 0)
	if ctx == nil {
		t.Fatal("no ctx")
	}
	var healerID UnitID
	for id, u := range ctx.Units {
		if u != nil && u.Side == TeamPlayer {
			for _, a := range u.Abilities() {
				if a == AbilityGroupHeal {
					healerID = id
					break
				}
			}
		}
	}
	if healerID == 0 {
		t.Fatal("no group healer")
	}
	var allyID UnitID
	for id, u := range ctx.Units {
		if u != nil && u.Side == TeamPlayer && id != healerID {
			allyID = id
			break
		}
	}
	enemyID := UnitID(0)
	for id, u := range ctx.Units {
		if u != nil && u.Side == TeamEnemy {
			enemyID = id
			break
		}
	}
	if allyID == 0 || enemyID == 0 {
		t.Fatalf("ally=%d enemy=%d", allyID, enemyID)
	}

	ctx.Units[healerID].State.HP = 3
	ctx.Units[allyID].State.HP = 2
	enemyHPBefore := ctx.Units[enemyID].State.HP

	per := ctx.Units[healerID].GroupHealPower()
	r := ResolveAbility(ctx, BattleAction{Actor: healerID, Ability: AbilityGroupHeal, Target: 0})
	if len(r.HealApplications) != 2 {
		t.Fatalf("HealApplications: %+v", r.HealApplications)
	}
	if ctx.Units[healerID].State.HP != 3+per || ctx.Units[allyID].State.HP != 2+per {
		t.Fatalf("hp healer=%d ally=%d (per=%d)", ctx.Units[healerID].State.HP, ctx.Units[allyID].State.HP, per)
	}
	if ctx.Units[enemyID].State.HP != enemyHPBefore {
		t.Fatal("enemy HP changed")
	}
}

func TestResolveAbility_groupHeal_clampsToMaxHP(t *testing.T) {
	enc := Encounter{
		Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}},
	}
	h := BuildPlayerCombatSeed(10, 1, 0, 5, []AbilityID{AbilityGroupHeal, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{h}, 0)
	var hid UnitID
	for id, u := range ctx.Units {
		if u != nil && u.Side == TeamPlayer {
			hid = id
			break
		}
	}
	ctx.Units[hid].State.HP = 10
	ResolveAbility(ctx, BattleAction{Actor: hid, Ability: AbilityGroupHeal, Target: 0})
	if ctx.Units[hid].State.HP != 10 {
		t.Fatalf("hp=%d", ctx.Units[hid].State.HP)
	}
}

func TestResolveAbility_singleHeal_oneTarget(t *testing.T) {
	enc := Encounter{
		Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}},
	}
	healer := BuildPlayerCombatSeed(10, 1, 0, 5, []AbilityID{AbilityHeal, AbilityBasicAttack}, 1, 0)
	ally := BuildPlayerCombatSeed(10, 2, 0, 3, []AbilityID{AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{healer, ally}, 0)
	var healerID, allyID UnitID
	for id, u := range ctx.Units {
		if u == nil || u.Side != TeamPlayer {
			continue
		}
		hasHeal := false
		for _, a := range u.Abilities() {
			if a == AbilityHeal {
				hasHeal = true
				break
			}
		}
		if hasHeal {
			healerID = id
		} else {
			allyID = id
		}
	}
	if healerID == 0 || allyID == 0 {
		t.Fatal("ids")
	}
	ctx.Units[allyID].State.HP = 1
	ctx.Units[healerID].State.HP = 10
	beforeHealer := ctx.Units[healerID].State.HP
	r := ResolveAbility(ctx, BattleAction{Actor: healerID, Ability: AbilityHeal, Target: allyID})
	if r.HealAmount == 0 {
		t.Fatal("no heal")
	}
	if ctx.Units[allyID].State.HP <= 1 {
		t.Fatal("ally not healed")
	}
	if ctx.Units[healerID].State.HP != beforeHealer {
		t.Fatal("healer should not be healed by single-target on ally")
	}
}

func TestValidateAction_groupHeal_noUnitTarget(t *testing.T) {
	enc := Encounter{
		Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}},
	}
	h := BuildPlayerCombatSeed(10, 1, 0, 5, []AbilityID{AbilityGroupHeal, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{h}, 0)
	var hid UnitID
	for id, u := range ctx.Units {
		if u != nil && u.Side == TeamPlayer {
			hid = id
			break
		}
	}
	req := ActionRequest{Actor: hid, Ability: AbilityGroupHeal, Target: NoTarget()}
	if v := ValidateAction(ctx, req); !v.OK {
		t.Fatalf("valid: %s", v.Message)
	}
	reqBad := ActionRequest{Actor: hid, Ability: AbilityGroupHeal, Target: UnitTarget(hid)}
	if v := ValidateAction(ctx, reqBad); v.OK {
		t.Fatal("expected invalid with unit target")
	}
}
