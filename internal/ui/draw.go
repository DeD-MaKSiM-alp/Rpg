package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
)

// DrawDebugInputDirection рисует raw и выданное (emit) направление ввода (временный debug для проверки диагонали).
func DrawDebugInputDirection(screen *ebiten.Image, rawDX, rawDY, emitDX, emitDY int) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Input raw: dx=%d dy=%d | emit: dx=%d dy=%d\n", rawDX, rawDY, emitDX, emitDY))
}

// DrawHUD рисует поверх кадра элементы HUD: предметы, знаки обучения, прогресс лидера, строка готовности повышения (если не пустая).
// leader может быть nil; lay — зоны из ComputeScreenLayout / BuildExploreLayoutBundle; promotionLine — из game.PromotionExploreHUDLine.
func DrawHUD(screen *ebiten.Image, pickupCount, trainingMarks int, hudFace *text.GoTextFace, leader *hero.Hero, lay ScreenLayout, promotionLine string) {
	drawHUDText(screen, pickupCount, trainingMarks, hudFace, leader, lay, promotionLine)
}

// DrawBattleOverlay рисует поверх кадра HUD для боевого режима.
// Использует battle.LayoutStyle: v1 = табличный overlay, v2 = Disciples-like (сцена по центру, ростеры по бокам, панель внизу).
// inspectOpenID/inspectOpen — визуальная связь карточек/токенов с открытой battle inspect (hover рисуется отдельным слоем).
func DrawBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, screenWidth, screenHeight int, inspectOpenID battlepkg.UnitID, inspectOpen bool) {
	if battle == nil {
		return
	}
	layout := battle.ComputeBattleHUDLayoutAnchored(screenWidth, screenHeight)
	if layout.Style == battlepkg.LayoutStyleV2Disciples {
		drawBattleScreenV2(screen, hudFace, battle, layout, inspectOpenID, inspectOpen, screenWidth, screenHeight)
		return
	}
	drawBattleOverlayPanel(screen, screenWidth, screenHeight, layout)
	drawBattleOverlayText(screen, hudFace, battle, layout, screenWidth, screenHeight)
}

// DrawResolutionIndicator рисует в правом верхнем углу строку "Resolution: WxH" (runtime switch по F6/F7).
func DrawResolutionIndicator(screen *ebiten.Image, hudFace *text.GoTextFace, screenWidth, screenHeight int) {
	if hudFace == nil || screenWidth < 100 || screenHeight < 24 {
		return
	}
	const lineH = 18
	const width = 220
	r := rect{
		X: float32(screenWidth) - width - 8,
		Y: 8,
		W: width,
		H: lineH,
	}
	metrics := battlepkg.HUDMetrics{LineH: lineH}
	drawSingleLineInRect(screen, hudFace, r, fmt.Sprintf("Resolution: %dx%d", screenWidth, screenHeight), metrics, Theme.TextMuted)
}
