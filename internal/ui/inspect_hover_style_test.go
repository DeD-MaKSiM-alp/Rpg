package ui

import (
	"testing"

	battlepkg "mygame/internal/battle"
)

func TestInspectHoverStrength_zeroHover(t *testing.T) {
	if v := InspectHoverStrength(true, 3, 0); v != 0 {
		t.Fatalf("want 0, got %v", v)
	}
}

func TestInspectHoverStrength_softerWhenOpenOtherUnit(t *testing.T) {
	a := InspectHoverStrength(true, 5, 3)
	b := InspectHoverStrength(false, 0, 3)
	if a >= b {
		t.Fatalf("open inspect on another unit should soften hover: %v vs %v", a, b)
	}
}

func TestInspectHoverStrength_fullWhenHoverMatchesOpen(t *testing.T) {
	if v := InspectHoverStrength(true, 7, 7); v != 1.0 {
		t.Fatalf("want 1.0, got %v", v)
	}
}

func TestInspectHoverStrength_fullWhenInspectClosed(t *testing.T) {
	if v := InspectHoverStrength(false, 0, 4); v != 1.0 {
		t.Fatalf("want 1.0, got %v", v)
	}
}

func TestBuildInspectBattleHighlightPlan_hoverOnly(t *testing.T) {
	p := BuildInspectBattleHighlightPlan(4, 0, false)
	if p.CombinedUnitID != 0 || p.ActiveUnitID != 0 || p.HoverUnitID != 4 || p.HoverStrength != 1.0 {
		t.Fatalf("unexpected plan: %+v", p)
	}
}

func TestBuildInspectBattleHighlightPlan_activeOnly(t *testing.T) {
	// Contract for modal battle inspect: the game layer passes hoverID=0 while the overlay is open,
	// so DrawBattleInspectHighlights shows only the opened unit (no "hover other unit" layer).
	p := BuildInspectBattleHighlightPlan(0, 7, true)
	if p.CombinedUnitID != 0 || p.ActiveUnitID != 7 || p.HoverUnitID != 0 || p.HoverStrength != 0 {
		t.Fatalf("unexpected plan: %+v", p)
	}
}

func TestBuildInspectBattleHighlightPlan_combined(t *testing.T) {
	p := BuildInspectBattleHighlightPlan(3, 3, true)
	if p.CombinedUnitID != 3 || p.ActiveUnitID != 0 || p.HoverUnitID != 0 {
		t.Fatalf("unexpected plan: %+v", p)
	}
}

func TestBuildInspectBattleHighlightPlan_activePlusHoverOther(t *testing.T) {
	p := BuildInspectBattleHighlightPlan(2, 9, true)
	if p.CombinedUnitID != 0 || p.ActiveUnitID != 9 || p.HoverUnitID != 2 || p.HoverStrength != 0.52 {
		t.Fatalf("unexpected plan: %+v", p)
	}
}

func TestBuildFormationInspectHighlightPlan_matchesBattleLogic(t *testing.T) {
	open := 9
	h := 2
	if FormationInspectHoverStrength(true, open, h) != InspectHoverStrength(true, battlepkg.UnitID(open), battlepkg.UnitID(h)) {
		t.Fatal("formation hover strength should match battle for same numeric ids")
	}
	p := BuildFormationInspectHighlightPlan(h, open, true)
	if p.CombinedGlobalIdx != -1 || p.ActiveGlobalIdx != open || p.HoverGlobalIdx != h || p.HoverStrength != 0.52 {
		t.Fatalf("unexpected formation plan: %+v", p)
	}
}

func TestBuildFormationInspectHighlightPlan_combinedRow(t *testing.T) {
	p := BuildFormationInspectHighlightPlan(4, 4, true)
	if p.CombinedGlobalIdx != 4 || p.ActiveGlobalIdx != -1 || p.HoverGlobalIdx != -1 {
		t.Fatalf("unexpected plan: %+v", p)
	}
}
