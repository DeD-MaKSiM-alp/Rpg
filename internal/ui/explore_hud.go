package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/party"
)

// DrawExplorePartyStrip — компактная панель отряда в explore (канонический HP).
func DrawExplorePartyStrip(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, screenW int) {
	if hudFace == nil || p == nil || len(p.Active) == 0 {
		return
	}
	const lineH = float32(18)
	pad := float32(10)
	maxW := float32(320)
	if float32(screenW)-20 < maxW {
		maxW = float32(screenW) - 20
	}
	n := len(p.Active)
	panelH := pad*2 + lineH*float32(n) + float32(n)*6
	x := float32(10)
	y := float32(46)

	vector.FillRect(screen, x, y, maxW, panelH, Theme.PanelBG, false)
	vector.StrokeRect(screen, x, y, maxW, panelH, 1, Theme.PanelBorder, false)
	DrawThinAccentLine(screen, x+6, y+4, maxW-12)

	metrics := battlepkg.HUDMetrics{LineH: lineH}
	titleR := rect{X: x + 8, Y: y + 8, W: maxW - 16, H: lineH * 0.9}
	drawSingleLineInRect(screen, hudFace, titleR, "Отряд (HP между боями)", metrics, Theme.TextMuted)

	rowY := y + 8 + lineH + 4
	for i := range p.Active {
		h := &p.Active[i]
		role := party.MemberRoleCaption(i)
		lbl := fmt.Sprintf("%d. %s", i+1, role)
		if h.CurrentHP <= 0 {
			lbl += "  — выбыл"
		}
		row := rect{X: x + 10, Y: rowY, W: maxW - 100, H: lineH}
		col := Theme.TextPrimary
		if h.CurrentHP <= 0 {
			col = Theme.DeadText
		}
		drawSingleLineInRect(screen, hudFace, row, lbl, metrics, col)
		barX := x + maxW - 78
		barW := float32(68)
		barH := float32(5)
		barY := rowY + lineH*0.55
		DrawHPBarMicro(screen, barX, barY, barW, barH, h.CurrentHP, h.MaxHP, h.CurrentHP > 0, false)
		rowY += lineH + 6
	}
}

// DrawExploreHintPanelLayout возвращает Y первой строки текста и шаг для подсказок explore (после отрисовки подложки).
func DrawExploreHintPanelLayout(screen *ebiten.Image, screenW, screenH int, restFeedback string) (firstY, lineStep float32) {
	lineStep = 22
	n := 2
	if restFeedback != "" {
		n = 3
	}
	pad := float32(8)
	h := float32(n)*lineStep + pad*2
	y0 := float32(screenH) - h
	vector.FillRect(screen, 0, y0, float32(screenW), h, Theme.ExploreBarBG, false)
	vector.StrokeRect(screen, 0, y0, float32(screenW), h, 1, Theme.ExploreBarBorder, false)
	return y0 + pad, lineStep
}

// DrawExploreFormationHintLines — текст подсказок поверх DrawExploreHintPanelLayout.
func DrawExploreFormationHintLines(screen *ebiten.Image, hudFace *text.GoTextFace, firstY, lineStep float32, restFeedback string) {
	if hudFace == nil {
		return
	}
	y := firstY
	if restFeedback != "" {
		opF := &text.DrawOptions{}
		opF.GeoM.Translate(14, float64(y))
		opF.ColorScale.ScaleWithColor(Theme.RecoveryBanner)
		text.Draw(screen, restFeedback, hudFace, opF)
		y += lineStep
	}
	opR := &text.DrawOptions{}
	opR.GeoM.Translate(14, float64(y))
	opR.ColorScale.ScaleWithColor(Theme.HintLine)
	text.Draw(screen, "R — отдых: +HP живым (¼ MaxHP), 0 HP не поднимает · затем ход мира", hudFace, opR)
	y += lineStep
	op := &text.DrawOptions{}
	op.GeoM.Translate(14, float64(y))
	op.ColorScale.ScaleWithColor(Theme.TextSecondary)
	text.Draw(screen, "F5 — порядок отряда и слоты в бою", hudFace, op)
}

// DrawExploreFormationHint — подсказки F5/R и баннер recovery; общий стиль с explore bar.
func DrawExploreFormationHint(screen *ebiten.Image, hudFace *text.GoTextFace, screenW, screenH int, restFeedback string) {
	if hudFace == nil || screenH < 40 {
		return
	}
	firstY, step := DrawExploreHintPanelLayout(screen, screenW, screenH, restFeedback)
	DrawExploreFormationHintLines(screen, hudFace, firstY, step, restFeedback)
}
