package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// DrawRecruitOfferOverlay — минимальное подтверждение найма с лагеря на карте (лазурный маркер).
func DrawRecruitOfferOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, screenW, screenH int) {
	if hudFace == nil || screenW < 100 || screenH < 100 {
		return
	}
	w := float32(screenW)
	h := float32(screenH)
	vector.FillRect(screen, 0, 0, w, h, Theme.OverlayDim, false)

	panelW := float32(420)
	if panelW > w-40 {
		panelW = w - 40
	}
	lineH := float32(20)
	panelH := lineH*6 + 52
	px := (w - panelW) / 2
	py := (h - panelH) / 2

	drawUnifiedModalPanelChrome(screen, px, py, panelW, panelH)

	metrics := battlepkg.HUDMetrics{LineH: lineH}
	innerX := px + 20
	y := py + 18

	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 40, H: lineH * 1.2}, "Лагерь наёмников", metrics, Theme.TextPrimary)
	DrawThinAccentLine(screen, innerX, y+lineH*1.15, panelW-40)
	y += lineH + 14
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 40, H: lineH * 1.2}, "Принять нового бойца в резерв?", metrics, Theme.TextSecondary)
	y += lineH + 14
	// Блок-подложка под вопросом (карточное ощущение без смены hit-test).
	qY := y
	qH := lineH*2 + 16
	vector.FillRect(screen, innerX, qY, panelW-40, qH, Theme.RosterCardContentWell, false)
	vector.StrokeRect(screen, innerX, qY, panelW-40, qH, 1, Theme.RosterCardInnerStroke, false)
	drawSingleLineInRect(screen, hudFace, rect{X: innerX + 8, Y: qY + 8, W: panelW - 56, H: lineH}, "Enter / Space / Y — да · Esc / N — отказ", metrics, Theme.TextMuted)
	drawSingleLineInRect(screen, hudFace, rect{X: innerX + 8, Y: qY + 8 + lineH + 4, W: panelW - 56, H: lineH * 1.1}, "Принятый союзник идёт в резерв (F5 — состав).", metrics, Theme.TextSecondary)
}
