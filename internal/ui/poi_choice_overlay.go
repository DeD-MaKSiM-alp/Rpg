package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/world/entity"
)

// DrawPOIChoiceOverlay — руины / алтарь: два варианта-карточки, кнопка подтверждения, зона «уйти».
// hoverOpt: -1 нет; 0/1 — наведение на вариант (визуально независимо от sel).
// hoverConfirm / hoverCancel — наведение на кнопки нижнего ряда.
func DrawPOIChoiceOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, screenW, screenH int, kind entity.PickupKind, sel int, altarBoldHPLoss int, hoverOpt int, hoverConfirm, hoverCancel bool) {
	if hudFace == nil || screenW < 100 || screenH < 100 {
		return
	}
	lay := LayoutPOIChoice(screenW, screenH, kind)
	if lay.Panel.W <= 0 {
		return
	}
	w := float32(screenW)
	h := float32(screenH)
	vector.FillRect(screen, 0, 0, w, h, Theme.OverlayDim, false)

	px, py := lay.Panel.X, lay.Panel.Y
	panelW, panelH := lay.Panel.W, lay.Panel.H
	drawUnifiedModalPanelChrome(screen, px, py, panelW, panelH)

	lineH := float32(18)
	metrics := battlepkg.HUDMetrics{LineH: lineH}
	innerX := px + 16
	y := py + 14

	title := "Точка интереса"
	switch kind {
	case entity.PickupKindPOIRuins:
		title = "Руины"
	case entity.PickupKindPOIAltar:
		title = "Алтарь"
	default:
		return
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.2}, title, metrics, Theme.TextPrimary)
	DrawThinAccentLine(screen, innerX, y+lineH*1.1, panelW-32)
	y += lineH + 12

	var lineA, lineB string
	switch kind {
	case entity.PickupKindPOIRuins:
		lineA = "Осторожно: +1 боевого опыта каждому в строю."
		lineB = "Риск: 50% — +3 опыта каждому; иначе засада −2 ОЗ (не ниже 1)."
	case entity.PickupKindPOIAltar:
		boldExtra := "лидер теряет ОЗ"
		if altarBoldHPLoss > 0 {
			boldExtra = fmt.Sprintf("лидер −%d ОЗ", altarBoldHPLoss)
		}
		lineA = "Скромная жертва: +1 знак обучения."
		lineB = "Смелая клятва: +2 знака; " + boldExtra + "."
	}
	drawPOIChoiceOptionRow(screen, hudFace, lay.Option0, lineA, 0, sel, hoverOpt, metrics, lineH)
	drawPOIChoiceOptionRow(screen, hudFace, lay.Option1, lineB, 1, sel, hoverOpt, metrics, lineH)

	// Confirm — primary
	confirmBg := Theme.PanelBGDeep
	confirmBrd := Theme.AccentStrip
	if hoverConfirm {
		confirmBg = Theme.ButtonHoverBG
		confirmBrd = Theme.ButtonHoverBorder
	}
	vector.FillRect(screen, lay.ConfirmBtn.X, lay.ConfirmBtn.Y, lay.ConfirmBtn.W, lay.ConfirmBtn.H, confirmBg, false)
	vector.StrokeRect(screen, lay.ConfirmBtn.X, lay.ConfirmBtn.Y, lay.ConfirmBtn.W, lay.ConfirmBtn.H, 1.5, confirmBrd, false)
	drawSingleLineInRect(screen, hudFace, rect{X: lay.ConfirmBtn.X + 10, Y: lay.ConfirmBtn.Y + 9, W: lay.ConfirmBtn.W - 20, H: lineH * 1.15}, "Подтвердить выбор", metrics, Theme.TextPrimary)

	// Cancel line (secondary)
	cancelCol := Theme.TextMuted
	if hoverCancel {
		cancelCol = Theme.TextPrimary
	}
	drawSingleLineInRect(screen, hudFace, rect{X: lay.CancelZone.X + 4, Y: lay.CancelZone.Y + 2, W: lay.CancelZone.W - 8, H: lineH * 1.1}, "Уйти без награды (Esc / клик по затемнению)", metrics, cancelCol)
}

func drawPOIChoiceOptionRow(screen *ebiten.Image, hudFace *text.GoTextFace, box FRect, txt string, idx, sel, hoverOpt int, metrics battlepkg.HUDMetrics, lineH float32) {
	line := fmt.Sprintf("%d) %s", idx+1, txt)
	line = trimTextToWidth(hudFace, line, box.W-16)

	bg := Theme.RosterCardContentWell
	stroke := Theme.RosterCardInnerStroke
	if sel == idx {
		bg = Theme.AbilitySelectedBG
		stroke = Theme.AbilitySelectedBrd
	}
	if hoverOpt == idx && hoverOpt != sel {
		bg = Theme.AbilityHoverBG
		stroke = Theme.HoverTarget
	}
	if sel == idx && hoverOpt == idx {
		stroke = Theme.ValidTarget
	}
	vector.FillRect(screen, box.X, box.Y, box.W, box.H, bg, false)
	vector.StrokeRect(screen, box.X, box.Y, box.W, box.H, 1.5, stroke, false)
	drawSingleLineInRect(screen, hudFace, rect{X: box.X + 8, Y: box.Y + 10, W: box.W - 16, H: lineH * 1.2}, line, metrics, Theme.TextSecondary)
}
