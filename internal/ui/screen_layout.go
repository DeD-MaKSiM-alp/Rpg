package ui

import "strings"

// ResolutionTier — грубая классификация экрана для paddings/плотности (не путать с окном F6/F7).
type ResolutionTier int

const (
	TierSmall ResolutionTier = iota
	TierMedium
	TierLarge
)

// TierFromScreen выбирает tier по ширине и меньшей стороне (устойчиво к ультрашироким окнам).
func TierFromScreen(screenW, screenH int) ResolutionTier {
	if screenW <= 0 || screenH <= 0 {
		return TierMedium
	}
	m := screenW
	if screenH < m {
		m = screenH
	}
	if m < 720 || screenW < 960 {
		return TierSmall
	}
	if m < 900 || screenW < 1440 {
		return TierMedium
	}
	return TierLarge
}

// ScreenLayoutPreset — численные параметры каркаса для одного tier (без «движка», только числа).
type ScreenLayoutPreset struct {
	Pad              float32
	TopHUDHeight     float32
	LeftPanelMaxW    float32
	Gap              float32 // между колонками left / center
	BottomLineStep   float32
	BottomMaxLines   int     // верхняя граница строк в нижней полосе explore
	BottomChromePad  float32 // паддинг подложки снизу
	LineH            float32 // базовая высота строки HUD
	FormationPanelW  float32 // предпочтительная ширина модалки состава
	ModalMaxFracW    float32 // макс. доля ширины экрана под центрированную модалку (0..1)
	ModalMaxFracH    float32 // макс. доля высоты
}

func presetForTier(t ResolutionTier) ScreenLayoutPreset {
	switch t {
	case TierSmall:
		return ScreenLayoutPreset{
			Pad:             8,
			TopHUDHeight:    84,
			LeftPanelMaxW:   280,
			Gap:             8,
			BottomLineStep:  19,
			BottomMaxLines:  5,
			BottomChromePad: 10,
			LineH:           16,
			FormationPanelW: 480,
			ModalMaxFracW:   0.94,
			ModalMaxFracH:   0.92,
		}
	case TierLarge:
		return ScreenLayoutPreset{
			Pad:             16,
			TopHUDHeight:    96,
			LeftPanelMaxW:   372,
			Gap:             12,
			BottomLineStep:  21,
			BottomMaxLines:  9,
			BottomChromePad: 10,
			LineH:           18,
			FormationPanelW: 600,
			ModalMaxFracW:   0.88,
			ModalMaxFracH:   0.88,
		}
	default: // TierMedium
		return ScreenLayoutPreset{
			Pad:             12,
			TopHUDHeight:    88,
			LeftPanelMaxW:   336,
			Gap:             10,
			BottomLineStep:  20,
			BottomMaxLines:  7,
			BottomChromePad: 10,
			LineH:           18,
			FormationPanelW: 560,
			ModalMaxFracW:   0.90,
			ModalMaxFracH:   0.90,
		}
	}
}

// ScreenLayout — единый каркас зон в экранных пикселях. Прямоугольники не обязаны быть непересекающимися
// там, где слои намеренно стыкуются (например TopHUD и LeftPanel — разные слоты по вертикали).
type ScreenLayout struct {
	ScreenW, ScreenH int
	Tier             ResolutionTier
	Preset           ScreenLayoutPreset

	// Safe — внутренняя область с отступами от краёв (контент-фрейм).
	Safe FRect
	// TopHUD — верхняя полоса под статический HUD (предметы, прогресс, повышение).
	TopHUD FRect
	// LeftPanel — левая колонка под explore party strip (под TopHUD).
	LeftPanel FRect
	// CenterStage — центр без левой колонки, между верхом и зарезервированным низом.
	CenterStage FRect
	// BottomBar — зарезервированная полоса снизу под help bar (максимальная высота при полном тексте).
	BottomBar FRect
	// Modal — область для центрирования модалок (внутри Safe, не заезжая на BottomBar).
	Modal FRect
	// TransientBanner — узкая полоса над нижней полосой (баннеры/тосты); может быть нулевой.
	TransientBanner FRect
}

// ComputeScreenLayout строит зоны из размера окна. bottomBarH — фактическая или максимальная высота нижней полосы explore.
func ComputeScreenLayout(screenW, screenH int, bottomBarH float32) ScreenLayout {
	var out ScreenLayout
	out.ScreenW, out.ScreenH = screenW, screenH
	if screenW <= 0 || screenH <= 0 {
		return out
	}
	t := TierFromScreen(screenW, screenH)
	p := presetForTier(t)
	out.Tier = t
	out.Preset = p

	sw := float32(screenW)
	sh := float32(screenH)

	out.Safe = FRect{X: p.Pad, Y: p.Pad, W: sw - 2*p.Pad, H: sh - 2*p.Pad}

	topH := p.TopHUDHeight
	if topH > out.Safe.H*0.45 {
		topH = out.Safe.H * 0.45
	}

	bbh := bottomBarH
	if bbh < 0 {
		bbh = 0
	}
	maxBottom := sh * 0.42
	if bbh > maxBottom {
		bbh = maxBottom
	}

	leftW := p.LeftPanelMaxW
	if leftW > out.Safe.W-p.Gap-80 {
		leftW = out.Safe.W - p.Gap - 80
	}
	if leftW < 120 {
		leftW = 120
	}

	// TopHUD: верх safe-фрейма.
	out.TopHUD = FRect{
		X: out.Safe.X,
		Y: out.Safe.Y,
		W: out.Safe.W,
		H: topH,
	}

	// Низ: full-bleed по ширине экрана (как текущий explore bar).
	out.BottomBar = FRect{X: 0, Y: sh - bbh, W: sw, H: bbh}

	// Левая колонка: под TopHUD, до верхней границы нижней полосы.
	leftTop := out.TopHUD.Y + out.TopHUD.H + p.Gap
	leftBot := out.BottomBar.Y - p.Gap
	if leftBot < leftTop {
		leftBot = leftTop
	}
	out.LeftPanel = FRect{
		X: out.Safe.X,
		Y: leftTop,
		W: leftW,
		H: leftBot - leftTop,
	}

	centerLeft := out.LeftPanel.X + out.LeftPanel.W + p.Gap
	centerW := out.Safe.X + out.Safe.W - centerLeft
	if centerW < 0 {
		centerW = 0
	}
	out.CenterStage = FRect{
		X: centerLeft,
		Y: leftTop,
		W: centerW,
		H: leftBot - leftTop,
	}

	// Modal: центр экрана внутри Safe, выше нижней полосы.
	modalTop := out.Safe.Y
	modalH := out.BottomBar.Y - p.Gap - modalTop
	if modalH < 0 {
		modalH = 0
	}
	out.Modal = FRect{X: out.Safe.X, Y: modalTop, W: out.Safe.W, H: modalH}

	// Баннер: тонкая полоса прямо над bottom bar.
	bannerH := float32(0)
	if bbh > 0 && p.BottomChromePad > 0 {
		bannerH = minF(28, bbh*0.35)
	}
	if bannerH > 0 && out.BottomBar.Y > bannerH+p.Gap {
		out.TransientBanner = FRect{
			X: 0,
			Y: out.BottomBar.Y - bannerH - 2,
			W: sw,
			H: bannerH,
		}
	}

	return out
}

// ExploreBottomBarHeight считает высоту нижней полосы explore по тем же правилам, что и отрисовка.
func ExploreBottomBarHeight(screenW, screenH int, lineStep float32, chromePad float32, zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint string) float32 {
	_ = screenW
	_ = screenH
	n := exploreHintLineCount(zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint)
	return float32(n)*lineStep + chromePad*2
}

func exploreHintLineCount(zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint string) int {
	n := 1 // строка хоткеев (R · F5 · F9)
	if strings.TrimSpace(zoneLine) != "" {
		n++
	}
	if strings.TrimSpace(interactionHint) != "" {
		n++
	}
	if strings.TrimSpace(restFeedback) != "" {
		n++
	}
	if strings.TrimSpace(recruitFeedback) != "" {
		n++
	}
	if strings.TrimSpace(poiFeedback) != "" {
		n++
	}
	return n
}

// ExploreHUDLayout / BuildExploreHUDLayout / BuildExploreLayoutBundle — см. explore_hud_layout.go (единый расчёт explore HUD).

// CenterPanelInModal размещает прямоугольник panelW×panelH: ширина — в safe/modal, высота — до доли экрана
// (не ужимается до lay.Modal.H, чтобы длинные списки могли масштабироваться вызывающим кодом).
func CenterPanelInModal(lay ScreenLayout, panelW, panelH float32) FRect {
	p := lay.Preset
	sw := float32(lay.ScreenW)
	sh := float32(lay.ScreenH)

	maxW := sw * p.ModalMaxFracW
	if panelW > maxW {
		panelW = maxW
	}
	if panelW > lay.Modal.W {
		panelW = lay.Modal.W
	}
	maxH := sh * p.ModalMaxFracH
	if maxH > sh-8 {
		maxH = sh - 8
	}
	if panelH > maxH {
		panelH = maxH
	}
	px := lay.Modal.X + (lay.Modal.W-panelW)*0.5
	py := (sh - panelH) * 0.5
	if py < 4 {
		py = 4
	}
	if px < 0 {
		px = 0
	}
	return FRect{X: px, Y: py, W: panelW, H: panelH}
}

// FormationPanelBaseSize — ширина модалки состава и горизонтальные отступы (Y стартует под заголовком overlay).
func FormationPanelBaseSize(screenW int, tier ResolutionTier) (panelW, pad float32) {
	p := presetForTier(tier)
	sw := float32(screenW)
	pad = p.Pad * 1.2
	panelW = p.FormationPanelW
	if sw-pad*2 < panelW {
		panelW = sw - pad*2
	}
	return panelW, pad
}
