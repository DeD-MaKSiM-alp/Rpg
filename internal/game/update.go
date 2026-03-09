package game

import (
	battlepkg "mygame/internal/battle"
	playerpkg "mygame/internal/player"
)

// Update обрабатывает один кадр игры.
func (g *Game) Update() error {
	if g.mode == ModeBattle {
		g.updateBattleMode()
		return nil
	}
	return g.updateExploreMode()
}

func (g *Game) readPlayerAction() PlayerAction {
	dir, ok := g.input.ConsumeDirection()
	if ok {
		return PlayerAction{
			Type: ActionMove,
			DX:   dir.Dx,
			DY:   dir.Dy,
		}
	}

	if g.input.WaitPressed() {
		return PlayerAction{
			Type: ActionWait,
		}
	}

	return PlayerAction{Type: ActionNone}
}

func (g *Game) processWorldTurn() {
	px, py := g.player.Position()
	enemyID, startedBattle := g.world.AdvanceTurn(px, py)
	if startedBattle && enemyID != 0 {
		g.startBattle(enemyID)
		return
	}
	g.updateCamera()
	g.updateStreamingWorld()
}

func (g *Game) updateExploreMode() error {
	action := g.readPlayerAction()

	if action.Type == ActionNone {
		g.updateCamera()
		g.updateStreamingWorld()
		return nil
	}

	switch action.Type {
	case ActionMove:
		moved, enemyID, pickedUp := playerpkg.TryMovePlayer(&g.player, g.world, action.DX, action.DY)
		if pickedUp {
			g.pickupCount++
		}
		if !moved {
			return nil
		}
		if enemyID != 0 {
			g.startBattle(enemyID)
			return nil
		}
		g.processWorldTurn()

	case ActionWait:
		g.processWorldTurn()
	}

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
