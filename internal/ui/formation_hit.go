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
	sw := float32(screenW)
	pad := float32(20)
	lineH := uiLineH
	panelW := float32(560)
	if sw-pad*2 < panelW {
		panelW = sw - pad*2
	}
	panelX := (sw - panelW) * 0.5
	panelY := pad * 1.2

	na, nr := len(p.Active), len(p.Reserve)
	rowH := lineH*2.4 + 10
	headerH := lineH * 4.2
	reserveTitleH := lineH * 1.15

	panelH := headerH
	if na > 0 {
		panelH += float32(na)*rowH + float32(max(0, na-1))*6
	}
	if nr > 0 {
		panelH += reserveTitleH + 6 + float32(nr)*rowH + float32(max(0, nr-1))*6
	}
	if na == 0 && nr == 0 {
		panelH += lineH * 2
	}
	footerH := lineH * 2.4
	panelH += footerH

	mxf := float32(mx)
	myf := float32(my)
	if mxf < panelX || mxf > panelX+panelW || myf < panelY || myf > panelY+panelH {
		return -1
	}

	innerX := panelX + 16
	y := panelY + 14 + lineH*1.35 + lineH*2.0

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
