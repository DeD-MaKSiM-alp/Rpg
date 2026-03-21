package postbattle

import "testing"

func TestBuildPostBattleKeyboardIntents(t *testing.T) {
	tests := []struct {
		name                          string
		space, enter                  bool
		up, left, down, right         bool
		wantConfirm, wantPrev, wantNext bool
	}{
		{name: "empty"},
		{name: "confirm_space", space: true, wantConfirm: true},
		{name: "confirm_enter", enter: true, wantConfirm: true},
		{name: "prev_up", up: true, wantPrev: true},
		{name: "prev_left", left: true, wantPrev: true},
		{name: "next_down", down: true, wantNext: true},
		{name: "next_right", right: true, wantNext: true},
		{name: "all_move", up: true, right: true, wantPrev: true, wantNext: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildPostBattleKeyboardIntents(tc.space, tc.enter, tc.up, tc.left, tc.down, tc.right)
			if got.Confirm != tc.wantConfirm || got.Prev != tc.wantPrev || got.Next != tc.wantNext {
				t.Fatalf("got %+v want confirm=%v prev=%v next=%v", got, tc.wantConfirm, tc.wantPrev, tc.wantNext)
			}
		})
	}
}

func TestBuildPostBattleMouseButtons(t *testing.T) {
	got := BuildPostBattleMouseButtons(true, false)
	if !got.LeftJustPressed || got.RightJustPressed {
		t.Fatalf("%+v", got)
	}
}
