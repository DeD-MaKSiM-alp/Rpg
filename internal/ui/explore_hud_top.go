package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ExploreTopComposition — явная композиция верхнего explore HUD (без отрисовки).
// Panel == TopHUD; Background — подложка под мета-блок; Text — зона текста (inset).
type ExploreTopComposition struct {
	Panel          FRect
	Background     FRect
	Text           FRect
	MetaLineStep   float64
	NumLines       int
}

// FinalizeExploreHUDTopComposition заполняет поля верхнего слоя в layout (прогресс лидера — в левой панели отряда).
func FinalizeExploreHUDTopComposition(hud ExploreHUDLayout, promotionLine string) ExploreHUDLayout {
	pol := ExploreHUDTextPolicyForTier(hud.Layout.Tier)
	comp := computeExploreTopComposition(hud.Layout.TopHUD, hud.Layout.Preset, pol, promotionLine)
	hud.TopPanel = comp.Panel
	hud.TopBackground = comp.Background
	hud.TopText = comp.Text
	hud.TopMetaLineStep = comp.MetaLineStep
	hud.TopMetaLines = comp.NumLines
	return hud
}

func computeExploreTopComposition(topHUD FRect, preset ScreenLayoutPreset, pol ExploreHUDTextPolicy, promotionLine string) ExploreTopComposition {
	nLines := 1 // ресурсы (предметы + знаки) одной строкой
	if strings.TrimSpace(promotionLine) != "" {
		nLines++
	}
	lineH := float32(preset.LineH) + pol.TopLineGapExtra
	if lineH < 12 {
		lineH = 12
	}
	pad := pol.TopBackgroundPad
	hdr := pol.TopStatusHeaderH
	bgH := pad + hdr + 4 + lineH*float32(nLines) + pad
	if topHUD.H > 0 && bgH > topHUD.H {
		bgH = topHUD.H
	}
	bg := FRect{X: topHUD.X, Y: topHUD.Y, W: topHUD.W, H: bgH}
	textR := exploreTopContentRect(topHUD, pol, lineH, nLines, bgH)
	return ExploreTopComposition{
		Panel:        topHUD,
		Background:   bg,
		Text:         textR,
		MetaLineStep: float64(lineH),
		NumLines:     nLines,
	}
}

func exploreTopContentRect(topHUD FRect, pol ExploreHUDTextPolicy, lineH float32, nLines int, bgH float32) FRect {
	pad := pol.TopBackgroundPad
	hdr := pol.TopStatusHeaderH
	y0 := topHUD.Y + pad + hdr + 4
	h := lineH*float32(nLines) + 6
	if bgH > 0 {
		maxH := topHUD.Y + bgH - y0 - pad
		if maxH > 0 && h > maxH {
			h = maxH
		}
	}
	r := FRect{
		X: topHUD.X + pol.TopContentPadX,
		Y: y0,
		W: topHUD.W - 2*pol.TopContentPadX,
		H: h,
	}
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}

// DrawExploreTopBackground рисует подложку верхнего HUD по уже рассчитанной композиции.
func DrawExploreTopBackground(screen *ebiten.Image, hudFace *text.GoTextFace, hud ExploreHUDLayout) {
	bg := hud.TopBackground
	if bg.W <= 0 || bg.H <= 0 {
		return
	}
	pol := ExploreHUDTextPolicyForTier(hud.Layout.Tier)
	pad := pol.TopBackgroundPad
	hdr := pol.TopStatusHeaderH

	bgCol := Theme.ExploreStatusBG
	bgCol.A = uint8(polTopBackgroundAlpha(hud.Layout.Tier))
	vector.FillRect(screen, bg.X, bg.Y, bg.W, bg.H, bgCol, false)
	vector.FillRect(screen, bg.X, bg.Y, 4, bg.H, Theme.AccentStrip, false)
	vector.FillRect(screen, bg.X, bg.Y+bg.H-1, bg.W, 1, Theme.ExploreModuleEdge, false)

	// Шапка «статусной» панели
	if hdr > 1 && pad*2+hdr < bg.H {
		vector.FillRect(screen, bg.X+pad, bg.Y+pad, bg.W-pad*2, hdr, Theme.PanelBGDeep, false)
		DrawThinAccentLine(screen, bg.X+6, bg.Y+pad+2, bg.W-12)
		if hudFace != nil {
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(bg.X+pad+10), float64(bg.Y+pad+3))
			op.ColorScale.ScaleWithColor(Theme.TextMuted)
			text.Draw(screen, exploreTopStatusTitle, hudFace, op)
		}
		sepY := bg.Y + pad + hdr
		vector.StrokeLine(screen, bg.X+pad, sepY, bg.X+bg.W-pad, sepY, 1, Theme.PanelTitleSep, false)
	}
}

const exploreTopStatusTitle = "СТАТУС"

func polTopBackgroundAlpha(tier ResolutionTier) int {
	switch tier {
	case TierSmall:
		return 210
	default:
		return 215
	}
}
