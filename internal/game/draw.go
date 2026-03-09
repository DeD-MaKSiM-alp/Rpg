package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"mygame/internal/ui"
)

// Draw рисует один кадр игры.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	g.world.Draw(screen, g.cameraX, g.cameraY, visibleTilesX, visibleTilesY, tileSize)
	g.drawGrid(screen)

	if debugShowChunkOverlay {
		g.world.DrawChunkDebugOverlay(screen, g.cameraX, g.cameraY, visibleTilesX, visibleTilesY, tileSize, ScreenWidth, ScreenHeight)
	}

	g.player.Draw(screen, g.cameraX, g.cameraY, tileSize)

	if debugShowChunkOverlay {
		g.drawDebugInfo(screen)
	}

	ui.DrawHUD(screen, g.pickupCount, g.hudFace)

	if g.mode == ModeBattle {
		ui.DrawBattleOverlay(screen, g.hudFace, g.battle, ScreenWidth, ScreenHeight)
	}
}
