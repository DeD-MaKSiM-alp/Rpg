package ui

import (
	"testing"

	"mygame/world/entity"
)

func TestLayoutRecruitOffer_buttonsInsidePanel(t *testing.T) {
	lay := LayoutRecruitOffer(800, 600)
	if lay.Panel.W <= 0 {
		t.Fatal("empty panel")
	}
	p := lay.Panel
	for _, name := range []struct {
		n string
		r FRect
	}{{"Accept", lay.AcceptBtn}, {"Decline", lay.DeclineBtn}} {
		r := name.r
		if r.X < p.X || r.Y < p.Y {
			t.Fatalf("%s outside panel top-left: panel=%+v btn=%+v", name.n, p, r)
		}
		if r.X+r.W > p.X+p.W+0.5 || r.Y+r.H > p.Y+p.H+0.5 {
			t.Fatalf("%s outside panel bottom-right: panel=%+v btn=%+v", name.n, p, r)
		}
	}
}

func TestLayoutRecruitOffer_panelUsesCenterModal(t *testing.T) {
	lay := LayoutRecruitOffer(1024, 768)
	if lay.Panel.X < 0 || lay.Panel.Y < 4 {
		t.Fatalf("panel should be placed in-screen: %+v", lay.Panel)
	}
}

func TestLayoutPOIChoice_optionsDoNotOverlap(t *testing.T) {
	lay := LayoutPOIChoice(800, 600, entity.PickupKindPOIAltar)
	a, b := lay.Option0, lay.Option1
	if a.Y+a.H > b.Y+0.5 {
		t.Fatalf("option rects overlap vertically: opt0=%+v opt1=%+v", a, b)
	}
}

func TestLayoutPOIChoice_confirmBelowOptions(t *testing.T) {
	lay := LayoutPOIChoice(900, 700, entity.PickupKindPOIRuins)
	if lay.ConfirmBtn.Y < lay.Option1.Y+lay.Option1.H-0.5 {
		t.Fatalf("confirm should be below options: confirmY=%v opt1=%+v", lay.ConfirmBtn.Y, lay.Option1)
	}
}
