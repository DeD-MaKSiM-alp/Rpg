package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	battlepkg "mygame/internal/battle"
	"mygame/internal/party"
	playerpkg "mygame/internal/player"
)

// exploreRestFeedbackDurationTicks — длительность баннера после отдыха (R) в explore (~3s @ 60fps).
const exploreRestFeedbackDurationTicks = 180

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
		if inpututil.IsKeyJustPressed(ebiten.KeyF8) {
			if g.BattleHUDStyle == 0 {
				g.BattleHUDStyle = 1
			} else {
				g.BattleHUDStyle = 0
			}
		}
		g.updateBattleMode()
		return nil
	}
	if g.mode == ModeFormation {
		g.updateFormationMode()
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
	if g.exploreRestMsgTicks > 0 {
		g.exploreRestMsgTicks--
		if g.exploreRestMsgTicks <= 0 {
			g.exploreRestMsg = ""
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		g.mode = ModeFormation
		g.formationSel = 0
		if n := len(g.party.Active); n == 0 {
			g.mode = ModeExplore
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if len(g.party.Active) > 0 {
			g.party.ApplyWorldRest()
			g.exploreRestMsg = party.RestExploreBanner
			g.exploreRestMsgTicks = exploreRestFeedbackDurationTicks
			g.advanceWorldTurn()
		}
		return nil
	}

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

func (g *Game) updateFormationMode() {
	n := len(g.party.Active)
	if n == 0 {
		g.mode = ModeExplore
		return
	}
	if g.formationSel >= n {
		g.formationSel = n - 1
	}
	if g.formationSel < 0 {
		g.formationSel = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		g.mode = ModeExplore
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && g.formationSel > 0 {
		g.formationSel--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && g.formationSel < n-1 {
		g.formationSel++
	}

	if n >= 2 {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			if g.party.MoveActiveEarlier(g.formationSel) {
				g.formationSel--
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			if g.party.MoveActiveLater(g.formationSel) {
				g.formationSel++
			}
		}
	}
}

func (g *Game) updateBattleMode() {
	if g.battle == nil {
		g.endBattle()
		return
	}

	// Post-battle flow: result screen → (on victory) reward selection → return to world.
	if g.postBattle.IsActive() {
		if g.postBattle.Update(&g.party, ScreenWidth, ScreenHeight) {
			g.endBattle()
		}
		return
	}

	g.battle.LayoutStyle = g.BattleHUDStyle
	outcome := g.battle.Update()

	switch outcome {
	case battlepkg.BattleOutcomeVictory:
		g.syncPartyFromBattle()
		g.resolveBattleResult(outcome)
		g.BattlesWon++
		g.postBattle.Begin(outcome)
		return
	case battlepkg.BattleOutcomeDefeat:
		g.syncPartyFromBattle()
		g.resolveBattleResult(outcome)
		g.postBattle.Begin(outcome)
		return
	case battlepkg.BattleOutcomeRetreat:
		g.syncPartyFromBattle()
		g.resolveBattleResult(outcome)
		g.postBattle.Begin(outcome)
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
