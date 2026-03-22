package ui

import "testing"

func TestComputePostBattleLayout_panelOnScreen(t *testing.T) {
	l := ComputePostBattleLayout(1280, 720, false, 0, 2)
	if l.PanelW <= 0 || l.PanelH <= 0 {
		t.Fatalf("empty panel %+v", l)
	}
	if l.PanelX < 0 || l.PanelY < 0 {
		t.Fatalf("negative origin %+v", l)
	}
	if l.PanelX+l.PanelW > float32(l.ScreenW)+1 {
		t.Fatalf("panel past right edge")
	}
	if l.PanelY+l.PanelH > float32(l.ScreenH)+1 {
		t.Fatalf("panel past bottom edge")
	}
}

func TestComputePostBattleLayout_resultButtonInsideInner(t *testing.T) {
	l := ComputePostBattleLayout(1280, 720, false, 0, 1)
	b := l.ResultContinueButton
	if b.X < l.InnerX-0.5 || b.X+b.W > l.InnerX+l.InnerW+0.5 {
		t.Fatalf("continue button not in inner width: innerX=%v innerW=%v btn=%+v", l.InnerX, l.InnerW, b)
	}
	if b.Y < l.InnerY-0.5 || b.Y+b.H > l.PanelY+l.PanelH+0.5 {
		t.Fatalf("continue button vertical out of panel: btn=%+v panelY=%v panelH=%v", b, l.PanelY, l.PanelH)
	}
}

func TestComputePostBattleLayout_rewardOptionRectsSequential(t *testing.T) {
	l := ComputePostBattleLayout(1280, 720, true, 3, 0)
	if len(l.RewardOptionRects) != 3 {
		t.Fatalf("want 3 option rects, got %d", len(l.RewardOptionRects))
	}
	for i := 1; i < len(l.RewardOptionRects); i++ {
		prev, cur := l.RewardOptionRects[i-1], l.RewardOptionRects[i]
		// Ряды сдвигаются с шагом rowH+rowGap; прямоугольники клика расширены (Y-2) и могут слегка пересекаться — важен монотонный порядок.
		if cur.Y <= prev.Y {
			t.Fatalf("option %d should be below previous: prev=%+v cur=%+v", i, prev, cur)
		}
	}
}

func TestComputePostBattleLayout_rewardOptionIndexAt_centerHits(t *testing.T) {
	l := ComputePostBattleLayout(1280, 720, true, 2, 0)
	for i, r := range l.RewardOptionRects {
		mx := int(r.X + r.W/2)
		my := int(r.Y + r.H/2)
		if got := l.RewardOptionIndexAt(mx, my); got != i {
			t.Fatalf("center of option %d: want index %d, got %d", i, i, got)
		}
	}
}

func TestComputePostBattleLayout_tierSmallVsLargeLineH(t *testing.T) {
	s := ComputePostBattleLayout(800, 600, false, 0, 0)
	l := ComputePostBattleLayout(1920, 1080, false, 0, 0)
	if s.Tier != TierSmall || l.Tier != TierLarge {
		t.Fatalf("unexpected tiers: small layout tier=%v large layout tier=%v", s.Tier, l.Tier)
	}
	if s.LineH >= l.LineH {
		t.Fatalf("small tier lineH should be < large: %v vs %v", s.LineH, l.LineH)
	}
}
