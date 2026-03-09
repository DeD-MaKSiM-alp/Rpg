package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"

	battlepkg "mygame/internal/battle"
)

// DrawDebugInputDirection рисует одну строку с текущим направлением ввода (временный debug для проверки диагонали).
func DrawDebugInputDirection(screen *ebiten.Image, dx, dy int) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Input: dx=%d dy=%d\n", dx, dy))
}

// DrawHUD рисует поверх кадра элементы HUD (например, счётчик собранных предметов).
func DrawHUD(screen *ebiten.Image, pickupCount int, hudFace *text.GoTextFace) {
	drawHUDText(screen, pickupCount, hudFace)
}

// DrawBattleOverlay рисует поверх кадра HUD для боевого режима:
// затемнение фона, центральная панель и текстовые блоки.
func DrawBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, screenWidth, screenHeight int) {
	drawBattleOverlayPanel(screen, screenWidth, screenHeight)
	drawBattleOverlayText(screen, hudFace, battle)
}
