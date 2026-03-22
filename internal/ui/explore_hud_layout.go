package ui

// ExploreHUDLayout — единый результат расчёта геометрии и контента explore HUD (player-facing).
// Поле Layout — полный ScreenLayout (safe, TopHUD, LeftPanel, BottomBar, Modal, TransientBanner, …).
type ExploreHUDLayout struct {
	Layout ScreenLayout

	ZoneLine        string
	RestFeedback    string
	RecruitFeedback string
	POIFeedback     string
	InteractionHint string

	// TransientBannerText — склеенные временные сообщения (отдых/рекрут/POI) для полосы Layout.TransientBanner.
	// Пусто, если сообщения остаются в нижней панели или баннер недоступен по геометрии.
	TransientBannerText string

	// Верх: TopPanel == TopHUD; TopBackground/TopText/TopMetaLineStep заполняются FinalizeExploreHUDTopComposition(hud, promotionLine) перед draw.
	TopPanel        FRect
	TopBackground   FRect
	TopText         FRect
	TopMetaLineStep float64
	TopMetaLines    int

	// BottomPanel — дублирует Layout.BottomBar для явной семантики «панель подсказок».
	BottomPanel FRect
	// BottomText — прямоугольник многострочного текста внутри нижней панели (после chrome pad).
	BottomText FRect
	LineStep   float32
}

// ExploreLayoutBundle — совместимое имя типа (старые вызовы и тесты).
type ExploreLayoutBundle = ExploreHUDLayout

// BuildExploreHUDLayout — единая точка расчёта: overflow строк, высота низа, ComputeScreenLayout, rect’ы контента.
func BuildExploreHUDLayout(screenW, screenH int, zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint string) ExploreHUDLayout {
	tier := TierFromScreen(screenW, screenH)
	p := presetForTier(tier)
	pol := ExploreHUDTextPolicyForTier(tier)
	z, rest, rec, poi, inter := ApplyExploreHintOverflow(tier, zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint)

	transient := combineTransientFeedback(rest, rec, poi)
	restB, recB, poiB := rest, rec, poi
	if transient != "" {
		restB, recB, poiB = "", "", ""
	}

	h := ExploreBottomBarHeight(screenW, screenH, p.BottomLineStep, p.BottomChromePad, z, restB, recB, poiB, inter)
	lay := ComputeScreenLayout(screenW, screenH, h)

	if transient != "" && !transientBannerUsable(lay) {
		transient = ""
		restB, recB, poiB = rest, rec, poi
		h = ExploreBottomBarHeight(screenW, screenH, p.BottomLineStep, p.BottomChromePad, z, restB, recB, poiB, inter)
		lay = ComputeScreenLayout(screenW, screenH, h)
	}

	topText := topTextRectFromTopHUD(lay.TopHUD, pol)
	bottomText := bottomTextRectFromBottomBar(lay.BottomBar, pol)
	lineStep := p.BottomLineStep
	if lineStep < 1 {
		lineStep = 16
	}

	return ExploreHUDLayout{
		Layout:              lay,
		ZoneLine:            z,
		RestFeedback:        restB,
		RecruitFeedback:     recB,
		POIFeedback:         poiB,
		InteractionHint:     inter,
		TransientBannerText: transient,
		TopPanel:  lay.TopHUD,
		TopText:   topText,
		BottomPanel:         lay.BottomBar,
		BottomText:          bottomText,
		LineStep:            lineStep,
	}
}

// NewExploreHUDLayoutFromScreenLayout строит ExploreHUDLayout из уже посчитанного ScreenLayout (режимы без полного explore bundle).
func NewExploreHUDLayoutFromScreenLayout(lay ScreenLayout) ExploreHUDLayout {
	pol := ExploreHUDTextPolicyForTier(lay.Tier)
	topText := topTextRectFromTopHUD(lay.TopHUD, pol)
	bottomText := bottomTextRectFromBottomBar(lay.BottomBar, pol)
	lineStep := lay.Preset.BottomLineStep
	if lineStep < 1 {
		lineStep = 16
	}
	return ExploreHUDLayout{
		Layout:      lay,
		TopPanel:    lay.TopHUD,
		TopText:     topText,
		BottomPanel:    lay.BottomBar,
		BottomText:     bottomText,
		LineStep:       lineStep,
	}
}

func topTextRectFromTopHUD(top FRect, pol ExploreHUDTextPolicy) FRect {
	r := FRect{
		X: top.X + pol.TopContentPadX,
		Y: top.Y + pol.TopContentPadY,
		W: top.W - 2*pol.TopContentPadX,
		H: top.H - 2*pol.TopContentPadY,
	}
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}

func bottomTextRectFromBottomBar(bottomBar FRect, pol ExploreHUDTextPolicy) FRect {
	if bottomBar.W <= 0 || bottomBar.H <= 0 {
		return FRect{}
	}
	r := FRect{
		X: pol.BottomTextInsetX,
		Y: bottomBar.Y + pol.BottomChromePad,
		W: bottomBar.W - 2*pol.BottomTextInsetX,
		H: bottomBar.H - 2*pol.BottomChromePad,
	}
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}

// BuildExploreLayoutBundle совместимость: делегирует в BuildExploreHUDLayout.
func BuildExploreLayoutBundle(screenW, screenH int, zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint string) ExploreHUDLayout {
	return BuildExploreHUDLayout(screenW, screenH, zoneLine, restFeedback, recruitFeedback, poiFeedback, interactionHint)
}
