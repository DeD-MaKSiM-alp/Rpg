package ui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mygame/internal/party"
)

// DrawExplorePartyStrip — компактная панель отряда в explore (канонический HP). promoStrip — короткая строка готовности повышения (может быть пустой).
// Геометрия панели — ComputePartyStripLayout (зона: Layout.LeftPanel).
func DrawExplorePartyStrip(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, hud ExploreHUDLayout, promoStrip string) {
	if hudFace == nil || p == nil || len(p.Active) == 0 {
		return
	}
	lay := hud.Layout
	pl, ok := ComputePartyStripLayout(lay.LeftPanel, lay.Tier, p, promoStrip)
	if !ok {
		return
	}
	pol := ExploreHUDTextPolicyForTier(lay.Tier)
	pad := pl.Pad
	x := pl.Panel.X
	y := pl.Panel.Y
	maxW := pl.Panel.W
	lineH := pl.LineH
	metrics := pl.Metrics
	nr := len(p.Reserve)

	vector.FillRect(screen, x, y, maxW, pl.Panel.H, Theme.ExplorePartyBG, false)
	vector.FillRect(screen, x, y, 4, pl.Panel.H, Theme.ExplorePartyLeftStrip, false)
	vector.FillRect(screen, x, y+pl.Panel.H-1, maxW, 1, Theme.ExploreModuleEdge, false)
	vector.FillRect(screen, x+4, y+4, maxW-8, 2, Theme.PanelTitleSep, false)

	titleR := rect{X: x + 8, Y: y + pad, W: maxW - 16, H: lineH * 0.9}
	drawSingleLineInRect(screen, hudFace, titleR, PartyStripTitle(p), metrics, Theme.TextPrimary)
	sepY := titleR.Y + titleR.H
	vector.StrokeLine(screen, x+8, sepY, x+maxW-8, sepY, 1, Theme.PanelTitleSep, false)

	var leaderProg string
	if lh := p.Leader(); lh != nil {
		leaderProg = FormatLeaderExploreStripLine(lh)
	}
	rowY := y + pad + lineH + pol.PartyTitleToBodyGap
	if leaderProg != "" {
		pr := rect{X: x + 8, Y: rowY, W: maxW - 16, H: lineH * 1.05}
		drawSingleLineInRect(screen, hudFace, pr, PrimaryLine(hudFace, leaderProg, maxW-16), metrics, Theme.TextSecondary)
		rowY += lineH + pol.PartyLeaderGap
	}
	if strings.TrimSpace(promoStrip) != "" {
		pr := rect{X: x + 8, Y: rowY, W: maxW - 16, H: lineH * 1.05}
		drawSingleLineInRect(screen, hudFace, pr, PrimaryLine(hudFace, promoStrip, maxW-16), metrics, Theme.RecoveryBanner)
		rowY += lineH + pol.PartyLeaderGap
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
		rowY += lineH + pol.PartyRowGap
	}
	if nr > 0 {
		row := rect{X: x + 10, Y: rowY, W: maxW - 20, H: lineH}
		drawSingleLineInRect(screen, hudFace, row, fmt.Sprintf("Резерв не в бою: %d", nr), metrics, Theme.TextSecondary)
	}
}

// DrawExploreHintPanelLayout возвращает Y первой строки текста и шаг для подсказок explore (после отрисовки подложки).
func DrawExploreHintPanelLayout(screen *ebiten.Image, bundle ExploreHUDLayout) (firstY, lineStep float32) {
	lineStep = bundle.LineStep
	if lineStep < 16 {
		lineStep = 16
	}
	h := bundle.BottomPanel.H
	sw := float32(bundle.Layout.ScreenW)
	y0 := bundle.BottomPanel.Y
	drawUnifiedBottomBarChrome(screen, 0, y0, sw, h)
	return bundle.BottomText.Y, lineStep
}

func exploreBottomLineColor(kind ExploreBottomLineKind) color.Color {
	switch kind {
	case BottomKindZone:
		return Theme.TextSecondary
	case BottomKindInteraction:
		return Theme.HoverTarget
	case BottomKindHotkeys:
		return Theme.HintLine
	case BottomKindBannerRest:
		return Theme.RecoveryBanner
	case BottomKindBannerRecruit:
		return Theme.TextSuccess
	case BottomKindBannerPOI:
		return Theme.ValidTarget
	default:
		return Theme.TextSecondary
	}
}

// DrawExploreFormationHintLines — текст подсказок поверх DrawExploreHintPanelLayout (порядок из PlanExploreBottomLines).
func DrawExploreFormationHintLines(screen *ebiten.Image, hudFace *text.GoTextFace, bundle ExploreHUDLayout, firstY, lineStep float32) {
	if hudFace == nil {
		return
	}
	maxW := bundle.BottomText.W
	if maxW < 50 {
		maxW = float32(bundle.Layout.ScreenW) - 28
	}
	narrow := bundle.Layout.Tier == TierSmall
	x := bundle.BottomText.X
	y := firstY
	var prevKind ExploreBottomLineKind
	var drew bool
	for _, line := range PlanExploreBottomLines(bundle) {
		s := FormatExploreBottomLineForWidth(hudFace, line, maxW, narrow)
		if strings.TrimSpace(s) == "" {
			continue
		}
		if drew && line.Kind == BottomKindHotkeys && (prevKind == BottomKindZone || prevKind == BottomKindInteraction) {
			y += 8
		}
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		op.ColorScale.ScaleWithColor(exploreBottomLineColor(line.Kind))
		text.Draw(screen, s, hudFace, op)
		y += lineStep
		prevKind = line.Kind
		drew = true
	}
}

// DrawExploreFormationHint — transient-полоса (если есть), затем нижняя панель подсказок.
func DrawExploreFormationHint(screen *ebiten.Image, hudFace *text.GoTextFace, bundle ExploreHUDLayout) {
	if hudFace == nil || bundle.Layout.ScreenH < 40 {
		return
	}
	DrawExploreTransientBanner(screen, hudFace, bundle.Layout, bundle.TransientBannerText, bundle.LineStep)
	firstY, step := DrawExploreHintPanelLayout(screen, bundle)
	DrawExploreFormationHintLines(screen, hudFace, bundle, firstY, step)
}
