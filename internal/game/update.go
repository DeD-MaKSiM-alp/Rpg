package game

import (
	battlepkg "mygame/internal/battle"
	inputpkg "mygame/internal/input"
)

// Update обрабатывает один кадр игры.
func (g *Game) Update() error {
	if g.mode == ModeBattle {
		g.updateBattleMode()
		return nil
	}
	return g.updateExploreMode()
}

func (g *Game) updateExploreMode() error {
	newDirection := inputpkg.ReadDirectionInput()

	if g.hasBufferedInput {
		inputpkg.UpdateBufferedInput(newDirection, &g.bufferedDirection, &g.bufferTicksLeft, &g.hasBufferedInput, g.TryMovePlayer)
	} else {
		inputpkg.StartInputBufferIfNeeded(newDirection, &g.bufferedDirection, &g.bufferTicksLeft, &g.hasBufferedInput, inputBufferTicks)
	}

	g.updateCamera()
	g.updateStreamingWorld()
	return nil
}

func (g *Game) updateBattleMode() {
	if g.battle == nil {
		g.endBattle()
		return
	}

	action := g.battle.Update()

	switch action {
	case battlepkg.BattleActionVictory:
		g.world.RemoveEnemy(g.battle.EnemyID)
		g.endBattle()
		return
	case battlepkg.BattleActionDefeat:
		g.endBattle()
		return
	case battlepkg.BattleActionRetreat:
		g.endBattle()
		return
	case battlepkg.BattleActionNone:
		return
	}
}
