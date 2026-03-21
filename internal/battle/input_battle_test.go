package battle

import "testing"

func TestBuildBattleKeyboardIntents(t *testing.T) {
	tests := []struct {
		name                    string
		space, enter            bool
		back                    bool
		up, left, down, right   bool
		esc                     bool
		wantConfirm, wantBack   bool
		wantPrev, wantNext      bool
		wantEscape              bool
	}{
		{
			name: "empty", wantConfirm: false, wantBack: false, wantPrev: false, wantNext: false, wantEscape: false,
		},
		{
			name: "confirm_space", space: true,
			wantConfirm: true,
		},
		{
			name: "confirm_enter", enter: true,
			wantConfirm: true,
		},
		{
			name: "confirm_both", space: true, enter: true,
			wantConfirm: true,
		},
		{
			name: "back", back: true,
			wantBack: true,
		},
		{
			name: "prev_up", up: true,
			wantPrev: true,
		},
		{
			name: "prev_left", left: true,
			wantPrev: true,
		},
		{
			name: "next_down", down: true,
			wantNext: true,
		},
		{
			name: "next_right", right: true,
			wantNext: true,
		},
		{
			name: "escape", esc: true,
			wantEscape: true,
		},
		{
			name: "prev_and_next", up: true, right: true,
			wantPrev: true, wantNext: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildBattleKeyboardIntents(tc.space, tc.enter, tc.back, tc.up, tc.left, tc.down, tc.right, tc.esc)
			if got.Confirm != tc.wantConfirm || got.Back != tc.wantBack || got.Prev != tc.wantPrev || got.Next != tc.wantNext || got.Escape != tc.wantEscape {
				t.Fatalf("got %+v want confirm=%v back=%v prev=%v next=%v esc=%v",
					got, tc.wantConfirm, tc.wantBack, tc.wantPrev, tc.wantNext, tc.wantEscape)
			}
		})
	}
}

func TestBuildBattleMouseButtons(t *testing.T) {
	got := BuildBattleMouseButtons(true, false)
	if !got.LeftJustPressed || got.RightJustPressed {
		t.Fatalf("%+v", got)
	}
	got2 := BuildBattleMouseButtons(false, true)
	if got2.LeftJustPressed || !got2.RightJustPressed {
		t.Fatalf("%+v", got2)
	}
}
