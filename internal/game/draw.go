package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"mygame/internal/postbattle"
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

	if g.mode == ModeExplore {
		ui.DrawExplorePartyStrip(screen, g.hudFace, &g.party, ScreenWidth)
		ui.DrawExploreFormationHint(screen, g.hudFace, ScreenWidth, ScreenHeight, g.exploreRestMsg)
	}

	if g.mode == ModeFormation {
		ui.DrawFormationOverlay(screen, g.hudFace, &g.party, g.formationSel, ScreenWidth, ScreenHeight)
	}

	if g.mode == ModeExplore && debugShowInputDirection {
		rawX, rawY := g.input.DebugRaw()
		emitX, emitY := g.input.DebugEmit()
		ui.DrawDebugInputDirection(screen, rawX, rawY, emitX, emitY)
	}

	if g.mode == ModeBattle {
		ui.DrawBattleOverlay(screen, g.hudFace, g.battle, ScreenWidth, ScreenHeight)
		if g.postBattle.IsActive() {
			params := postbattle.BuildPostBattleParams(&g.postBattle, ScreenWidth, ScreenHeight)
			ui.DrawPostBattleOverlay(screen, g.hudFace, params)
		}
	}

	ui.DrawResolutionIndicator(screen, g.hudFace, ScreenWidth, ScreenHeight)
}
