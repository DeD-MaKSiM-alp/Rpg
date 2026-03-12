package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"

	battlepkg "mygame/internal/battle"
)

// DrawDebugInputDirection рисует raw и выданное (emit) направление ввода (временный debug для проверки диагонали).
func DrawDebugInputDirection(screen *ebiten.Image, rawDX, rawDY, emitDX, emitDY int) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Input raw: dx=%d dy=%d | emit: dx=%d dy=%d\n", rawDX, rawDY, emitDX, emitDY))
}

// DrawHUD рисует поверх кадра элементы HUD (например, счётчик собранных предметов).
func DrawHUD(screen *ebiten.Image, pickupCount int, hudFace *text.GoTextFace) {
	drawHUDText(screen, pickupCount, hudFace)
}

// DrawBattleOverlay рисует поверх кадра HUD для боевого режима:
// затемнение фона, центральная панель и текстовые блоки.
func DrawBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, screenWidth, screenHeight int) {
	if battle == nil {
		return
	}
	layout := battle.ComputeBattleHUDLayout(screenWidth, screenHeight)
	drawBattleOverlayPanel(screen, screenWidth, screenHeight, layout)
	drawBattleOverlayText(screen, hudFace, battle, layout)
}
