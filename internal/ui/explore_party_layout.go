package ui

import (
	"fmt"
	"strings"

	battlepkg "mygame/internal/battle"
	"mygame/internal/party"
)

// PartyStripLayout — геометрия полоски отряда в LeftPanel (расчёт без отрисовки).
type PartyStripLayout struct {
	Panel   FRect
	LineH   float32
	Pad     float32
	Metrics battlepkg.HUDMetrics
}

// ComputePartyStripLayout считает размер панели и масштаб строки внутри LeftPanel.
func ComputePartyStripLayout(lp FRect, tier ResolutionTier, p *party.Party, promoStrip string) (PartyStripLayout, bool) {
	var out PartyStripLayout
	if p == nil || len(p.Active) == 0 || lp.W <= 8 || lp.H <= 8 {
		return out, false
	}
	pol := ExploreHUDTextPolicyForTier(tier)
	lineH := pol.PartyStripBaseLineH
	if lineH < 12 {
		lineH = 12
	}
	pad := pol.PartyStripPad
	if tier == TierSmall {
		pad = pol.PartyStripPadSmall
	}
	maxW := lp.W
	x := lp.X
	y := lp.Y

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
	panelH := pad*2 + lineH*float32(n+extraLines+extraProg+extraPromo) + float32(n)*pol.PartyRowGap + float32(extraLines)*pol.PartyReserveExtraGap
	if lp.H > 0 && panelH > lp.H {
		scale := lp.H / panelH
		if scale < pol.PartyMinLineScale {
			scale = pol.PartyMinLineScale
		}
		lineH *= scale
		if lineH < pol.PartyMinLineH {
			lineH = pol.PartyMinLineH
		}
		panelH = pad*2 + lineH*float32(n+extraLines+extraProg+extraPromo) + float32(n)*pol.PartyRowGap + float32(extraLines)*pol.PartyReserveExtraGap
	}

	out.Panel = FRect{X: x, Y: y, W: maxW, H: panelH}
	out.LineH = lineH
	out.Pad = pad
	out.Metrics = battlepkg.HUDMetrics{LineH: lineH}
	return out, true
}

// PartyStripTitle — заголовок левой панели отряда (player-facing).
func PartyStripTitle(p *party.Party) string {
	if p == nil {
		return ""
	}
	n := len(p.Active)
	nr := len(p.Reserve)
	if nr > 0 {
		return fmt.Sprintf("ОТРЯД · в бою %d · резерв %d", n, nr)
	}
	return fmt.Sprintf("ОТРЯД · в бою %d", n)
}
