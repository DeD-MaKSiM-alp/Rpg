package ui

import (
	"testing"

	"mygame/internal/party"
)

func TestFormationHitTestGlobalIndex_outside(t *testing.T) {
	p := party.DefaultParty()
	if FormationHitTestGlobalIndex(800, 600, -500, -500, &p) != -1 {
		t.Fatalf("expected -1 for far outside")
	}
}

func TestFormationHitTestGlobalIndex_firstActiveRow(t *testing.T) {
	p := party.DefaultParty()
	sw, sh := 800, 600
	pad := float32(20)
	lineH := uiLineH
	panelW := float32(560)
	panelX := (float32(sw) - panelW) * 0.5
	panelY := pad * 1.2
	innerX := panelX + 16
	y := panelY + 14 + lineH*1.35 + lineH*2.0
	rowH := lineH*2.4 + 10
	mx := int(innerX + 50)
	my := int(y + rowH/2)
	if g := FormationHitTestGlobalIndex(sw, sh, mx, my, &p); g != 0 {
		t.Fatalf("want global index 0, got %d", g)
	}
}
