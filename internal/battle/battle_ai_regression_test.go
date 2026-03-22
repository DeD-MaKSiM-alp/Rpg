package battle

import (
	"testing"

	"mygame/world/entity"
)

// Block E: AI safety (phase 2)
// Block F: regression pack

func testingValidateBattleAction(ctx *BattleContext, act BattleAction) bool {
	if ctx == nil {
		return false
	}
	ab := GetAbility(act.Ability)
	req := ActionRequest{Actor: act.Actor, Ability: act.Ability}
	switch ab.TargetRule {
	case TargetSelf:
		req.Target = SelfTarget()
	case TargetEnemySingle, TargetAllySingle:
		req.Target = UnitTarget(act.Target)
	case TargetAllyTeam:
		req.Target = NoTarget()
	default:
		req.Target = NoTarget()
	}
	return ValidateAction(ctx, req).OK
}

func TestBattleAI_phase2_actionPassesValidateAndResolve(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{}, 0)
	enemy := testingFirstEnemy(ctx)
	if enemy == nil {
		t.Fatal("no enemy")
	}
	act, ok := BuildEnemyAction(ctx, enemy)
	if !ok {
		t.Fatal("expected enemy to act")
	}
	if !testingValidateBattleAction(ctx, act) {
		t.Fatalf("BuildEnemyAction returned action that fails ValidateAction: %+v", act)
	}
	res := ResolveAbility(ctx, act)
	if res.Actor == 0 {
		t.Fatal("ResolveAbility should run for AI-valid action")
	}
}

func TestBattleAI_phase2_prefersLaterAbilityWhenEarlierUnavailable(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	// Archer loadout: RangedAttack costs energy; BasicAttack free — use any unit as actor.
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityRangedAttack, AbilityBasicAttack}, 0, 0)
	seed.Def.Role = RoleArcher
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Energy = 0
	act, ok := BuildEnemyAction(ctx, actor)
	if !ok {
		t.Fatal("want fallback to basic attack")
	}
	if act.Ability != AbilityBasicAttack {
		t.Fatalf("want basic when ranged unaffordable, got %v", act.Ability)
	}
	if !testingValidateBattleAction(ctx, act) {
		t.Fatal("fallback must pass validation")
	}
}

func TestBattleAI_phase2_returnsFalseWhenNoAvailableAbilityWithTargets(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{}, 0)
	enemy := testingFirstEnemy(ctx)
	if enemy == nil {
		t.Fatal("no enemy")
	}
	for _, id := range ctx.TurnOrder {
		if u := ctx.Units[id]; u != nil && u.Side == TeamPlayer {
			u.State.Alive = false
			u.State.HP = 0
		}
	}
	_, ok := BuildEnemyAction(ctx, enemy)
	if ok {
		t.Fatal("no player targets — enemy should not build a single-target attack")
	}
}

// --- Regression block F (phase 2) ---

func TestBattleRegression_doubleCancelEscStillSafe(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	if !ctx.cancelPlayerSpecialOrTargetToBasic() {
		t.Fatal("first cancel")
	}
	if ctx.cancelPlayerSpecialOrTargetToBasic() {
		t.Fatal("second cancel should be no-op (already basic)")
	}
	if ctx.PlayerTurn.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("still basic, got %d", ctx.PlayerTurn.SelectedAbilityID)
	}
}
