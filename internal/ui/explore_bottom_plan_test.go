package ui

import "testing"

func TestPlanExploreBottomLines_order(t *testing.T) {
	b := ExploreHUDLayout{
		ZoneLine:        "зона",
		InteractionHint: "действие",
		RestFeedback:    "отдых",
		POIFeedback:     "poi",
	}
	lines := PlanExploreBottomLines(b)
	if len(lines) < 5 {
		t.Fatalf("expected zone, interaction, hotkeys, rest, poi; got %d lines", len(lines))
	}
	if lines[0].Kind != BottomKindZone || lines[0].Text != "зона" {
		t.Fatalf("first line: %+v", lines[0])
	}
	if lines[1].Kind != BottomKindInteraction {
		t.Fatalf("second should be interaction: %+v", lines[1])
	}
	if lines[2].Kind != BottomKindHotkeys {
		t.Fatalf("hotkeys line expected at 2: %+v", lines[2])
	}
}
