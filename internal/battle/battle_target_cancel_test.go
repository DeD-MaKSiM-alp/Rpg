package battle

import (
	"testing"

	"mygame/world/entity"
)

// Block B: target invalidation / stale selection
// Block C: cancel / back / confirm edge cases

func TestBattleTarget_staleDeadTargetOnConfirmResetsSafely(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	actor.State.Energy = 12
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	confirm := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, confirm)
	if ctx.PlayerTurn.Phase != PlayerChooseTarget {
		t.Fatal("want choose target")
	}
	enemy.State.HP = 0
	enemy.State.Alive = false
	_, ok := ctx.updatePlayerTurnStateMachine(actor, confirm)
	if ok {
		t.Fatal("confirm on dead target should not resolve action")
	}
	pt := &ctx.PlayerTurn
	if pt.Phase != PlayerChooseAbility {
		t.Fatalf("want safe return to ChooseAbility, got %v", pt.Phase)
	}
	if pt.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("want reset to basic attack, got ability %d", pt.SelectedAbilityID)
	}
}

func TestBattleTarget_emptyValidTargetsAfterReenumerateResets(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	actor.State.Energy = 12
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	kbd := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, kbd)
	enemy.State.HP = 0
	enemy.State.Alive = false
	// Force defensive path: stale list cleared so ListValidTargets re-runs.
	ctx.PlayerTurn.ValidTargets = nil
	ctx.PlayerTurn.SelectedTargetIdx = 0
	noOp := BuildBattleKeyboardIntents(false, false, false, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, noOp)
	if ctx.PlayerTurn.Phase != PlayerChooseAbility {
		t.Fatalf("defensive re-enumeration should drop to ChooseAbility, got %v", ctx.PlayerTurn.Phase)
	}
}

func TestBattleCancel_backFromTargetThenBasicAttackPath(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Energy = 12
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	confirm := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, confirm)
	back := BuildBattleKeyboardIntents(false, false, true, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, back)
	if ctx.PlayerTurn.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("want basic after back, got %d", ctx.PlayerTurn.SelectedAbilityID)
	}
	// second cycle: confirm power strike again then back — still recoverable
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, confirm)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, back)
	if ctx.PlayerTurn.SelectedAbilityID != AbilityBasicAttack {
		t.Fatal("repeat back should still reset")
	}
}

func TestBattleCancel_escSpecialModeResetsLikeBack(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	if !ctx.cancelPlayerSpecialOrTargetToBasic() {
		t.Fatal("cancel should handle special selection")
	}
	if ctx.PlayerTurn.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("esc path should reset to basic, got %d", ctx.PlayerTurn.SelectedAbilityID)
	}
}

func TestBattleCancel_escFromTargetSelection(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Energy = 12
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	kbd := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, kbd)
	if ctx.PlayerTurn.Phase != PlayerChooseTarget {
		t.Fatal("want target phase")
	}
	if !ctx.cancelPlayerSpecialOrTargetToBasic() {
		t.Fatal("cancel from target")
	}
	if ctx.PlayerTurn.Phase != PlayerChooseAbility || ctx.PlayerTurn.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("want basic choose ability, phase=%v abil=%d", ctx.PlayerTurn.Phase, ctx.PlayerTurn.SelectedAbilityID)
	}
}

func TestBattleCancel_confirmNoTargetMassHealWhenInvalidMana(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	hSeed := BuildPlayerCombatSeed(10, 1, 0, 100, []AbilityID{AbilityGroupHeal, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{hSeed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Mana = 0
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityGroupHeal)
	confirm := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, ok := ctx.updatePlayerTurnStateMachine(actor, confirm)
	if ok {
		t.Fatal("group heal should not cast with 0 mana")
	}
	if ctx.PlayerTurn.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("want fallback to basic after failed no-target cast, got %d", ctx.PlayerTurn.SelectedAbilityID)
	}
}
