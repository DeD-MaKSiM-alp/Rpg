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
	panelH := lineH*6 + 48
	px := (w - panelW) / 2
	py := (h - panelH) / 2

	vector.FillRect(screen, px, py, panelW, panelH, Theme.PostBattlePanelBG, false)
	vector.StrokeRect(screen, px, py, panelW, panelH, 2, Theme.PostBattleBorder, false)

	metrics := battlepkg.HUDMetrics{LineH: lineH}
	innerX := px + 16
	y := py + 16

	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.2}, "Лагерь наёмников", metrics, Theme.TextPrimary)
	y += lineH + 8
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.2}, "Принять нового бойца в резерв?", metrics, Theme.TextSecondary)
	y += lineH + 12
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH}, "Enter / Space / Y — да · Esc / N — отказ", metrics, Theme.TextMuted)
	y += lineH + 8
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.2}, "Принятый союзник идёт в резерв (F5 — состав).", metrics, Theme.TextMuted)
}
