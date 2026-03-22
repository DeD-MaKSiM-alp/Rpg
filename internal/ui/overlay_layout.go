package ui

// BattleOverlayScreenLayout — ScreenLayout для полноэкранных overlay (бой, postbattle, inspect): без нижней полосы explore.
func BattleOverlayScreenLayout(screenW, screenH int) ScreenLayout {
	return ComputeScreenLayout(screenW, screenH, 0)
}

// PostBattleMetrics — плотность post-battle UI по tier (единый контракт draw + ComputePostBattleLayout).
func PostBattleMetrics(tier ResolutionTier) (lineH, pad, rowH, rowGap, buttonH float32) {
	p := presetForTier(tier)
	lineH = 20
	if tier == TierSmall {
		lineH = 18
	}
	if tier == TierLarge {
		lineH = 22
	}
	pad = p.Pad * 1.5
	if pad < 16 {
		pad = 16
	}
	if pad > 28 {
		pad = 28
	}
	rowH = lineH * 1.35
	if rowH < 26 {
		rowH = 26
	}
	rowGap = float32(3)
	if tier == TierLarge {
		rowGap = 5
	}
	buttonH = lineH * 1.55
	if buttonH < 32 {
		buttonH = 32
	}
	return lineH, pad, rowH, rowGap, buttonH
}

// PostBattlePanelMaxWidth — ширина панели post-battle с учётом modal/safe.
func PostBattlePanelMaxWidth(screenW int, tier ResolutionTier) float32 {
	p := presetForTier(tier)
	sw := float32(screenW)
	maxW := sw * p.ModalMaxFracW
	if maxW > sw-24 {
		maxW = sw - 24
	}
	w := float32(400)
	if tier == TierLarge {
		w = 440
	}
	if tier == TierSmall {
		w = 360
	}
	if w > maxW {
		w = maxW
	}
	if w < 220 {
		w = 220
	}
	return w
}

// InspectPanelWidth — ширина inspect-карточки: formation (узкая) или battle (широкая портрет).
func InspectPanelWidth(screenW int, tier ResolutionTier, battleWide bool) float32 {
	sw := float32(screenW)
	p := presetForTier(tier)
	maxW := sw * p.ModalMaxFracW
	if maxW > sw-32 {
		maxW = sw - 32
	}
	var w float32
	if battleWide {
		w = 620
		if tier == TierSmall {
			w = 480
		}
		if tier == TierMedium {
			w = 560
		}
	} else {
		w = inspectCardPanelW
		if tier == TierSmall {
			w = 420
		}
		if tier == TierLarge {
			w = 500
		}
	}
	if w > maxW {
		w = maxW
	}
	if w < 260 {
		w = minF(maxW, 260)
	}
	return w
}

// InspectContentLineH — базовая высота строки текста в inspect-карточке по tier.
func InspectContentLineH(tier ResolutionTier) float32 {
	switch tier {
	case TierSmall:
		return 15
	case TierLarge:
		return 18
	default:
		return uiLineH
	}
}

// InspectOverlayPanelRect — центрирование карточки в modal; при нехватке высоты уменьшает lineH и пересчитывает высоту.
func InspectOverlayPanelRect(screenW, screenH int, panelW float32, m InspectCardModel) (rect FRect, lineH float32) {
	sl := BattleOverlayScreenLayout(screenW, screenH)
	lineH = InspectContentLineH(sl.Tier)
	for iter := 0; iter < 40; iter++ {
		h := estimateInspectCardHeight(m, lineH)
		rect = CenterPanelInModal(sl, panelW, h)
		if h <= rect.H+1 || iter == 39 {
			return rect, lineH
		}
		lineH *= rect.H / h
		if lineH < 10.5 {
			lineH = 10.5
			h = estimateInspectCardHeight(m, lineH)
			rect = CenterPanelInModal(sl, panelW, h)
			return rect, lineH
		}
	}
	return rect, lineH
}
