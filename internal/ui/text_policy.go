package ui

import (
	"strings"

	text "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// TextOverflowPolicy задаёт осознанное поведение при нехватке ширины/высоты.
type TextOverflowPolicy int

const (
	// OverflowTruncate — одна строка, обрезка с многоточием (через trimTextToWidth).
	OverflowTruncate TextOverflowPolicy = iota
	// OverflowOmit — не рисовать, если не помещается по политике «сначала вторичное».
	OverflowOmit
)

// TextLineRole — слой текста для иерархии подсказок/HUD.
type TextLineRole int

const (
	LinePrimary TextLineRole = iota
	LineSecondary
	LineCompact
)

// PrimaryLine обрезает главную строку под maxW.
func PrimaryLine(face *text.GoTextFace, s string, maxW float32) string {
	return trimTextToWidth(face, strings.TrimSpace(s), maxW)
}

// SecondaryLine — вторичная строка; при очень узкой области можно вернуть пусто.
func SecondaryLine(face *text.GoTextFace, s string, maxW float32, narrow bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if narrow && maxW < 200 {
		return ""
	}
	return trimTextToWidth(face, s, maxW)
}

// CompactLine выбирает короткий вариант, если полный текст не помещается по ширине.
func CompactLine(face *text.GoTextFace, full, compact string, maxW float32) string {
	full = strings.TrimSpace(full)
	compact = strings.TrimSpace(compact)
	if compact != "" && measureTextWidth(face, full) > maxW {
		return trimTextToWidth(face, compact, maxW)
	}
	return trimTextToWidth(face, full, maxW)
}

// ApplyExploreHintOverflow урезает опциональные строки, если их больше, чем MaxBottomLines у tier.
// Порядок вытеснения должен совпадать с ExploreBottomEvictionPriority() (explore_hud_policy.go):
// POI → recruit → rest → interaction → zone. Базовая строка хоткеев не вытесняется (см. ExploreBottomBaselineSlotCount).
func ApplyExploreHintOverflow(tier ResolutionTier, zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint string) (z, rest, rec, poi, inter string) {
	maxL := presetForTier(tier).BottomMaxLines
	maxOptional := maxL - ExploreBottomBaselineSlotCount()
	if maxOptional < 0 {
		maxOptional = 0
	}

	z = strings.TrimSpace(zoneLine)
	inter = strings.TrimSpace(interactionHint)
	rest = strings.TrimSpace(restFeedback)
	rec = strings.TrimSpace(recruitFeedback)
	poi = strings.TrimSpace(poiFeedback)

	optional := []string{z, inter, rest, rec, poi}
	// Приоритет отбрасывания: сначала poi, recruit, rest, interaction, zone (см. порядок ниже).
	for countOptional(optional) > maxOptional {
		if poi != "" {
			poi = ""
			optional[4] = ""
			continue
		}
		if rec != "" {
			rec = ""
			optional[3] = ""
			continue
		}
		if rest != "" {
			rest = ""
			optional[2] = ""
			continue
		}
		if inter != "" {
			inter = ""
			optional[1] = ""
			continue
		}
		if z != "" {
			z = ""
			optional[0] = ""
			continue
		}
		break
	}
	return z, rest, rec, poi, inter
}

func countOptional(parts []string) int {
	n := 0
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			n++
		}
	}
	return n
}
