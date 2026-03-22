package ui

import (
	"testing"

	"mygame/world/entity"
)

func TestHitTestRecruitOffer_acceptButtonCenter(t *testing.T) {
	lay := LayoutRecruitOffer(800, 600)
	if lay.AcceptBtn.W <= 0 {
		t.Fatal("empty layout")
	}
	cx := int(lay.AcceptBtn.X + lay.AcceptBtn.W/2)
	cy := int(lay.AcceptBtn.Y + lay.AcceptBtn.H/2)
	if got := HitTestRecruitOffer(cx, cy, 800, 600); got != RecruitHitAccept {
		t.Fatalf("center of accept: want %v, got %v", RecruitHitAccept, got)
	}
}

func TestHitTestRecruitOffer_backdrop(t *testing.T) {
	if got := HitTestRecruitOffer(0, 0, 800, 600); got != RecruitHitBackdrop {
		t.Fatalf("corner outside panel: want backdrop, got %v", got)
	}
}

func TestHitTestPOIChoice_confirmButton(t *testing.T) {
	lay := LayoutPOIChoice(800, 600, entity.PickupKindPOIAltar)
	if lay.ConfirmBtn.W <= 0 {
		t.Fatal("empty layout")
	}
	cx := int(lay.ConfirmBtn.X + lay.ConfirmBtn.W/2)
	cy := int(lay.ConfirmBtn.Y + lay.ConfirmBtn.H/2)
	if got := HitTestPOIChoice(cx, cy, 800, 600, entity.PickupKindPOIAltar); got != POIHitConfirm {
		t.Fatalf("want confirm, got %v", got)
	}
}
