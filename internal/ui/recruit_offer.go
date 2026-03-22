package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// DrawRecruitOfferOverlay — лагерь наёмников: две явные кнопки + модальный dim.
// hoverBtn: -1 ни на чём, 0 на «Принять», 1 на «Отказаться».
func DrawRecruitOfferOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, screenW, screenH int, hoverBtn int) {
	if hudFace == nil || screenW < 100 || screenH < 100 {
		return
	}
	lay := LayoutRecruitOffer(screenW, screenH)
	if lay.Panel.W <= 0 {
		return
	}
	w := float32(screenW)
	h := float32(screenH)
	vector.FillRect(screen, 0, 0, w, h, Theme.OverlayDim, false)

	px, py := lay.Panel.X, lay.Panel.Y
	panelW, panelH := lay.Panel.W, lay.Panel.H
	drawUnifiedModalPanelChrome(screen, px, py, panelW, panelH)

	lineH := float32(20)
	metrics := battlepkg.HUDMetrics{LineH: lineH}
	innerX := px + 20
	y := py + 18

	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 40, H: lineH * 1.2}, "Лагерь наёмников", metrics, Theme.TextPrimary)
	DrawThinAccentLine(screen, innerX, y+lineH*1.15, panelW-40)
	y += lineH + 14
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 40, H: lineH * 1.2}, "Принять нового бойца в резерв?", metrics, Theme.TextSecondary)
	y += lineH + 10
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 40, H: lineH}, "Резерв — F5 (состав).", metrics, Theme.TextMuted)
	y += lineH + 12

	drawModalChoiceButton(screen, lay.AcceptBtn, hudFace, "Принять", metrics, hoverBtn == 0, true)
	drawModalChoiceButton(screen, lay.DeclineBtn, hudFace, "Отказаться", metrics, hoverBtn == 1, false)

	y = lay.DeclineBtn.Y + lay.DeclineBtn.H + 10
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 40, H: lineH}, "Enter / Esc · клик по кнопкам или по затемнению — отказ", metrics, Theme.TextMuted)
}

// drawModalChoiceButton — primary (greenish success) vs secondary decline.
func drawModalChoiceButton(screen *ebiten.Image, r FRect, hudFace *text.GoTextFace, label string, metrics battlepkg.HUDMetrics, hovered, primary bool) {
	bg := Theme.ButtonBG
	brd := Theme.ButtonBorder
	if primary {
		bg = Theme.PanelBGDeep
		brd = Theme.AccentStrip
	}
	if hovered {
		bg = Theme.ButtonHoverBG
		brd = Theme.ButtonHoverBorder
		if primary {
			brd = Theme.ValidTarget
		}
	}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, bg, false)
	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1.5, brd, false)
	drawSingleLineInRect(screen, hudFace, rect{X: r.X + 8, Y: r.Y + 9, W: r.W - 16, H: metrics.LineH * 1.1}, label, metrics, Theme.TextPrimary)
}
