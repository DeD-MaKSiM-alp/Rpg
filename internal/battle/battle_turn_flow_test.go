package battle

import (
	"testing"

	"mygame/world/entity"
)

// Block A + D: turn lifecycle, round boundaries, action pause, post-action cleanup, next actor state.

func TestBattleTurnFlow_playerActionPauseThenNextActor(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	if enemy == nil {
		t.Fatal("no enemy")
	}
	if ctx.Round != 1 {
		t.Fatalf("want round 1, got %d", ctx.Round)
	}
	act := testingPlayerKeyboardBasicAttackConfirm(t, ctx, actor)
	testingApplyPlayerResolvedAction(ctx, act)
	if ctx.Phase != PhaseActionPause || ctx.PauseFrames != actionPauseFrames {
		t.Fatalf("want action pause, phase=%s pause=%d", ctx.PhaseString(), ctx.PauseFrames)
	}
	if ctx.PlayerTurn.Phase != PlayerTurnNone && ctx.PlayerTurn.Actor != 0 {
		t.Fatal("player turn state should be reset after resolve")
	}
	testingSimulateActionPauseThroughTurnEnd(ctx)
	if ctx.Phase != PhaseTurnStart {
		t.Fatalf("want PhaseTurnStart after pause, got %s", ctx.PhaseString())
	}
	u := ctx.ActiveUnit()
	if u == nil || u.Side != TeamEnemy || u.ID != enemy.ID {
		t.Fatalf("next actor should be enemy, got %+v", u)
	}
	testingSimulateTurnStartToAwait(ctx)
	if ctx.Phase != PhaseAwaitAction {
		t.Fatalf("want AwaitAction for enemy, got %s", ctx.PhaseString())
	}
}

func TestBattleTurnFlow_fullRoundIncrementsRoundAndTicksResources(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	player := testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	if enemy == nil {
		t.Fatal("no enemy")
	}
	player.State.AbilityCooldowns = map[AbilityID]int{AbilityPowerStrike: 2}
	cdBefore := player.State.AbilityCooldowns[AbilityPowerStrike]
	manaBefore := player.State.Mana
	energyBefore := player.State.Energy

	// Player turn
	actP := testingPlayerKeyboardBasicAttackConfirm(t, ctx, player)
	testingApplyPlayerResolvedAction(ctx, actP)
	testingSimulateActionPauseThroughTurnEnd(ctx)
	testingSimulateTurnStartToAwait(ctx)

	// Enemy turn (mirrors Update enemy branch)
	e := ctx.ActiveUnit()
	if e == nil || e.Side != TeamEnemy {
		t.Fatal("enemy should act")
	}
	actE, ok := BuildEnemyAction(ctx, e)
	if ok {
		res := ResolveAbility(ctx, actE)
		ctx.ApplyActionResult(res)
	}
	ctx.PauseFrames = actionPauseFrames
	ctx.PlayerTurn.Reset()
	ctx.Phase = PhaseActionPause
	testingSimulateActionPauseThroughTurnEnd(ctx)

	if ctx.Round != 2 {
		t.Fatalf("want round 2 after full orbit, got %d", ctx.Round)
	}
	if ctx.TurnIndex != 0 {
		t.Fatalf("want turn index 0 at new round, got %d", ctx.TurnIndex)
	}
	if player.State.AbilityCooldowns[AbilityPowerStrike] >= cdBefore {
		t.Fatalf("cooldown should tick down across round boundary: before %d after %d", cdBefore, player.State.AbilityCooldowns[AbilityPowerStrike])
	}
	if player.State.Mana == manaBefore && player.State.Energy == energyBefore {
		t.Log("note: regen may be zero on this profile; cooldown tick is the main signal")
	}
}

func TestBattleTurnFlow_actionFlowConsistencyBasicAttack(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	seed := BuildPlayerCombatSeed(10, 2, 0, 100, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{seed}, 0)
	actor := testingBeginPlayerAwaitAction(t, ctx)
	enemy := testingFirstEnemy(ctx)
	if g := AbilityResourceGate(ctx, actor, AbilityBasicAttack); !g.OK {
		t.Fatal(g.Message)
	}
	act := testingPlayerKeyboardBasicAttackConfirm(t, ctx, actor)
	if g := AbilityResourceGate(ctx, actor, act.Ability); !g.OK {
		t.Fatalf("at commit time basic should still be free: %s", g.Message)
	}
	hp := enemy.State.HP
	testingApplyPlayerResolvedAction(ctx, act)
	if enemy.State.HP >= hp {
		t.Fatal("expected damage applied")
	}
	if ctx.Phase != PhaseActionPause {
		t.Fatalf("want pause after apply, got %s", ctx.PhaseString())
	}
}
