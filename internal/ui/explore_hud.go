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
// lay — зоны из ComputeScreenLayout / BuildExploreLayoutBundle (левая колонка не заезжает на нижнюю полосу).
func DrawExplorePartyStrip(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, lay ScreenLayout, promoStrip string) {
	if hudFace == nil || p == nil || len(p.Active) == 0 {
		return
	}
	if lay.LeftPanel.W <= 8 || lay.LeftPanel.H <= 8 {
		return
	}
	lineH := lay.Preset.LineH
	if lineH < 12 {
		lineH = 12
	}
	pad := float32(8)
	if lay.Tier == TierSmall {
		pad = 6
	}
	maxW := lay.LeftPanel.W
	x := lay.LeftPanel.X
	y := lay.LeftPanel.Y

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
	if lay.LeftPanel.H > 0 && panelH > lay.LeftPanel.H {
		scale := lay.LeftPanel.H / panelH
		if scale < 0.5 {
			scale = 0.5
		}
		lineH *= scale
		if lineH < 11 {
			lineH = 11
		}
		panelH = pad*2 + lineH*float32(n+extraLines+extraProg+extraPromo) + float32(n)*6 + float32(extraLines)*4
	}

	metrics := battlepkg.HUDMetrics{LineH: lineH}
	vector.FillRect(screen, x, y, maxW, panelH, Theme.PanelBGDeep, false)
	vector.FillRect(screen, x, y, 4, panelH, Theme.AccentStrip, false)
	vector.StrokeRect(screen, x, y, maxW, panelH, 1, Theme.PostBattleBorder, false)
	DrawThinAccentLine(screen, x+6, y+4, maxW-12)

	title := "В строю (между боями)"
	if nr > 0 {
		title = fmt.Sprintf("В строю · резерв %d", nr)
	}
	titleR := rect{X: x + 8, Y: y + 8, W: maxW - 16, H: lineH * 0.9}
	drawSingleLineInRect(screen, hudFace, titleR, title, metrics, Theme.TextMuted)

	rowY := y + 8 + lineH + 4
	if leaderProg != "" {
		pr := rect{X: x + 8, Y: rowY, W: maxW - 16, H: lineH * 1.05}
		drawSingleLineInRect(screen, hudFace, pr, PrimaryLine(hudFace, leaderProg, maxW-16), metrics, Theme.TextSecondary)
		rowY += lineH + 2
	}
	if strings.TrimSpace(promoStrip) != "" {
		pr := rect{X: x + 8, Y: rowY, W: maxW - 16, H: lineH * 1.05}
		drawSingleLineInRect(screen, hudFace, pr, PrimaryLine(hudFace, promoStrip, maxW-16), metrics, Theme.RecoveryBanner)
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
func DrawExploreHintPanelLayout(screen *ebiten.Image, bundle ExploreLayoutBundle) (firstY, lineStep float32) {
	lineStep = bundle.Layout.Preset.BottomLineStep
	if lineStep < 16 {
		lineStep = 16
	}
	n := exploreHintLineCount(bundle.ZoneLine, bundle.RestFeedback, bundle.RecruitFeedback, bundle.POIFeedback, bundle.InteractionHint)
	pad := bundle.Layout.Preset.BottomChromePad
	h := float32(n)*lineStep + pad*2
	sw := float32(bundle.Layout.ScreenW)
	y0 := float32(bundle.Layout.ScreenH) - h
	drawUnifiedBottomBarChrome(screen, 0, y0, sw, h)
	return y0 + pad, lineStep
}

// DrawExploreFormationHintLines — текст подсказок поверх DrawExploreHintPanelLayout.
func DrawExploreFormationHintLines(screen *ebiten.Image, hudFace *text.GoTextFace, bundle ExploreLayoutBundle, firstY, lineStep float32) {
	if hudFace == nil {
		return
	}
	y := firstY
	maxW := float32(bundle.Layout.ScreenW) - 28
	narrow := bundle.Layout.Tier == TierSmall
	if strings.TrimSpace(bundle.ZoneLine) != "" {
		line := PrimaryLine(hudFace, bundle.ZoneLine, maxW)
		opZ := &text.DrawOptions{}
		opZ.GeoM.Translate(14, float64(y))
		opZ.ColorScale.ScaleWithColor(Theme.TextSecondary)
		text.Draw(screen, line, hudFace, opZ)
		y += lineStep
	}
	if bundle.InteractionHint != "" {
		line := SecondaryLine(hudFace, bundle.InteractionHint, maxW, narrow)
		op := &text.DrawOptions{}
		op.GeoM.Translate(14, float64(y))
		op.ColorScale.ScaleWithColor(Theme.HoverTarget)
		text.Draw(screen, line, hudFace, op)
		y += lineStep
	}
	if bundle.RestFeedback != "" {
		line := PrimaryLine(hudFace, bundle.RestFeedback, maxW)
		opF := &text.DrawOptions{}
		opF.GeoM.Translate(14, float64(y))
		opF.ColorScale.ScaleWithColor(Theme.RecoveryBanner)
		text.Draw(screen, line, hudFace, opF)
		y += lineStep
	}
	if bundle.RecruitFeedback != "" {
		line := PrimaryLine(hudFace, bundle.RecruitFeedback, maxW)
		opRec := &text.DrawOptions{}
		opRec.GeoM.Translate(14, float64(y))
		opRec.ColorScale.ScaleWithColor(Theme.TextSuccess)
		text.Draw(screen, line, hudFace, opRec)
		y += lineStep
	}
	if strings.TrimSpace(bundle.POIFeedback) != "" {
		line := PrimaryLine(hudFace, bundle.POIFeedback, maxW)
		opP := &text.DrawOptions{}
		opP.GeoM.Translate(14, float64(y))
		opP.ColorScale.ScaleWithColor(Theme.ValidTarget)
		text.Draw(screen, line, hudFace, opP)
		y += lineStep
	}
	rFull := "R — отдых: ход мира без лечения ОЗ (лечение — бой, POI, предметы…)"
	rCompact := "R — отдых: без лечения ОЗ в этом режиме"
	rLine := CompactLine(hudFace, rFull, rCompact, maxW)
	opR := &text.DrawOptions{}
	opR.GeoM.Translate(14, float64(y))
	opR.ColorScale.ScaleWithColor(Theme.HintLine)
	text.Draw(screen, rLine, hudFace, opR)
	y += lineStep
	f5Full := "F5 — состав (I: опыт, знаки, повышение) · лагерь · F9 — демо-рекрут"
	f5Compact := "F5 — состав · F9 — рекрут"
	f5Line := CompactLine(hudFace, f5Full, f5Compact, maxW)
	op := &text.DrawOptions{}
	op.GeoM.Translate(14, float64(y))
	op.ColorScale.ScaleWithColor(Theme.TextSecondary)
	text.Draw(screen, f5Line, hudFace, op)
}

// DrawExploreFormationHint — подсказки F5/R/F9 и баннеры recovery/recruit/POI; общий стиль с explore bar.
func DrawExploreFormationHint(screen *ebiten.Image, hudFace *text.GoTextFace, bundle ExploreLayoutBundle) {
	if hudFace == nil || bundle.Layout.ScreenH < 40 {
		return
	}
	firstY, step := DrawExploreHintPanelLayout(screen, bundle)
	DrawExploreFormationHintLines(screen, hudFace, bundle, firstY, step)
}
