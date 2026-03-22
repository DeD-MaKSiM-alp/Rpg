package ui

import (
	"testing"

	battlepkg "mygame/internal/battle"
)

func TestBattleContentRectMatchesScreenSafe(t *testing.T) {
	cases := []struct{ w, h int }{
		{800, 600},
		{1280, 720},
		{1920, 1080},
	}
	for _, c := range cases {
		lay := ComputeScreenLayout(c.w, c.h, 0)
		cx, cy, cw, ch := battlepkg.BattleContentRectForHUDAnchor(c.w, c.h)
		if float32(cx) != lay.Safe.X || float32(cy) != lay.Safe.Y {
			t.Fatalf("%dx%d: origin mismatch ui.Safe=%+v battle=(%v,%v)", c.w, c.h, lay.Safe, cx, cy)
		}
		if float32(cw) != lay.Safe.W || float32(ch) != lay.Safe.H {
			t.Fatalf("%dx%d: size mismatch ui.Safe=%+v battle=(%v,%v)", c.w, c.h, lay.Safe, cw, ch)
		}
	}
}

func TestBattleHUDTierOrdinalMatchesUI(t *testing.T) {
	ports := []struct{ w, h int }{
		{800, 600},
		{1280, 720},
		{1920, 1080},
	}
	for _, p := range ports {
		if got, want := battlepkg.BattleHUDTierOrdinal(p.w, p.h), int(TierFromScreen(p.w, p.h)); got != want {
			t.Fatalf("%dx%d: battle tier ordinal=%d ui.Tier=%d", p.w, p.h, got, want)
		}
	}
}
