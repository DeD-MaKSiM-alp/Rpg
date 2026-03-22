package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"

	battlepkg "mygame/internal/battle"
)

// DrawDebugInputDirection рисует raw и выданное (emit) направление ввода (только при DevHUDOverlay в game).
func DrawDebugInputDirection(screen *ebiten.Image, rawDX, rawDY, emitDX, emitDY int) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Ввод: raw %d,%d → %d,%d\n", rawDX, rawDY, emitDX, emitDY))
}

// DrawHUD рисует поверх кадра верхний статус-блок: ресурсы и при необходимости готовность к повышению.
// hud — ExploreHUDLayout (BuildExploreHUDLayout или NewExploreHUDLayoutFromScreenLayout); promotionLine — из game.PromotionExploreHUDLine.
func DrawHUD(screen *ebiten.Image, pickupCount, trainingMarks int, hudFace *text.GoTextFace, hud ExploreHUDLayout, promotionLine string) {
	hud = FinalizeExploreHUDTopComposition(hud, promotionLine)
	drawHUDText(screen, pickupCount, trainingMarks, hudFace, hud, promotionLine)
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

// DrawResolutionIndicator рисует в правом верхнем углу строку "Окно WxH" (вкл. только при DevHUDOverlay; F6/F7 меняют пресет).
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
	drawSingleLineInRect(screen, hudFace, r, fmt.Sprintf("Окно %d×%d", screenWidth, screenHeight), metrics, Theme.TextMuted)
}
