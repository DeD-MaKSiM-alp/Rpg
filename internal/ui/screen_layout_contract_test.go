package ui

import "testing"

func TestComputeScreenLayout_safeInsideScreen(t *testing.T) {
	sw, sh := 1280, 720
	lay := ComputeScreenLayout(sw, sh, 120)
	s := lay.Safe
	if s.X < 0 || s.Y < 0 || s.W <= 0 || s.H <= 0 {
		t.Fatalf("invalid Safe: %+v", s)
	}
	if s.X+s.W > float32(sw)+0.5 || s.Y+s.H > float32(sh)+0.5 {
		t.Fatalf("Safe exceeds screen: Safe=%+v screen=%dx%d", s, sw, sh)
	}
}

func TestComputeScreenLayout_modalWithinSafeWidth(t *testing.T) {
	sw, sh := 960, 720
	lay := ComputeScreenLayout(sw, sh, 0)
	if lay.Modal.X < lay.Safe.X-0.5 {
		t.Fatalf("Modal.X left of Safe: modal=%+v safe=%+v", lay.Modal, lay.Safe)
	}
	if lay.Modal.X+lay.Modal.W > lay.Safe.X+lay.Safe.W+0.5 {
		t.Fatalf("Modal wider than Safe: modal=%+v safe=%+v", lay.Modal, lay.Safe)
	}
}

func TestComputeScreenLayout_leftPanelAboveBottomBar(t *testing.T) {
	lay := ComputeScreenLayout(1280, 720, 100)
	if lay.LeftPanel.Y+lay.LeftPanel.H > lay.BottomBar.Y+0.5 {
		t.Fatalf("left panel should end at or above bottom bar: leftBot=%v bottomY=%v",
			lay.LeftPanel.Y+lay.LeftPanel.H, lay.BottomBar.Y)
	}
}

func TestComputeScreenLayout_topHUDDoesNotOverlapLeftPanelBody(t *testing.T) {
	lay := ComputeScreenLayout(1280, 720, 80)
	if lay.LeftPanel.Y < lay.TopHUD.Y+lay.TopHUD.H-0.1 {
		t.Fatalf("LeftPanel should start below TopHUD+gap: top=%+v leftY=%v", lay.TopHUD, lay.LeftPanel.Y)
	}
}

func TestCenterPanelInModal_fitsModalAndScreen(t *testing.T) {
	sl := ComputeScreenLayout(1280, 720, 0)
	r := CenterPanelInModal(sl, 900, 800)
	if r.W <= 0 || r.H <= 0 {
		t.Fatalf("empty panel rect %+v", r)
	}
	if r.X+r.W > float32(sl.ScreenW)+1 {
		t.Fatalf("panel beyond screen width: %+v", r)
	}
	if r.Y+r.H > float32(sl.ScreenH)+1 {
		t.Fatalf("panel beyond screen height: %+v", r)
	}
}

func TestTierFromScreen_distinctSizes(t *testing.T) {
	if TierFromScreen(800, 600) == TierFromScreen(1920, 1080) {
		t.Fatal("800x600 and 1920x1080 should not map to same tier")
	}
	if TierFromScreen(1280, 720) == TierFromScreen(800, 600) {
		t.Fatal("1280x720 should not equal small tier")
	}
}
