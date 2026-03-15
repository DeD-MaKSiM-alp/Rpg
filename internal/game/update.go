package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	battlepkg "mygame/internal/battle"
	playerpkg "mygame/internal/player"
)

// post-battle reward panel layout (must match ui/postbattle.go)
const (
	postBattlePanelW   = 400
	postBattlePad      = 24
	postBattleLineH    = 22
	postBattleOptionH  = 32
	postBattleOptionGap = 4
	postBattleOptionStartY = 77 // innerY + lineH*3.5 with innerY = panelY+pad
)

func wrapRewardIndex(idx, n int) int {
	if n <= 0 {
		return 0
	}
	for idx < 0 {
		idx += n
	}
	if idx >= n {
		idx %= n
	}
	return idx
}

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

	// Post-battle flow: result screen → (on victory) reward selection → return to world.
	if g.postBattleStep != PostBattleStepNone {
		g.updatePostBattle()
		return
	}

	g.battle.LayoutStyle = g.BattleHUDStyle
	outcome := g.battle.Update()

	switch outcome {
	case battlepkg.BattleOutcomeVictory:
		g.resolveBattleResult(outcome)
		g.postBattleOutcome = outcome
		g.postBattleStep = PostBattleStepResult
		return
	case battlepkg.BattleOutcomeDefeat:
		g.resolveBattleResult(outcome)
		g.postBattleOutcome = outcome
		g.postBattleStep = PostBattleStepResult
		return
	case battlepkg.BattleOutcomeRetreat:
		g.resolveBattleResult(outcome)
		g.postBattleOutcome = outcome
		g.postBattleStep = PostBattleStepResult
		return
	case battlepkg.BattleOutcomeNone:
		return
	}
}

func (g *Game) updatePostBattle() {
	n := len(RewardOptions)
	if n == 0 {
		n = 1
	}
	g.rewardSelectedIndex = wrapRewardIndex(g.rewardSelectedIndex, n)

	switch g.postBattleStep {
	case PostBattleStepResult:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if g.postBattleOutcome == battlepkg.BattleOutcomeVictory {
				g.postBattleStep = PostBattleStepReward
				g.rewardSelectedIndex = 0
			} else {
				g.endBattle()
			}
		}
	case PostBattleStepReward:
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			g.rewardSelectedIndex = wrapRewardIndex(g.rewardSelectedIndex-1, n)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			g.rewardSelectedIndex = wrapRewardIndex(g.rewardSelectedIndex+1, n)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			ApplyReward(&g.progression, RewardOptions[g.rewardSelectedIndex])
			g.endBattle()
			return
		}
		// Mouse: click on reward option (layout matches ui/postbattle.go)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if idx := g.rewardOptionAtCursor(); idx >= 0 && idx < n {
				ApplyReward(&g.progression, RewardOptions[idx])
				g.endBattle()
			}
		}
	}
}

func (g *Game) rewardOptionAtCursor() int {
	mx, my := ebiten.CursorPosition()
	w, h := ScreenWidth, ScreenHeight
	panelW := postBattlePanelW
	if panelW > w-postBattlePad*2 {
		panelW = w - postBattlePad*2
	}
	panelX := (w - panelW) / 2
	panelH := 220
	if len(RewardOptions) > 0 {
		panelH = 120 + len(RewardOptions)*36
	}
	panelY := (h - panelH) / 2
	innerX := panelX + postBattlePad
	innerW := panelW - postBattlePad*2
	if mx < innerX || mx > innerX+innerW {
		return -1
	}
	optionY := panelY + postBattlePad + postBattleOptionStartY
	for i := 0; i < len(RewardOptions); i++ {
		if my >= optionY && my < optionY+postBattleOptionH+postBattleOptionGap {
			return i
		}
		optionY += postBattleOptionH + postBattleOptionGap
	}
	return -1
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
