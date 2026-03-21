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

// DrawHUD рисует поверх кадра элементы HUD (например, счётчик собранных предметов и знаки обучения).
// Цена повышения зависит от tier цели — см. карточку бойца в составе (F5 → I).
func DrawHUD(screen *ebiten.Image, pickupCount, trainingMarks int, hudFace *text.GoTextFace) {
	drawHUDText(screen, pickupCount, trainingMarks, hudFace)
}

// DrawBattleOverlay рисует поверх кадра HUD для боевого режима.
// Использует battle.LayoutStyle: v1 = табличный overlay, v2 = Disciples-like (сцена по центру, ростеры по бокам, панель внизу).
// inspectOpenID/inspectOpen — визуальная связь карточек/токенов с открытой battle inspect (hover рисуется отдельным слоем).
func DrawBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, screenWidth, screenHeight int, inspectOpenID battlepkg.UnitID, inspectOpen bool) {
	if battle == nil {
		return
	}
	layout := battle.ComputeBattleHUDLayout(screenWidth, screenHeight)
	if layout.Style == battlepkg.LayoutStyleV2Disciples {
		drawBattleScreenV2(screen, hudFace, battle, layout, inspectOpenID, inspectOpen)
		return
	}
	drawBattleOverlayPanel(screen, screenWidth, screenHeight, layout)
	drawBattleOverlayText(screen, hudFace, battle, layout)
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
