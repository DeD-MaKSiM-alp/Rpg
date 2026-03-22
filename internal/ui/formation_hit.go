package ui

import (
	"mygame/internal/party"
)

// FormationHitTestGlobalIndex возвращает глобальный индекс строки состава под курсором или -1.
// Геометрия совпадает с DrawFormationOverlay.
func FormationHitTestGlobalIndex(screenW, screenH, mx, my int, p *party.Party) int {
	if p == nil {
		return -1
	}
	geom := ComputeFormationOverlayGeom(screenW, screenH, p, uiLineH)
	panelX := geom.Panel.X
	panelY := geom.Panel.Y
	panelW := geom.Panel.W
	na, nr := len(p.Active), len(p.Reserve)
	rowH := geom.RowH
	reserveTitleH := geom.ReserveTitleH

	mxf := float32(mx)
	myf := float32(my)
	if mxf < panelX || mxf > panelX+panelW || myf < panelY || myf > panelY+geom.Panel.H {
		return -1
	}

	innerX := geom.InnerX
	y := geom.RowY0

	hitRow := func(ry float32) bool {
		return mxf >= innerX && mxf <= innerX+panelW-32 && myf >= ry && myf <= ry+rowH
	}

	for i := 0; i < na; i++ {
		if hitRow(y) {
			return i
		}
		y += rowH + 6
	}
	if nr > 0 {
		y += reserveTitleH + 6
		for j := 0; j < nr; j++ {
			if hitRow(y) {
				return na + j
			}
			y += rowH + 6
		}
	}
	return -1
}
