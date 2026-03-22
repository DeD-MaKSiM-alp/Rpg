package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// combineTransientFeedback склеивает непустые баннеры отдыха/рекрута/POI в одну строку для полосы над низом.
// Порядок: POI → рекрут → отдых (более «редкие» события ближе к началу строки).
func combineTransientFeedback(rest, recruit, poi string) string {
	var parts []string
	if s := strings.TrimSpace(poi); s != "" {
		parts = append(parts, s)
	}
	if s := strings.TrimSpace(recruit); s != "" {
		parts = append(parts, s)
	}
	if s := strings.TrimSpace(rest); s != "" {
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " · ")
}

// transientBannerUsable — можно ли использовать зарезервированный TransientBanner из ScreenLayout.
func transientBannerUsable(lay ScreenLayout) bool {
	b := lay.TransientBanner
	return b.W > 8 && b.H > 6
}

// DrawExploreTransientBanner рисует временные сообщения над нижней полосой (player-facing, не debug).
func DrawExploreTransientBanner(screen *ebiten.Image, hudFace *text.GoTextFace, lay ScreenLayout, textLine string, lineStep float32) {
	if hudFace == nil || strings.TrimSpace(textLine) == "" || !transientBannerUsable(lay) {
		return
	}
	tb := lay.TransientBanner
	pol := ExploreHUDTextPolicyForTier(lay.Tier)
	padX := pol.TransientBannerPadX
	vector.FillRect(screen, tb.X, tb.Y, tb.W, tb.H, Theme.PanelBGDeep, false)
	vector.FillRect(screen, tb.X, tb.Y, 4, tb.H, Theme.AccentStrip, false)
	vector.StrokeRect(screen, tb.X, tb.Y, tb.W, tb.H, 1, Theme.PostBattleBorder, false)
	if lineStep < 16 {
		lineStep = 16
	}
	maxW := tb.W - padX*2
	line := PrimaryLine(hudFace, strings.TrimSpace(textLine), maxW)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(tb.X+padX), float64(tb.Y+(tb.H-lineStep)*0.5))
	op.ColorScale.ScaleWithColor(Theme.TextPrimary)
	text.Draw(screen, line, hudFace, op)
}
