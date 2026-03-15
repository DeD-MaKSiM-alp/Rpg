package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	battlepkg "mygame/internal/battle"
	"mygame/internal/ui"
)

// Draw рисует один кадр игры.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	g.world.Draw(screen, g.cameraX, g.cameraY, WorldViewport.WidthTiles, WorldViewport.HeightTiles, tileSize)
	g.drawGrid(screen)

	if debugShowChunkOverlay {
		viewportW := WorldViewport.WidthTiles * tileSize
		viewportH := WorldViewport.HeightTiles * tileSize
		g.world.DrawChunkDebugOverlay(screen, g.cameraX, g.cameraY, WorldViewport.WidthTiles, WorldViewport.HeightTiles, tileSize, viewportW, viewportH)
	}

	g.player.Draw(screen, g.cameraX, g.cameraY, tileSize)

	if debugShowChunkOverlay {
		g.drawDebugInfo(screen)
	}

	ui.DrawHUD(screen, g.pickupCount, g.hudFace)

	if g.mode == ModeExplore && debugShowInputDirection {
		rawX, rawY := g.input.DebugRaw()
		emitX, emitY := g.input.DebugEmit()
		ui.DrawDebugInputDirection(screen, rawX, rawY, emitX, emitY)
	}

	if g.mode == ModeBattle {
		ui.DrawBattleOverlay(screen, g.hudFace, g.battle, ScreenWidth, ScreenHeight)
		if g.postBattleStep != PostBattleStepNone {
			g.drawPostBattleOverlay(screen)
		}
	}

	ui.DrawResolutionIndicator(screen, g.hudFace, ScreenWidth, ScreenHeight)
}

func (g *Game) drawPostBattleOverlay(screen *ebiten.Image) {
	var resultText string
	switch g.postBattleOutcome {
	case battlepkg.BattleOutcomeVictory:
		resultText = "Victory!"
	case battlepkg.BattleOutcomeDefeat:
		resultText = "Defeat"
	case battlepkg.BattleOutcomeRetreat:
		resultText = "Escaped"
	default:
		resultText = "Battle ended"
	}
	params := ui.PostBattleParams{
		ResultText:    resultText,
		IsRewardStep:  g.postBattleStep == PostBattleStepReward,
		SelectedIndex: g.rewardSelectedIndex,
		ScreenWidth:   ScreenWidth,
		ScreenHeight:  ScreenHeight,
	}
	if params.IsRewardStep && len(g.rewardOffer) > 0 {
		params.OptionLabels = make([]string, len(g.rewardOffer))
		params.OptionDescs = make([]string, len(g.rewardOffer))
		for i := range g.rewardOffer {
			params.OptionLabels[i] = RewardLabel(g.rewardOffer[i])
			params.OptionDescs[i] = RewardDescription(g.rewardOffer[i])
		}
	}
	ui.DrawPostBattleOverlay(screen, g.hudFace, params)
}
