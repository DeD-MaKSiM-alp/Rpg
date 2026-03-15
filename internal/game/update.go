package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	battlepkg "mygame/internal/battle"
	playerpkg "mygame/internal/player"
)

// Update обрабатывает один кадр игры.
func (g *Game) Update() error {
	// Runtime resolution switch: F6 = previous preset, F7 = next preset (cyclic).
	n := len(ResolutionPresets)
	if n > 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyF6) {
			ActivePresetIndex = (ActivePresetIndex - 1 + n) % n
			applyResolutionPreset()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF7) {
			ActivePresetIndex = (ActivePresetIndex + 1) % n
			applyResolutionPreset()
		}
	}

	if g.mode == ModeBattle {
		g.updateBattleMode()
		return nil
	}
	return g.updateExploreMode()
}

// readPlayerAction — единственное место чтения explore input; контракт: Input.ReadExploreInput().
func (g *Game) readPlayerAction() PlayerAction {
	dx, dy, wait := g.input.ReadExploreInput()
	g.debugInputDX, g.debugInputDY = dx, dy // временный debug: для отрисовки "Input: dx= dy="
	if wait {
		return PlayerAction{Type: ActionWait}
	}
	if dx != 0 || dy != 0 {
		return PlayerAction{Type: ActionMove, DX: dx, DY: dy}
	}
	return PlayerAction{Type: ActionNone}
}

// advanceWorldTurn — единственная точка вызова AdvanceTurn: ход врагов, затем обновление камеры и стриминга.
func (g *Game) advanceWorldTurn() {
	px, py := g.player.Position()
	enemyID, startedBattle := g.world.AdvanceTurn(px, py)
	if startedBattle && enemyID != 0 {
		g.startBattle(enemyID)
		return
	}
	g.updateCamera()
	g.updateStreamingWorld()
}

// updateExploreMode: Input → PlayerAction → применение действия → при успехе завершение хода → world turn → возможный бой.
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
		g.advanceWorldTurn()

	case ActionWait:
		g.advanceWorldTurn()
	}

	return nil
}

func (g *Game) updateBattleMode() {
	if g.battle == nil {
		g.endBattle()
		return
	}

	outcome := g.battle.Update()

	switch outcome {
	case battlepkg.BattleOutcomeVictory:
		g.resolveBattleResult(outcome)
		g.endBattle()
		return
	case battlepkg.BattleOutcomeDefeat:
		g.resolveBattleResult(outcome)
		g.endBattle()
		return
	case battlepkg.BattleOutcomeRetreat:
		g.resolveBattleResult(outcome)
		g.endBattle()
		return
	case battlepkg.BattleOutcomeNone:
		return
	}
}

// resolveBattleResult применяет результат боя к миру (удаление врагов при победе и т.д.).
func (g *Game) resolveBattleResult(outcome battlepkg.BattleOutcome) {
	switch outcome {
	case battlepkg.BattleOutcomeVictory:
		for _, e := range g.battle.Encounter.Enemies {
			g.world.RemoveEnemy(e.EnemyID)
		}
	case battlepkg.BattleOutcomeDefeat, battlepkg.BattleOutcomeRetreat:
		// Пока ничего не делаем; позже: respawn, потеря прогресса и т.д.
	}
}
