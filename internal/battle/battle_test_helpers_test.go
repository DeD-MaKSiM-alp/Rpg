package battle

import "testing"

// Shared helpers for battle test suites (no Ebiten).

func testingBeginPlayerAwaitAction(t *testing.T, ctx *BattleContext) *BattleUnit {
	t.Helper()
	for i, id := range ctx.TurnOrder {
		u := ctx.Units[id]
		if u != nil && u.Side == TeamPlayer && u.IsAlive() {
			ctx.TurnIndex = i
			ctx.Phase = PhaseAwaitAction
			ctx.Result = ResultNone
			ctx.PlayerTurn.Reset()
			return u
		}
	}
	t.Fatal("no living player in turn order")
	return nil
}

func testingBeginEnemyAwaitAction(t *testing.T, ctx *BattleContext) *BattleUnit {
	t.Helper()
	for i, id := range ctx.TurnOrder {
		u := ctx.Units[id]
		if u != nil && u.Side == TeamEnemy && u.IsAlive() {
			ctx.TurnIndex = i
			ctx.Phase = PhaseAwaitAction
			ctx.Result = ResultNone
			ctx.PlayerTurn.Reset()
			return u
		}
	}
	t.Fatal("no living enemy in turn order")
	return nil
}

func testingFirstEnemy(ctx *BattleContext) *BattleUnit {
	for _, id := range ctx.TurnOrder {
		u := ctx.Units[id]
		if u != nil && u.Side == TeamEnemy && u.IsAlive() {
			return u
		}
	}
	return nil
}

func testingFirstPlayer(ctx *BattleContext) *BattleUnit {
	for _, id := range ctx.TurnOrder {
		u := ctx.Units[id]
		if u != nil && u.Side == TeamPlayer && u.IsAlive() {
			return u
		}
	}
	return nil
}

// testingApplyPlayerResolvedAction mirrors Update when the player submits a valid BattleAction.
func testingApplyPlayerResolvedAction(ctx *BattleContext, act BattleAction) {
	res := ResolveAbility(ctx, act)
	ctx.ApplyActionResult(res)
	ctx.PauseFrames = actionPauseFrames
	ctx.PlayerTurn.Reset()
	ctx.Phase = PhaseActionPause
}

// testingSimulateActionPauseThroughTurnEnd drains PhaseActionPause exactly as Update does, then runs PhaseTurnEnd.
func testingSimulateActionPauseThroughTurnEnd(ctx *BattleContext) {
	for ctx.Phase == PhaseActionPause {
		ctx.PauseFrames--
		if ctx.PauseFrames <= 0 {
			if ctx.IsFinished() {
				return
			}
			ctx.Phase = PhaseTurnEnd
			break
		}
	}
	if ctx.Phase != PhaseTurnEnd {
		return
	}
	ctx.UpdateResultIfFinished()
	if ctx.IsFinished() {
		return
	}
	ctx.AdvanceTurn()
	ctx.UpdateResultIfFinished()
	if ctx.IsFinished() {
		return
	}
	ctx.Phase = PhaseTurnStart
}

// testingSimulateTurnStartToAwait mirrors PhaseTurnStart in Update until PhaseAwaitAction or PhaseRoundEnd.
func testingSimulateTurnStartToAwait(ctx *BattleContext) {
	ctx.UpdateResultIfFinished()
	if ctx.IsFinished() {
		return
	}
	for ctx.TurnIndex < len(ctx.TurnOrder) {
		u := ctx.ActiveUnit()
		if u != nil && u.IsAlive() {
			ctx.Phase = PhaseAwaitAction
			return
		}
		ctx.TurnIndex++
	}
	ctx.Phase = PhaseRoundEnd
}

// testingSimulateRoundEndToTurnStart mirrors PhaseRoundEnd → PhaseTurnStart.
func testingSimulateRoundEndToTurnStart(ctx *BattleContext) {
	if ctx.Phase != PhaseRoundEnd {
		return
	}
	ctx.Phase = PhaseTurnStart
}

// testingPlayerKeyboardBasicAttackConfirm moves fighter from default basic → choose target → confirm on current target.
func testingPlayerKeyboardBasicAttackConfirm(t *testing.T, ctx *BattleContext, actor *BattleUnit) BattleAction {
	t.Helper()
	ctx.ensurePlayerTurnInitialized(actor)
	confirm := BuildBattleKeyboardIntents(true, false, false, false, false, false, false, false)
	_, ok := ctx.updatePlayerTurnStateMachine(actor, confirm)
	if ok || ctx.PlayerTurn.Phase != PlayerChooseTarget {
		t.Fatalf("want transition to ChooseTarget, ok=%v phase=%v", ok, ctx.PlayerTurn.Phase)
	}
	act, ok2 := ctx.updatePlayerTurnStateMachine(actor, confirm)
	if !ok2 {
		t.Fatal("second confirm should resolve basic attack")
	}
	return act
}
