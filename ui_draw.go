package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// drawHUD рисует поверх кадра элементы HUD (например, счётчик собранных предметов).
func (g *Game) drawHUD(screen *ebiten.Image) {
	g.drawHUDText(screen)
}

/*
drawBattleOverlay рисует поверх кадра HUD для боевого режима.
- Затемняет фон;
- Рисует центральную панель;
- Показывает ID активного врага;
- Показывает кнопки для победы и отступления.
*/
func (g *Game) drawBattleOverlay(screen *ebiten.Image) {
	g.drawBattleOverlayPanel(screen)
	g.drawBattleOverlayText(screen)
}
