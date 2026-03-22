package ui

import "testing"

func TestBuildExploreHUDLayout_topTextInsideTopHUD(t *testing.T) {
	h := BuildExploreHUDLayout(1280, 720, "zone", "", "", "", "")
	top := h.Layout.TopHUD
	tt := h.TopText
	if tt.X < top.X || tt.Y < top.Y {
		t.Fatalf("TopText should be inset in TopHUD: top=%+v text=%+v", top, tt)
	}
	if tt.X+tt.W > top.X+top.W+0.5 || tt.Y+tt.H > top.Y+top.H+0.5 {
		t.Fatalf("TopText exceeds TopHUD: top=%+v text=%+v", top, tt)
	}
}

func TestBuildExploreHUDLayout_bottomTextInsideBottomBar(t *testing.T) {
	h := BuildExploreHUDLayout(1280, 720, "z", "a", "b", "c", "hint")
	bb := h.Layout.BottomBar
	bt := h.BottomText
	if bt.W <= 0 || bt.H <= 0 {
		t.Fatalf("expected non-empty BottomText: %+v", bt)
	}
	if bt.X < 0 || bt.Y < bb.Y {
		t.Fatalf("BottomText Y should be inside bottom bar: bb=%+v bt=%+v", bb, bt)
	}
	if bt.X+bt.W > bb.X+bb.W+0.5 || bt.Y+bt.H > bb.Y+bb.H+0.5 {
		t.Fatalf("BottomText exceeds BottomBar: bb=%+v bt=%+v", bb, bt)
	}
}

func TestExploreBottomEvictionPriority_matchesApplyExploreHintOrder(t *testing.T) {
	p := ExploreBottomEvictionPriority()
	if len(p) != 5 {
		t.Fatalf("expected 5 optional kinds, got %d", len(p))
	}
	// Должно совпадать с порядком в ApplyExploreHintOverflow (text_policy.go).
	if p[0] != BottomKindBannerPOI || p[4] != BottomKindZone {
		t.Fatalf("unexpected eviction order: %v", p)
	}
}
