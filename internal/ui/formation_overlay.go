package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/party"
)

// DrawFormationOverlay — порядок отряда (formation) в едином стиле с battle/explore foundation.
func DrawFormationOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, selected int, screenW, screenH int) {
	if hudFace == nil || p == nil {
		return
	}
	sw := float32(screenW)
	vector.FillRect(screen, 0, 0, sw, float32(screenH), Theme.OverlayDim, false)

	pad := float32(20)
	lineH := uiLineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}

	panelW := float32(520)
	if sw-pad*2 < panelW {
		panelW = sw - pad*2
	}
	panelX := (sw - panelW) * 0.5
	panelY := pad * 1.2

	n := len(p.Active)
	rowH := lineH*2.4 + 10
	headerH := lineH * 4.2
	footerH := lineH * 1.6
	panelH := headerH
	if n > 0 {
		panelH += float32(n)*rowH + 8
	}
	panelH += footerH

	vector.FillRect(screen, panelX, panelY, panelW, panelH, Theme.PanelBG, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1, Theme.PanelBorder, false)
	DrawThinAccentLine(screen, panelX+10, panelY+8, panelW-20)

	innerX := panelX + 16
	y := panelY + 14
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.1}, "Порядок отряда · построение в бою", metrics, Theme.TextPrimary)
	y += lineH * 1.35
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.2},
		"Сверху вниз: передний ряд (3), затем задний (слоты = PlayerSlotForPartyIndex).", metrics, Theme.TextSecondary)
	y += lineH * 2.0

	if n == 0 {
		drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH}, "Нет участников.", metrics, Theme.TextDanger)
		return
	}

	for i := 0; i < n; i++ {
		h := p.Active[i]
		slot := party.FormationSlotCaption(i)
		role := party.MemberRoleCaption(i)
		ry := y
		rowW := panelW - 32
		fill := Theme.PanelBGDeep
		border := Theme.AllyAccent
		if i == selected {
			fill = Theme.AbilitySelectedBG
			border = Theme.ActiveTurn
		}
		vector.FillRect(screen, innerX, ry, rowW, rowH, fill, false)
		vector.StrokeRect(screen, innerX, ry, rowW, rowH, 2, border, false)

		lbl := fmt.Sprintf("%d. %s", i+1, role)
		if i == selected {
			lbl = "▶ " + lbl
		}
		col := Theme.TextPrimary
		if h.CurrentHP <= 0 {
			col = Theme.DeadText
		} else if i == selected {
			col = color.RGBA{R: 255, G: 235, B: 160, A: 255}
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 10, Y: ry + 6, W: rowW - 20, H: lineH}, lbl, metrics, col)

		slotShort := slot
		if len([]rune(slotShort)) > 38 {
			rs := []rune(slotShort)
			slotShort = string(rs[:35]) + "…"
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 10, Y: ry + 6 + lineH*1.05, W: rowW - 100, H: lineH * 0.95}, slotShort, metrics, Theme.TextMuted)

		hpTxt := fmt.Sprintf("%d/%d", h.CurrentHP, h.MaxHP)
		if h.CurrentHP <= 0 {
			hpTxt = "выбыл"
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + rowW - 88, Y: ry + 6, W: 78, H: lineH}, hpTxt, metrics, Theme.TextSecondary)

		barY := ry + rowH - 9
		DrawHPBarMicro(screen, innerX+10, barY, rowW-20, 5, h.CurrentHP, h.MaxHP, h.CurrentHP > 0, false)

		y += rowH + 6
	}

	y += lineH * 0.35
	help := "↑↓ выбор   ←→ сдвиг (слот в бою)   Esc / F5 — выход"
	if n < 2 {
		help = "Нужно ≥2 участника для сдвига.   Esc / F5 — выход"
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH}, help, metrics, Theme.TextMuted)
}
