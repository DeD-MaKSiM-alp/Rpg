package ui

import (
	"strings"

	text "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ExploreBottomLine — одна строка нижней панели после overflow (порядок = порядок отрисовки).
type ExploreBottomLine struct {
	Kind ExploreBottomLineKind
	Text string
}

const (
	exploreHotkeysFull    = "R — короткий отдых (без лечения ОЗ) · F5 — лагерь и состав · F9 — наём"
	exploreHotkeysCompact = "R отдых · F5 лагерь · F9 наём"
)

// PlanExploreBottomLines возвращает строки нижней панели в порядке отрисовки:
// зона → взаимодействие → хоткеи → временные баннеры (rest, recruit, POI).
// Содержимое уже прошло ApplyExploreHintOverflow в BuildExploreHUDLayout.
func PlanExploreBottomLines(bundle ExploreHUDLayout) []ExploreBottomLine {
	var out []ExploreBottomLine
	if strings.TrimSpace(bundle.ZoneLine) != "" {
		out = append(out, ExploreBottomLine{Kind: BottomKindZone, Text: strings.TrimSpace(bundle.ZoneLine)})
	}
	if strings.TrimSpace(bundle.InteractionHint) != "" {
		out = append(out, ExploreBottomLine{Kind: BottomKindInteraction, Text: strings.TrimSpace(bundle.InteractionHint)})
	}
	out = append(out, ExploreBottomLine{Kind: BottomKindHotkeys, Text: exploreHotkeysFull})
	if bundle.RestFeedback != "" {
		out = append(out, ExploreBottomLine{Kind: BottomKindBannerRest, Text: bundle.RestFeedback})
	}
	if bundle.RecruitFeedback != "" {
		out = append(out, ExploreBottomLine{Kind: BottomKindBannerRecruit, Text: bundle.RecruitFeedback})
	}
	if strings.TrimSpace(bundle.POIFeedback) != "" {
		out = append(out, ExploreBottomLine{Kind: BottomKindBannerPOI, Text: strings.TrimSpace(bundle.POIFeedback)})
	}
	return out
}

// FormatExploreBottomLineForWidth подбирает полный/компактный текст для baseline-строк и обрезает по ширине.
func FormatExploreBottomLineForWidth(face *text.GoTextFace, line ExploreBottomLine, maxW float32, narrow bool) string {
	switch line.Kind {
	case BottomKindHotkeys:
		return CompactLine(face, exploreHotkeysFull, exploreHotkeysCompact, maxW)
	case BottomKindInteraction:
		return SecondaryLine(face, line.Text, maxW, narrow)
	default:
		return PrimaryLine(face, line.Text, maxW)
	}
}
