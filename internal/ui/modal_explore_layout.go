package ui

import "mygame/world/entity"

// FRect — float32 rect for modal hit-testing (screen pixels).
type FRect struct {
	X, Y, W, H float32
}

func (r FRect) Contains(px, py int) bool {
	if r.W <= 0 || r.H <= 0 {
		return false
	}
	fx, fy := float32(px), float32(py)
	return fx >= r.X && fx <= r.X+r.W && fy >= r.Y && fy <= r.Y+r.H
}

// RecruitOfferLayout — shared geometry for DrawRecruitOfferOverlay + hit-test.
type RecruitOfferLayout struct {
	ScreenW, ScreenH int
	Panel            FRect
	AcceptBtn        FRect
	DeclineBtn       FRect
}

// LayoutRecruitOffer computes panel and button rects (must match DrawRecruitOfferOverlay).
func LayoutRecruitOffer(screenW, screenH int) RecruitOfferLayout {
	var out RecruitOfferLayout
	out.ScreenW, out.ScreenH = screenW, screenH
	if screenW < 100 || screenH < 100 {
		return out
	}
	w := float32(screenW)
	h := float32(screenH)
	panelW := float32(440)
	if panelW > w-40 {
		panelW = w - 40
	}
	lineH := float32(20)
	btnH := float32(38)
	btnGap := float32(12)
	innerPad := float32(20)
	// Must match DrawRecruitOfferOverlay: py+18, then +lineH+14, +lineH+10, +lineH+12 before buttons
	topToBtn := float32(18) + lineH + 14 + lineH + 10 + lineH + 12
	panelH := topToBtn + btnH + 10 + lineH*1.1 + 22
	px := (w - panelW) / 2
	py := (h - panelH) / 2
	out.Panel = FRect{X: px, Y: py, W: panelW, H: panelH}

	innerW := panelW - innerPad*2
	btnW := (innerW - btnGap) / 2
	bx := px + innerPad
	by := py + topToBtn
	out.AcceptBtn = FRect{X: bx, Y: by, W: btnW, H: btnH}
	out.DeclineBtn = FRect{X: bx + btnW + btnGap, Y: by, W: btnW, H: btnH}
	return out
}

// RecruitOfferHit — result of pointer hit-test.
type RecruitOfferHit int

const (
	RecruitHitNone RecruitOfferHit = iota
	RecruitHitBackdrop
	RecruitHitAccept
	RecruitHitDecline
)

// HitTestRecruitOffer returns what the cursor is over (backdrop = outside panel).
func HitTestRecruitOffer(mx, my int, screenW, screenH int) RecruitOfferHit {
	lay := LayoutRecruitOffer(screenW, screenH)
	if lay.Panel.W <= 0 {
		return RecruitHitNone
	}
	if lay.AcceptBtn.Contains(mx, my) {
		return RecruitHitAccept
	}
	if lay.DeclineBtn.Contains(mx, my) {
		return RecruitHitDecline
	}
	if lay.Panel.Contains(mx, my) {
		return RecruitHitNone
	}
	return RecruitHitBackdrop
}

// POIChoiceLayout — geometry for POI ruins/altar modal.
type POIChoiceLayout struct {
	ScreenW, ScreenH int
	Panel            FRect
	Option0          FRect
	Option1          FRect
	ConfirmBtn       FRect
	CancelZone       FRect // text/area for "leave"
}

// LayoutPOIChoice must stay in sync with DrawPOIChoiceOverlay.
func LayoutPOIChoice(screenW, screenH int, kind entity.PickupKind) POIChoiceLayout {
	var out POIChoiceLayout
	out.ScreenW, out.ScreenH = screenW, screenH
	if screenW < 100 || screenH < 100 {
		return out
	}
	switch kind {
	case entity.PickupKindPOIRuins, entity.PickupKindPOIAltar:
	default:
		return out
	}
	w := float32(screenW)
	h := float32(screenH)
	panelW := float32(480)
	if panelW > w-40 {
		panelW = w - 40
	}
	lineH := float32(18)
	rowBlock := lineH*2 + 20
	rowGap := float32(10)
	btnH := float32(36)
	footerH := lineH*1.15 + 8
	// Rows start at same y as after title + «y += lineH+12» in DrawPOIChoiceOverlay.
	row0Top := float32(14) + lineH + 12
	panelH := row0Top + rowBlock*2 + rowGap + btnH + 12 + footerH + 22
	px := (w - panelW) / 2
	py := (h - panelH) / 2
	out.Panel = FRect{X: px, Y: py, W: panelW, H: panelH}

	innerX := px + 16
	innerW := panelW - 32
	y0 := py + row0Top
	out.Option0 = FRect{X: innerX, Y: y0, W: innerW, H: rowBlock}
	out.Option1 = FRect{X: innerX, Y: y0 + rowBlock + rowGap, W: innerW, H: rowBlock}

	by := y0 + rowBlock*2 + rowGap + 10
	out.ConfirmBtn = FRect{X: innerX, Y: by, W: innerW, H: btnH}

	// Cancel: bottom strip (footer line area) — thin hit for "уйти"
	out.CancelZone = FRect{X: innerX, Y: by + btnH + 6, W: innerW, H: footerH + 4}
	return out
}

// POIChoiceHit identifies interactive regions.
type POIChoiceHit int

const (
	POIHitNone POIChoiceHit = iota
	POIHitBackdrop
	POIHitOption0
	POIHitOption1
	POIHitConfirm
	POIHitCancel
)

// HitTestPOIChoice — mx,my in screen pixels.
func HitTestPOIChoice(mx, my int, screenW, screenH int, kind entity.PickupKind) POIChoiceHit {
	lay := LayoutPOIChoice(screenW, screenH, kind)
	if lay.Panel.W <= 0 {
		return POIHitNone
	}
	if lay.Option0.Contains(mx, my) {
		return POIHitOption0
	}
	if lay.Option1.Contains(mx, my) {
		return POIHitOption1
	}
	if lay.ConfirmBtn.Contains(mx, my) {
		return POIHitConfirm
	}
	if lay.CancelZone.Contains(mx, my) {
		return POIHitCancel
	}
	if lay.Panel.Contains(mx, my) {
		return POIHitNone
	}
	return POIHitBackdrop
}
