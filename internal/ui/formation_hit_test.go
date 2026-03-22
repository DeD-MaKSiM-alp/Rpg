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
	geom := ComputeFormationOverlayGeom(sw, sh, &p, uiLineH)
	innerX := geom.InnerX
	y := geom.RowY0
	rowH := geom.RowH
	mx := int(innerX + 50)
	my := int(y + rowH/2)
	if g := FormationHitTestGlobalIndex(sw, sh, mx, my, &p); g != 0 {
		t.Fatalf("want global index 0, got %d", g)
	}
}
