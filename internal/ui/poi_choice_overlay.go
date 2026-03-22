package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/world/entity"
)

// DrawPOIChoiceOverlay — два варианта risk/reward для руин или алтаря; sel 0/1 подсвечен.
func DrawPOIChoiceOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, screenW, screenH int, kind entity.PickupKind, sel int, altarBoldHPLoss int) {
	if hudFace == nil || screenW < 100 || screenH < 100 {
		return
	}
	w := float32(screenW)
	h := float32(screenH)
	vector.FillRect(screen, 0, 0, w, h, Theme.OverlayDim, false)

	panelW := float32(460)
	if panelW > w-40 {
		panelW = w - 40
	}
	lineH := float32(18)
	rowBlock := lineH*2 + 20
	panelH := lineH*2 + rowBlock*2 + 96
	px := (w - panelW) / 2
	py := (h - panelH) / 2

	drawUnifiedModalPanelChrome(screen, px, py, panelW, panelH)

	metrics := battlepkg.HUDMetrics{LineH: lineH}
	innerX := px + 16
	y := py + 14

	title := "Точка интереса"
	switch kind {
	case entity.PickupKindPOIRuins:
		title = "Руины"
	case entity.PickupKindPOIAltar:
		title = "Алтарь"
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.2}, title, metrics, Theme.TextPrimary)
	DrawThinAccentLine(screen, innerX, y+lineH*1.1, panelW-32)
	y += lineH + 12

	switch kind {
	case entity.PickupKindPOIRuins:
		drawPOIChoiceTwoRows(screen, hudFace, innerX, y, panelW-32, lineH, metrics, sel, rowBlock,
			"Осторожно: +1 боевого опыта каждому в строю.",
			"Риск: 50% — +3 опыта каждому; иначе засада −2 ОЗ (не ниже 1).",
		)
	case entity.PickupKindPOIAltar:
		boldExtra := "лидер теряет ОЗ"
		if altarBoldHPLoss > 0 {
			boldExtra = fmt.Sprintf("лидер −%d ОЗ", altarBoldHPLoss)
		}
		drawPOIChoiceTwoRows(screen, hudFace, innerX, y, panelW-32, lineH, metrics, sel, rowBlock,
			"Скромная жертва: +1 знак обучения.",
			"Смелая клятва: +2 знака; "+boldExtra+".",
		)
	default:
		return
	}

	y += rowBlock*2 + 28
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.1}, "←/→/WASD или 1/2 · Tab — переключить · Enter — подтвердить · Esc — уйти без награды", metrics, Theme.TextMuted)
}

func drawPOIChoiceTwoRows(screen *ebiten.Image, hudFace *text.GoTextFace, innerX, y, maxW, lineH float32, metrics battlepkg.HUDMetrics, sel int, rowH float32, lineA, lineB string) {
	for i, txt := range []string{lineA, lineB} {
		ry := y + float32(i)*(rowH+10)
		bg := Theme.RosterCardContentWell
		stroke := Theme.RosterCardInnerStroke
		if sel == i {
			bg = Theme.PanelBGDeep
			stroke = Theme.AccentStrip
		}
		vector.FillRect(screen, innerX, ry, maxW, rowH, bg, false)
		vector.StrokeRect(screen, innerX, ry, maxW, rowH, 1, stroke, false)
		line := fmt.Sprintf("%d) %s", i+1, txt)
		line = trimTextToWidth(hudFace, line, maxW-16)
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 8, Y: ry + 10, W: maxW - 16, H: lineH * 1.2}, line, metrics, Theme.TextSecondary)
	}
}
