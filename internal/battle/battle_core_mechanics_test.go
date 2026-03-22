package battle

import (
	"testing"

	"mygame/world/entity"
)

// =============================================================================
// A. Ability availability / validation (resource gate / basic attack)
// =============================================================================

func TestBattleCore_Availability_basicAttackAlwaysFree(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Mana = 0
	actor.State.Energy = 0
	if g := AbilityResourceGate(ctx, actor, AbilityBasicAttack); !g.OK {
		t.Fatalf("basic attack should be free: %s", g.Message)
	}
}

func TestBattleCore_Availability_powerStrikeBlockedByManaEnergyCooldown(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	// energy gate
	actor.State.Energy = 0
	if g := AbilityResourceGate(ctx, actor, AbilityPowerStrike); g.OK {
		t.Fatal("want block by energy")
	}
	actor.State.Energy = 12
	// cooldown gate
	actor.State.AbilityCooldowns = map[AbilityID]int{AbilityPowerStrike: 2}
	if g := AbilityResourceGate(ctx, actor, AbilityPowerStrike); g.OK || g.Code != ErrAbilityOnCooldown {
		t.Fatalf("want cooldown, got ok=%v code=%v", g.OK, g.Code)
	}
	delete(actor.State.AbilityCooldowns, AbilityPowerStrike)
	// heal mana block example (different ability)
	hSeed := BuildPlayerCombatSeed(10, 1, 0, 5, []AbilityID{AbilityHeal, AbilityBasicAttack}, 0, 0)
	ctx2 := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{hSeed}, 0)
	healer := testingBeginPlayerAwaitAction(t, ctx2)
	healer.State.Mana = 0
	if g := AbilityResourceGate(ctx2, healer, AbilityHeal); g.OK || g.Code != ErrInsufficientMana {
		t.Fatalf("want mana block, got ok=%v code=%v", g.OK, g.Code)
	}
}

func TestBattleCore_Availability_availableAbilitiesFiltersByGate(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Energy = 0
	avail := ctx.AvailableAbilities(actor)
	if len(avail) != 1 || avail[0] != AbilityBasicAttack {
		t.Fatalf("want only basic attack when energy 0, got %v", avail)
	}
}

// =============================================================================
// B. Player turn state machine (keyboard + domain click helper, no Ebiten)
// =============================================================================

func TestBattleCore_PlayerTurn_fighterDefaultSelectsBasicAttackNotIndexZero(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	ctx.ensurePlayerTurnInitialized(actor)
	pt := &ctx.PlayerTurn
	if pt.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("want basic attack selected, got %d", pt.SelectedAbilityID)
	}
	abs := actor.Abilities()
	wantIdx := 1
	if len(abs) > wantIdx && abs[wantIdx] == AbilityBasicAttack {
		if pt.SelectedAbilityIndex != wantIdx {
			t.Fatalf("fighter loadout: basic attack index want %d, got %d", wantIdx, pt.SelectedAbilityIndex)
		}
	}
}

func TestBattleCore_PlayerTurn_confirmInvalidTargetAbilityResetsToBasic(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Energy = 12
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	actor.State.AbilityCooldowns = map[AbilityID]int{AbilityPowerStrike: 1}
	kbd := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, ok := ctx.updatePlayerTurnStateMachine(actor, kbd)
	if ok {
		t.Fatal("confirm should not resolve on cooldown")
	}
	pt := &ctx.PlayerTurn
	if pt.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("after failed confirm should reset to basic attack, got ability %d phase %s", pt.SelectedAbilityID, pt.PhaseString())
	}
	if pt.Phase != PlayerChooseAbility {
		t.Fatalf("want ChooseAbility, got %v", pt.Phase)
	}
}

func TestBattleCore_PlayerTurn_targetFlowBackReturnsToBasic(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Energy = 12
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnSelectAbility(actor, &ctx.PlayerTurn, AbilityPowerStrike)
	confirm := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, confirm)
	if ctx.PlayerTurn.Phase != PlayerChooseTarget {
		t.Fatalf("want ChooseTarget, got %v", ctx.PlayerTurn.Phase)
	}
	back := BuildBattleKeyboardIntents(false, false, true, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, back)
	pt := &ctx.PlayerTurn
	if pt.Phase != PlayerChooseAbility || pt.SelectedAbilityID != AbilityBasicAttack {
		t.Fatalf("want back to basic + ChooseAbility, phase=%v abil=%d", pt.Phase, pt.SelectedAbilityID)
	}
}

// =============================================================================
// C. Resolve / cost / cooldown
// =============================================================================

func TestBattleCore_Resolve_powerStrikeSpendsEnergyAndSetsCooldown(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	if enemy == nil {
		t.Fatal("no enemy")
	}
	actor.State.Energy = 12
	beforeE := actor.State.Energy
	act := BattleAction{Actor: actor.ID, Ability: AbilityPowerStrike, Target: enemy.ID}
	res := ResolveAbility(ctx, act)
	if res.Damage <= 0 {
		t.Fatalf("expected damage, got %+v", res)
	}
	costE := GetAbility(AbilityPowerStrike).CostEnergy
	if actor.State.Energy != beforeE-costE {
		t.Fatalf("energy not deducted: got %d want %d", actor.State.Energy, beforeE-costE)
	}
	if rem := actor.State.AbilityCooldowns[AbilityPowerStrike]; rem != GetAbility(AbilityPowerStrike).CooldownRounds {
		t.Fatalf("cooldown want %d, got %d", GetAbility(AbilityPowerStrike).CooldownRounds, rem)
	}
}

func TestBattleCore_Cooldown_availableAfterCooldownTicks(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	if enemy == nil {
		t.Fatal("no enemy")
	}
	actor.State.Energy = 12
	ResolveAbility(ctx, BattleAction{Actor: actor.ID, Ability: AbilityPowerStrike, Target: enemy.ID})
	cd := GetAbility(AbilityPowerStrike).CooldownRounds
	for r := 0; r < cd; r++ {
		ctx.tickRoundResources()
	}
	if g := AbilityResourceGate(ctx, actor, AbilityPowerStrike); !g.OK {
		t.Fatalf("power strike should be off cooldown after %d round ticks: %s", cd, g.Message)
	}
}

// =============================================================================
// D. Round progression (regen + cooldown tick, no off-by-one)
// =============================================================================

func TestBattleCore_Round_tickRoundResourcesRegenAndCooldown(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	actor.State.Mana = 0
	actor.State.Energy = 0
	actor.State.AbilityCooldowns = map[AbilityID]int{AbilityPowerStrike: 2}
	ctx.tickRoundResources()
	if actor.State.AbilityCooldowns[AbilityPowerStrike] != 1 {
		t.Fatalf("cooldown 2 -> 1 after one tick, got %v", actor.State.AbilityCooldowns)
	}
	// regen from resource_profile_test expectations: fighter-like may get small mana tick
	if actor.State.Mana+actor.State.Energy == 0 {
		t.Fatal("expected some regen after tickRoundResources")
	}
}

// =============================================================================
// E. AI safety (never pick gated ability; fallback)
// =============================================================================

func TestBattleCore_AI_buildEnemyActionSkipsUnavailableAndFindsBasic(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	_ = testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	if enemy == nil {
		t.Fatal("no enemy")
	}
	// Drain energy so PowerStrike-like would fail if chosen first in iteration order
	enemy.State.Energy = 0
	enemy.State.AbilityCooldowns = map[AbilityID]int{}
	act, ok := BuildEnemyAction(ctx, enemy)
	if !ok {
		if len(enemy.Abilities()) == 0 {
			return
		}
		t.Fatal("expected some action when abilities exist and targets exist")
	}
	if v := AbilityResourceGate(ctx, enemy, act.Ability); !v.OK {
		t.Fatalf("AI picked gated ability: %v", v.Message)
	}
}

func TestBattleCore_AI_buildEnemyActionReturnsFalseWhenNoAbilities(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{}, 0)
	_, ok := BuildEnemyAction(ctx, nil)
	if ok {
		t.Fatal("nil actor should not produce action")
	}
}

// =============================================================================
// F. Regressions (Power Strike panel / cooldown / cancel)
// =============================================================================

func TestBattleCore_Regression_mouseSpecialClickPathDoesNotCommitOnCooldown(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	ctx.ensurePlayerTurnInitialized(actor)
	playerTurnResetToBasicAttack(actor, &ctx.PlayerTurn)
	actor.State.AbilityCooldowns = map[AbilityID]int{AbilityPowerStrike: 2}
	_, ok := playerTurnTrySpecialAbilityClick(ctx, actor, AbilityPowerStrike)
	if ok {
		t.Fatal("should not resolve on cooldown")
	}
	pt := &ctx.PlayerTurn
	if pt.SelectedAbilityID != AbilityBasicAttack || pt.Phase != PlayerChooseAbility {
		t.Fatalf("selection must stay basic/ChooseAbility when click invalid, got abil=%d phase=%s", pt.SelectedAbilityID, pt.PhaseString())
	}
}

func TestBattleCore_Regression_keyboardSyncDoesNotStickOnPowerStrikeIndexBug(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 5, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	ctx.ensurePlayerTurnInitialized(actor)
	// Simulate inconsistent tuple: index 0 vs id basic — SM must reconcile by index first.
	ctx.PlayerTurn.SelectedAbilityIndex = 0
	ctx.PlayerTurn.SelectedAbilityID = AbilityBasicAttack
	kbd := BuildBattleKeyboardIntents(false, false, false, false, false, false, false, false)
	_, _ = ctx.updatePlayerTurnStateMachine(actor, kbd)
	if ctx.PlayerTurn.SelectedAbilityID != AbilityPowerStrike {
		t.Fatalf("sync from index: index 0 is PowerStrike; got %d", ctx.PlayerTurn.SelectedAbilityID)
	}
	playerTurnResetToBasicAttack(actor, &ctx.PlayerTurn)
	if ctx.PlayerTurn.SelectedAbilityID != AbilityBasicAttack || ctx.PlayerTurn.SelectedAbilityIndex != 1 {
		t.Fatalf("reset must use loadout index for basic, got id=%d idx=%d", ctx.PlayerTurn.SelectedAbilityID, ctx.PlayerTurn.SelectedAbilityIndex)
	}
}
