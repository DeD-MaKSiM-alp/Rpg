package ui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/party"
)

// DrawExplorePartyStrip — компактная панель отряда в explore (канонический HP). promoStrip — короткая строка готовности повышения (может быть пустой).
func DrawExplorePartyStrip(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, screenW int, promoStrip string) {
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
	nr := len(p.Reserve)
	extraLines := 0
	if nr > 0 {
		extraLines = 1
	}
	var leaderProg string
	if lh := p.Leader(); lh != nil {
		leaderProg = FormatLeaderExploreStripLine(lh)
	}
	extraProg := 0
	if leaderProg != "" {
		extraProg = 1
	}
	extraPromo := 0
	if strings.TrimSpace(promoStrip) != "" {
		extraPromo = 1
	}
	panelH := pad*2 + lineH*float32(n+extraLines+extraProg+extraPromo) + float32(n)*6 + float32(extraLines)*4
	x := float32(10)
	y := float32(46)

	vector.FillRect(screen, x, y, maxW, panelH, Theme.PanelBGDeep, false)
	vector.FillRect(screen, x, y, 4, panelH, Theme.AccentStrip, false)
	vector.StrokeRect(screen, x, y, maxW, panelH, 1, Theme.PostBattleBorder, false)
	DrawThinAccentLine(screen, x+6, y+4, maxW-12)

	metrics := battlepkg.HUDMetrics{LineH: lineH}
	title := "В строю (между боями)"
	if nr > 0 {
		title = fmt.Sprintf("В строю · резерв %d", nr)
	}
	titleR := rect{X: x + 8, Y: y + 8, W: maxW - 16, H: lineH * 0.9}
	drawSingleLineInRect(screen, hudFace, titleR, title, metrics, Theme.TextMuted)

	rowY := y + 8 + lineH + 4
	if leaderProg != "" {
		pr := rect{X: x + 8, Y: rowY, W: maxW - 16, H: lineH * 1.05}
		drawSingleLineInRect(screen, hudFace, pr, trimTextToWidth(hudFace, leaderProg, maxW-16), metrics, Theme.TextSecondary)
		rowY += lineH + 2
	}
	if strings.TrimSpace(promoStrip) != "" {
		pr := rect{X: x + 8, Y: rowY, W: maxW - 16, H: lineH * 1.05}
		drawSingleLineInRect(screen, hudFace, pr, trimTextToWidth(hudFace, promoStrip, maxW-16), metrics, Theme.RecoveryBanner)
		rowY += lineH + 2
	}
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
	if nr > 0 {
		row := rect{X: x + 10, Y: rowY, W: maxW - 20, H: lineH}
		drawSingleLineInRect(screen, hudFace, row, fmt.Sprintf("Резерв не в бою: %d", nr), metrics, Theme.TextSecondary)
	}
}

// DrawExploreHintPanelLayout возвращает Y первой строки текста и шаг для подсказок explore (после отрисовки подложки).
func DrawExploreHintPanelLayout(screen *ebiten.Image, screenW, screenH int, restFeedback, recruitFeedback, interactionHint string) (firstY, lineStep float32) {
	lineStep = 22
	n := 2
	if interactionHint != "" {
		n++
	}
	if restFeedback != "" {
		n++
	}
	if recruitFeedback != "" {
		n++
	}
	pad := float32(8)
	h := float32(n)*lineStep + pad*2
	y0 := float32(screenH) - h
	drawUnifiedBottomBarChrome(screen, 0, y0, float32(screenW), h)
	return y0 + pad, lineStep
}

// DrawExploreFormationHintLines — текст подсказок поверх DrawExploreHintPanelLayout.
func DrawExploreFormationHintLines(screen *ebiten.Image, hudFace *text.GoTextFace, screenW int, firstY, lineStep float32, restFeedback, recruitFeedback, interactionHint string) {
	if hudFace == nil {
		return
	}
	y := firstY
	maxW := float32(screenW) - 28
	if interactionHint != "" {
		line := trimTextToWidth(hudFace, interactionHint, maxW)
		op := &text.DrawOptions{}
		op.GeoM.Translate(14, float64(y))
		op.ColorScale.ScaleWithColor(Theme.HoverTarget)
		text.Draw(screen, line, hudFace, op)
		y += lineStep
	}
	if restFeedback != "" {
		opF := &text.DrawOptions{}
		opF.GeoM.Translate(14, float64(y))
		opF.ColorScale.ScaleWithColor(Theme.RecoveryBanner)
		text.Draw(screen, restFeedback, hudFace, opF)
		y += lineStep
	}
	if recruitFeedback != "" {
		opRec := &text.DrawOptions{}
		opRec.GeoM.Translate(14, float64(y))
		opRec.ColorScale.ScaleWithColor(Theme.TextSuccess)
		text.Draw(screen, recruitFeedback, hudFace, opRec)
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
	text.Draw(screen, "F5 — состав (I: опыт, знаки, повышение) · лагерь · F9 — демо-рекрут", hudFace, op)
}

// DrawExploreFormationHint — подсказки F5/R/F9 и баннеры recovery/recruit; общий стиль с explore bar.
// interactionHint — динамическая строка «что рядом» (шаг на бой / добычу / лагерь); пусто = не рисовать.
func DrawExploreFormationHint(screen *ebiten.Image, hudFace *text.GoTextFace, screenW, screenH int, restFeedback, recruitFeedback, interactionHint string) {
	if hudFace == nil || screenH < 40 {
		return
	}
	firstY, step := DrawExploreHintPanelLayout(screen, screenW, screenH, restFeedback, recruitFeedback, interactionHint)
	DrawExploreFormationHintLines(screen, hudFace, screenW, firstY, step, restFeedback, recruitFeedback, interactionHint)
}
