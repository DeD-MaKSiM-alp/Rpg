package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
)

// DrawFormationOverlay — порядок Active, резерв и подсказки (F5 в explore).
// selected — глобальный индекс: [0, len(Active)) строки строя, [len(Active), len(Active)+len(Reserve)) — резерв.
// inspectOpen — открыта карточка бойца (I); подсказки сокращаются до навигации по карточке.
// hoverGlobalIdx — строка под курсором (-1 = нет); подсветка для ПКМ-inspect.
// promotionRowHints — короткие метки повышения по globalIdx (из game.PromotionFormationRowHints); nil или короче списка — без метки.
func DrawFormationOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, selected int, screenW, screenH int, inspectOpen bool, hoverGlobalIdx int, promotionRowHints []string) {
	if hudFace == nil || p == nil {
		return
	}
	sw := float32(screenW)
	vector.FillRect(screen, 0, 0, sw, float32(screenH), Theme.OverlayDim, false)

	geom := ComputeFormationOverlayGeom(screenW, screenH, p, uiLineH)
	lineH := geom.LineH
	if lineH < 1 {
		lineH = uiLineH
	}
	metrics := battlepkg.HUDMetrics{LineH: lineH}
	panelX := geom.Panel.X
	panelY := geom.Panel.Y
	panelW := geom.Panel.W
	panelH := geom.Panel.H
	na, nr := len(p.Active), len(p.Reserve)
	rowH := geom.RowH
	reserveTitleH := geom.ReserveTitleH

	drawUnifiedModalPanelChrome(screen, panelX, panelY, panelW, panelH)
	DrawThinAccentLine(screen, panelX+10, panelY+8, panelW-20)

	innerX := geom.InnerX
	y := panelY + 14
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.1}, "Состав отряда · в бою и резерв", metrics, Theme.TextPrimary)
	y += lineH * 1.35
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.2},
		"В бою: до "+fmt.Sprintf("%d", party.MaxActiveBattleSlots)+" в строю. Боевой опыт — только выжившим в строю; резерв в бою не участвует.", metrics, Theme.TextSecondary)
	y += lineH * 2.0

	if na == 0 && nr == 0 {
		drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH}, "Нет участников.", metrics, Theme.TextDanger)
		return
	}

	inspectPlan := BuildFormationInspectHighlightPlan(hoverGlobalIdx, selected, inspectOpen)

	promoHintAt := func(globalIdx int) string {
		if globalIdx < 0 || globalIdx >= len(promotionRowHints) {
			return ""
		}
		return promotionRowHints[globalIdx]
	}

	drawMemberRow := func(globalIdx int, h hero.Hero, slotCaption, roleCaption string, inReserve bool, promoHint string) {
		ry := y
		rowW := panelW - 32
		fill := Theme.RosterCardContentWell
		border := Theme.AllyAccent
		strokeW := float32(2)

		useCombined := inspectOpen && inspectPlan.CombinedGlobalIdx == globalIdx
		useActive := inspectOpen && inspectPlan.ActiveGlobalIdx == globalIdx && inspectPlan.CombinedGlobalIdx < 0
		useHover := inspectPlan.HoverGlobalIdx == globalIdx && inspectPlan.HoverStrength > 0 && inspectPlan.CombinedGlobalIdx < 0
		if useHover && !inspectOpen && globalIdx == selected {
			useHover = false
		}
		navSelected := globalIdx == selected && !inspectOpen

		switch {
		case useCombined:
			fill = formationInspectCombinedFill()
			border = Theme.ValidTarget
			strokeW = 2.95
		case useActive:
			fill = formationInspectActiveOpenFill()
			border = Theme.HoverTarget
			strokeW = 2.45
		case navSelected:
			fill = Theme.AbilitySelectedBG
			border = Theme.ActiveTurn
		case useHover:
			fill = formationHoverFill(inspectPlan.HoverStrength)
			border = Theme.HoverTarget
		}

		vector.FillRect(screen, innerX, ry, rowW, rowH, fill, false)
		vector.StrokeRect(screen, innerX, ry, rowW, rowH, strokeW, border, false)
		switch {
		case useCombined:
			vector.StrokeRect(screen, innerX-2, ry-2, rowW+4, rowH+4, 1, Theme.AccentStrip, false)
			vector.FillRect(screen, innerX, ry, 5, rowH, Theme.AccentStrip, false)
		case useActive:
			vector.StrokeRect(screen, innerX-1, ry-1, rowW+2, rowH+2, 1, Theme.AccentStrip, false)
			vector.FillRect(screen, innerX, ry, 4, rowH, Theme.AccentStrip, false)
		case useHover:
			vector.FillRect(screen, innerX, ry, 3, rowH, Theme.AccentStrip, false)
		}

		lbl := fmt.Sprintf("%s · %s", roleCaption, slotCaption)
		if globalIdx == selected {
			lbl = "▶ " + lbl
		}
		col := Theme.TextPrimary
		if h.CurrentHP <= 0 {
			col = Theme.DeadText
		} else if useCombined {
			col = color.RGBA{R: 255, G: 240, B: 175, A: 255}
		} else if useActive {
			col = color.RGBA{R: 255, G: 235, B: 165, A: 255}
		} else if globalIdx == selected {
			col = color.RGBA{R: 255, G: 235, B: 160, A: 255}
		} else if useHover {
			col = Theme.TextSecondary
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 10, Y: ry + 6, W: rowW - 20, H: lineH}, lbl, metrics, col)

		slotShort := slotCaption
		if len([]rune(slotShort)) > 36 {
			rs := []rune(slotShort)
			slotShort = string(rs[:33]) + "…"
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 10, Y: ry + 6 + lineH*1.05, W: rowW - 100, H: lineH * 0.95}, slotShort, metrics, Theme.TextMuted)

		if promoHint != "" {
			hCol := Theme.TextMuted
			switch promoHint {
			case "Повышение!":
				hCol = Theme.TextSuccess
			case "Лагерь":
				hCol = Theme.HoverTarget
			case "Знаки", "Ветка":
				hCol = Theme.ValidTarget
			}
			// Справа от подписи роли, левее блока HP (см. hpTxt ниже).
			drawSingleLineInRect(screen, hudFace, rect{X: innerX + rowW - 230, Y: ry + 6, W: 100, H: lineH * 0.92}, promoHint, metrics, hCol)
		}

		hpTxt := fmt.Sprintf("%d/%d", h.CurrentHP, h.MaxHP)
		if h.CurrentHP <= 0 {
			hpTxt = "выбыл"
		}
		if inReserve {
			hpTxt = hpTxt + " · резерв"
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + rowW - 120, Y: ry + 6, W: 110, H: lineH}, hpTxt, metrics, Theme.TextSecondary)

		barY := ry + rowH - 9
		DrawHPBarMicro(screen, innerX+10, barY, rowW-20, 5, h.CurrentHP, h.MaxHP, h.CurrentHP > 0, false)

		y += rowH + 6
	}

	for i := 0; i < na; i++ {
		slot := party.FormationSlotCaption(i)
		role := party.MemberRoleCaption(i)
		drawMemberRow(i, p.Active[i], slot, role, false, promoHintAt(i))
	}

	if nr > 0 {
		drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: reserveTitleH}, "— Резерв (не в бою) —", metrics, Theme.TextSecondary)
		y += reserveTitleH + 6
		for j := 0; j < nr; j++ {
			gidx := na + j
			role := party.ReserveRowCaption(j)
			drawMemberRow(gidx, p.Reserve[j], "не участвует в бою", role, true, promoHintAt(gidx))
		}
	}

	y += lineH * 0.2
	var help string
	if inspectOpen {
		help = "Карточка бойца · ↑↓ другой · P — повышение · I / Esc / F5 — закрыть · ПКМ по строке — сведения"
	} else {
		help = "↑↓ выбор   I / ПКМ — карточка бойца   ←→ сдвиг слота (строй)   Enter — резерв↔строй   Esc / F5 — выход"
		if na < 2 {
			help = "↑↓ выбор   I / ПКМ — карточка   Enter — резерв↔строй   Esc / F5 — выход   (сдвиг слотов при ≥2 в строю)"
		}
		if na >= party.MaxActiveBattleSlots && nr > 0 {
			help = "Строй полон (" + fmt.Sprintf("%d", party.MaxActiveBattleSlots) + "). Уберите в резерв, чтобы освободить место."
		}
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: panelW - 32, H: lineH * 1.1}, help, metrics, Theme.TextMuted)
}

func formationHoverFill(strength float32) color.RGBA {
	a := uint8(42 * strength)
	if a < 1 {
		a = 1
	}
	return color.RGBA{R: 48, G: 62, B: 88, A: a}
}

func formationInspectActiveOpenFill() color.RGBA {
	return color.RGBA{R: 48, G: 66, B: 94, A: 40}
}

func formationInspectCombinedFill() color.RGBA {
	return color.RGBA{R: 58, G: 78, B: 112, A: 54}
}
