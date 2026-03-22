package battle

import "testing"

func TestOffsetBattleHUDLayout_translatesTopBar(t *testing.T) {
	l := BattleHUDLayout{
		TopBar: HUDRect{X: 10, Y: 20, W: 100, H: 30},
	}
	o := OffsetBattleHUDLayout(l, 5, -7)
	if o.TopBar.X != 15 || o.TopBar.Y != 13 {
		t.Fatalf("unexpected offset: %+v", o.TopBar)
	}
}

func TestComputeBattleHUDLayoutAnchored_contentWithinScreen(t *testing.T) {
	b := &BattleContext{}
	b.LayoutStyle = LayoutStyleV1Table
	sw, sh := 1280, 720
	lay := b.ComputeBattleHUDLayoutAnchored(sw, sh)
	if lay.Content.W <= 0 || lay.Content.H <= 0 {
		t.Fatalf("empty content rect %+v", lay.Content)
	}
	if lay.Content.X < -0.5 || lay.Content.Y < -0.5 {
		t.Fatalf("negative content origin %+v", lay.Content)
	}
	if lay.Content.X+lay.Content.W > float32(sw)+1 {
		t.Fatalf("content past screen width %+v", lay.Content)
	}
	if lay.Content.Y+lay.Content.H > float32(sh)+1 {
		t.Fatalf("content past screen height %+v", lay.Content)
	}
}
