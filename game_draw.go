package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Draw — метод, который отвечает за отрисовку одного кадра игры.
// Ebiten каждый кадр даёт нам поверхность screen,
// на которой мы последовательно рисуем:
//
//	фон, мир, сетку, (опционально) overlay чанков, игрока, debug-текст и HUD.
func (g *Game) Draw(screen *ebiten.Image) {
	// рисуем фон в черном цвете
	screen.Fill(color.Black)

	// Сначала рисуем сам мир.
	g.world.Draw(screen, g.cameraX, g.cameraY, visibleTilesX, visibleTilesY, tileSize)

	// Затем обычную сетку клеток.
	g.drawGrid(screen)

	// После этого, при включённом debug-режиме,
	// рисуем границы чанков поверх мира и сетки.
	if debugShowChunkOverlay {
		g.world.DrawChunkDebugOverlay(screen, g.cameraX, g.cameraY, visibleTilesX, visibleTilesY, tileSize, screenWidth, screenHeight)
	}

	// Игрока рисуем уже поверх мира и всех сеток,
	// чтобы он не терялся за линиями.
	g.player.Draw(screen, g.cameraX, g.cameraY)

	// При включённом debug-режиме поверх всего рисуем текстовую debug-информацию.
	if debugShowChunkOverlay {
		g.drawDebugInfo(screen)
	}

	g.drawHUD(screen)

	if g.mode == ModeBattle {
		g.drawBattleOverlay(screen)
	}
}
