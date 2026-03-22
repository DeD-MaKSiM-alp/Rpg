package ui

import "testing"

func TestInspectOverlayPanelRect_positiveBounds(t *testing.T) {
	m := InspectCardModel{
		Title:       "T",
		ContextLine: "ctx",
		HPCur:       10,
		HPMax:       10,
		Alive:       true,
		StatsLine:   "stats",
		Footer:      "f",
	}
	w := InspectPanelWidth(1280, TierMedium, false)
	r, lh := InspectOverlayPanelRect(1280, 720, w, m)
	if r.W <= 0 || r.H <= 0 || lh <= 0 {
		t.Fatalf("non-positive rect/lineH: rect=%+v lh=%v", r, lh)
	}
	if r.X < 0 || r.Y < 0 {
		t.Fatalf("negative origin %+v", r)
	}
	if r.X+r.W > 1280+1 || r.Y+r.H > 720+1 {
		t.Fatalf("panel outside screen %+v", r)
	}
}

func TestInspectContentLineH_tierOrdering(t *testing.T) {
	s := InspectContentLineH(TierSmall)
	l := InspectContentLineH(TierLarge)
	if s >= l {
		t.Fatalf("small lineH should be < large: %v vs %v", s, l)
	}
}

func TestInspectPanelWidth_battleWideVsNarrow(t *testing.T) {
	n := InspectPanelWidth(1280, TierMedium, false)
	w := InspectPanelWidth(1280, TierMedium, true)
	if w < n {
		t.Fatalf("battle-wide should be >= narrow: narrow=%v wide=%v", n, w)
	}
}
