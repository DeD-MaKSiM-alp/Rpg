package ui

import "mygame/internal/party"

// FormationOverlayGeom — общая геометрия DrawFormationOverlay и FormationHitTestGlobalIndex.
type FormationOverlayGeom struct {
	Panel         FRect
	InnerX        float32
	RowY0         float32 // верх первой строки ростера
	RowH          float32
	RowGap        float32 // между строками (как в draw: +6)
	ReserveTitleH float32
	LineH         float32 // итоговая высота строки (может быть уменьшена под max высоту модалки)
}

// FormationPanelContentHeight — высота панели без позиционирования (как в formation_overlay).
func FormationPanelContentHeight(na, nr int, lineH float32) float32 {
	rowH := lineH*2.4 + 10
	headerH := lineH * 4.2
	footerH := lineH * 2.4
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
	panelH += footerH
	return panelH
}

// ComputeFormationOverlayGeom — центрирование модалки; при нехватке высоты уменьшает LineH до минимума.
func ComputeFormationOverlayGeom(screenW, screenH int, p *party.Party, baseLineH float32) FormationOverlayGeom {
	var out FormationOverlayGeom
	if p == nil || screenW < 1 || screenH < 1 {
		return out
	}
	lay := ComputeScreenLayout(screenW, screenH, 0)
	na, nr := len(p.Active), len(p.Reserve)
	lh := baseLineH
	if lh < 11 {
		lh = 11
	}
	for {
		natH := FormationPanelContentHeight(na, nr, lh)
		w, _ := FormationPanelBaseSize(screenW, lay.Tier)
		panel := CenterPanelInModal(lay, w, natH)
		if natH <= panel.H+0.5 || lh <= 11.05 {
			out.Panel = panel
			out.LineH = lh
			break
		}
		lh *= panel.H / natH
		if lh < 11 {
			lh = 11
			natH = FormationPanelContentHeight(na, nr, lh)
			w, _ := FormationPanelBaseSize(screenW, lay.Tier)
			out.Panel = CenterPanelInModal(lay, w, natH)
			out.LineH = lh
			break
		}
	}

	lineH := out.LineH
	out.RowH = lineH*2.4 + 10
	out.RowGap = 6
	out.ReserveTitleH = lineH * 1.15
	out.InnerX = out.Panel.X + 16
	out.RowY0 = out.Panel.Y + 14 + lineH*1.35 + lineH*2.0
	return out
}
