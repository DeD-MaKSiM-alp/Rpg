package ui

import "testing"

func TestTierFromScreen(t *testing.T) {
	if TierFromScreen(800, 600) != TierSmall {
		t.Fatalf("800x600 small")
	}
	if TierFromScreen(1280, 720) != TierMedium {
		t.Fatalf("1280x720 medium")
	}
	if TierFromScreen(1920, 1080) != TierLarge {
		t.Fatalf("1920x1080 large")
	}
}

func TestBuildExploreLayoutBundle_nonOverlap(t *testing.T) {
	b := BuildExploreLayoutBundle(1280, 720, "zone", "a", "b", "c", "hint")
	if b.Layout.BottomBar.Y+b.Layout.BottomBar.H > float32(b.Layout.ScreenH)+0.5 {
		t.Fatalf("bottom bar beyond screen")
	}
	if b.Layout.LeftPanel.Y+0.1 >= b.Layout.BottomBar.Y {
		t.Fatalf("left panel should sit above bottom bar")
	}
}
