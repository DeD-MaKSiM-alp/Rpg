// battle_panels.go — battle HUD v1 rendering. Uses shared helpers from panels.go (rect, drawPanelBox, drawSingleLineInRect, drawLinesInRect, fitTextToWidth, inset, clampF, minInt).

package ui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

const battlePanelBorder = float32(2)

func battleToRect(hr battlepkg.HUDRect) rect {
	return rect{X: hr.X, Y: hr.Y, W: hr.W, H: hr.H}
}

// drawBattleOverlayPanel рисует затемнённый фон и центральную панель боевого overlay.
func drawBattleOverlayPanel(screen *ebiten.Image, screenWidth, screenHeight int, layout battlepkg.BattleHUDLayout) rect {
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	panelColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	panelBorderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)

	ov := layout.Overlay
	vector.FillRect(screen, ov.X, ov.Y, ov.W, ov.H, panelColor, false)
	vector.StrokeRect(screen, ov.X, ov.Y, ov.W, ov.H, battlePanelBorder, panelBorderColor, false)
	return rect{X: ov.X, Y: ov.Y, W: ov.W, H: ov.H}
}

// drawBattleOverlayText рисует battle HUD v1: жёсткая сетка, только drawSingleLineInRect / drawLinesInRect.
func drawBattleOverlayText(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics

	// Top block hierarchy: title primary, info rows secondary.
	titleRow := battleToRect(layout.TitleRow)
	if titleRow.W > 0 && titleRow.H > 0 {
		title := "Battle"
		if battle != nil && len(battle.Encounter.Enemies) > 0 {
			title = fmt.Sprintf("Battle: enemy #%d", battle.Encounter.Enemies[0].EnemyID)
		}
		drawSingleLineInRect(screen, hudFace, titleRow, title, metrics, color.White)
	}

	if battle == nil {
		return
	}

	if battle.Result != battlepkg.ResultNone {
		info1 := battleToRect(layout.InfoRow1)
		var banner string
		switch battle.Result {
		case battlepkg.ResultVictory:
			banner = "VICTORY"
		case battlepkg.ResultDefeat:
			banner = "DEFEAT"
		case battlepkg.ResultEscape:
			banner = "ESCAPE"
		default:
			banner = battle.ResultString()
		}
		drawSingleLineInRect(screen, hudFace, info1, banner, metrics, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		info2 := battleToRect(layout.InfoRow2)
		drawSingleLineInRect(screen, hudFace, info2, "SPACE/ENTER: continue", metrics, color.RGBA{R: 180, G: 180, B: 180, A: 255})
	} else {
		info1 := battleToRect(layout.InfoRow1)
		line1 := fmt.Sprintf("Round %d | Phase: %s", battle.Round, battle.PhaseString())
		drawSingleLineInRect(screen, hudFace, info1, line1, metrics, color.RGBA{R: 200, G: 200, B: 200, A: 255})

		info2 := battleToRect(layout.InfoRow2)
		active := battle.ActiveUnit()
		activeStr := "-"
		if active != nil {
			activeStr = fmt.Sprintf("Active: %s (#%d)", active.Name(), active.ID)
			if active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
				activeStr = fmt.Sprintf("%s | %s", activeStr, battle.PlayerTurn.PhaseString())
			}
		}
		drawSingleLineInRect(screen, hudFace, info2, activeStr, metrics, color.RGBA{R: 180, G: 180, B: 180, A: 255})
	}

	footerRect := battleToRect(layout.Footer)
	playerPanel := battleToRect(layout.PlayerFormation)
	enemyPanel := battleToRect(layout.EnemyFormation)

	drawFormationPanel(screen, hudFace, battle, playerPanel, battlepkg.BattleSidePlayer, "PLAYER", layout)
	drawFormationPanel(screen, hudFace, battle, enemyPanel, battlepkg.BattleSideEnemy, "ENEMY", layout)

	abilitiesRect := battleToRect(layout.Abilities)
	confirmRect := battleToRect(layout.Action)

	drawAbilityPanel(screen, hudFace, battle, abilitiesRect, layout)
	drawConfirmPanel(screen, hudFace, battle, confirmRect, layout)

	drawFooterPanel(screen, hudFace, battle, footerRect, layout)
}

func drawFormationPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, side battlepkg.BattleSide, title string, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	titleRow := battleToRect(layout.PlayerFormationTitleRow)
	if side == battlepkg.BattleSideEnemy {
		titleRow = battleToRect(layout.EnemyFormationTitleRow)
	}
	drawPanelBox(screen, r, titleRow, title, hudFace, metrics)
	if battle == nil {
		return
	}
	inner := inset(r, metrics.Pad*0.6)
	inner.Y += metrics.LineH + metrics.SmallGap*0.5
	inner.H -= metrics.LineH + metrics.SmallGap*0.5
	if inner.H < 0 {
		inner.H = 0
	}

	active := battle.ActiveUnit()
	isPlayerTurn := active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction
	pt := battle.PlayerTurn

	validSet := map[battlepkg.UnitID]bool{}
	if isPlayerTurn && (pt.Phase == battlepkg.PlayerChooseTarget || pt.Phase == battlepkg.PlayerConfirmAction) {
		for _, td := range pt.ValidTargets {
			if td.Kind == battlepkg.TargetKindUnit {
				validSet[td.UnitID] = true
			}
		}
	}

	selectedTargetID := battlepkg.UnitID(0)
	if isPlayerTurn && (pt.Phase == battlepkg.PlayerChooseTarget || pt.Phase == battlepkg.PlayerConfirmAction) && pt.SelectedTarget.Kind == battlepkg.TargetKindUnit {
		selectedTargetID = pt.SelectedTarget.UnitID
	}
	hoverTargetID := battlepkg.UnitID(0)
	if isPlayerTurn && (pt.Phase == battlepkg.PlayerChooseTarget || pt.Phase == battlepkg.PlayerConfirmAction) {
		hoverTargetID = pt.HoverTargetUnitID
	}

	labelH := metrics.LineH
	drawRowLabel := func(label string, y float32) {
		row := rect{X: inner.X, Y: y, W: inner.W, H: labelH}
		drawSingleLineInRect(screen, hudFace, row, label, metrics, color.RGBA{R: 170, G: 170, B: 170, A: 255})
	}

	frontLabelY := inner.Y
	frontSlotsY := frontLabelY + labelH
	backLabelY := inner.Y + (inner.H-labelH*2)*0.5
	backSlotsY := backLabelY + labelH

	drawRowLabel("FRONT", frontLabelY)
	drawRowLabel("BACK", backLabelY)

	drawSlot := func(row battlepkg.BattleRow, idx int, x, y float32) {
		slot := battle.Slot(side, row, idx)
		var u *battlepkg.BattleUnit
		if slot != nil {
			u = battle.UnitInSlot(slot)
		}

		fill := color.RGBA{R: 45, G: 45, B: 45, A: 255}
		border := color.RGBA{R: 90, G: 90, B: 90, A: 255}
		textCol := color.RGBA{R: 230, G: 230, B: 230, A: 255}

		if u == nil {
			fill = color.RGBA{R: 35, G: 35, B: 35, A: 255}
			textCol = color.RGBA{R: 120, G: 120, B: 120, A: 255}
		} else if !u.IsAlive() {
			fill = color.RGBA{R: 25, G: 25, B: 25, A: 255}
			textCol = color.RGBA{R: 120, G: 120, B: 120, A: 255}
		}

		if u != nil && u.ID == selectedTargetID {
			border = color.RGBA{R: 240, G: 80, B: 80, A: 255}
		} else if u != nil && u.ID == hoverTargetID {
			border = color.RGBA{R: 120, G: 190, B: 255, A: 255}
		} else if u != nil && validSet[u.ID] {
			border = color.RGBA{R: 80, G: 150, B: 255, A: 255}
		} else if active != nil && u != nil && u.ID == active.ID {
			border = color.RGBA{R: 255, G: 215, B: 80, A: 255}
		}

		w := float32(0)
		h := float32(0)
		if u != nil {
			if rUnit, ok := layout.UnitRects[u.ID]; ok {
				x = rUnit.X
				y = rUnit.Y
				w = rUnit.W
				h = rUnit.H
			}
		}
		if w == 0 || h == 0 {
			w = (inner.W - metrics.Gap*2) / 3
			h = clampF((inner.H-labelH*2-metrics.Gap*0.6)/2, metrics.LineH*2.4, metrics.LineH*3.5)
		}
		vector.FillRect(screen, x, y, w, h, fill, false)
		vector.StrokeRect(screen, x, y, w, h, 2, border, false)

		line1 := "EMPTY"
		line2 := ""
		if u != nil {
			hp := fmt.Sprintf("%d/%d", u.State.HP, u.MaxHP())
			name := u.Name()
			if len([]rune(name)) > 10 {
				rs := []rune(name)
				name = string(rs[:10])
			}
			if !u.IsAlive() {
				line1 = name
				line2 = "DEAD"
			} else {
				line1 = name
				line2 = "HP " + hp
			}
		}

		row1 := rect{X: x + 6, Y: y, W: w - 6, H: metrics.LineH}
		drawSingleLineInRect(screen, hudFace, row1, line1, metrics, textCol)
		if line2 != "" {
			row2 := rect{X: x + 6, Y: y + metrics.LineH, W: w - 6, H: metrics.LineH}
			drawSingleLineInRect(screen, hudFace, row2, line2, metrics, textCol)
		}
	}

	for i := 0; i < 3; i++ {
		x := inner.X + float32(i)*((inner.W - metrics.Gap*2) / 3)
		drawSlot(battlepkg.BattleRowFront, i, x, frontSlotsY)
		drawSlot(battlepkg.BattleRowBack, i, x, backSlotsY)
	}
}

func drawAbilityPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, battleToRect(layout.AbilitiesTitleRow), "ABILITIES", hudFace, metrics)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		titleRow := battleToRect(layout.AbilitiesTitleRow)
		if titleRow.W > 0 && titleRow.H > 0 {
			drawSingleLineInRect(screen, hudFace, titleRow, "(waiting)", metrics, color.RGBA{R: 150, G: 150, B: 150, A: 255})
		}
		return
	}

	abs := battlepkg.SpecialAbilities(active)
	sel := battle.PlayerTurn.SelectedAbilityID
	hoverIdx := battle.PlayerTurn.HoverAbilityIndex

	for i, id := range abs {
		if i >= len(layout.AbilityItemRects) {
			break
		}
		rowRect := battleToRect(layout.AbilityItemRects[i])
		a := battlepkg.GetAbility(id)
		prefix := " "
		col := color.RGBA{R: 220, G: 220, B: 220, A: 255}
		bg := color.RGBA{R: 35, G: 35, B: 35, A: 255}
		border := color.RGBA{R: 70, G: 70, B: 70, A: 255}
		if id == sel && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			prefix = "▶"
			bg = color.RGBA{R: 52, G: 52, B: 28, A: 255}
			col = color.RGBA{R: 255, G: 230, B: 100, A: 255}
			border = color.RGBA{R: 170, G: 170, B: 90, A: 255}
		} else if hoverIdx == i && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			bg = color.RGBA{R: 40, G: 55, B: 70, A: 255}
			col = color.RGBA{R: 180, G: 220, B: 255, A: 255}
			border = color.RGBA{R: 140, G: 190, B: 255, A: 255}
		}
		vector.FillRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, bg, false)
		vector.StrokeRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, 1, border, false)

		line := fmt.Sprintf("%s %s", prefix, a.Name)
		line = fitTextToWidth(hudFace, line, rowRect.W-12)
		textRow := rect{X: rowRect.X + 6, Y: rowRect.Y, W: rowRect.W - 12, H: rowRect.H}
		drawSingleLineInRect(screen, hudFace, textRow, line, metrics, col)
	}
}

func drawConfirmPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, battleToRect(layout.ActionTitleRow), "ACTION", hudFace, metrics)
	if battle == nil {
		return
	}

	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		summaryRect := battleToRect(layout.ActionSummary)
		if summaryRect.W > 0 && summaryRect.H > 0 {
			drawSingleLineInRect(screen, hudFace, summaryRect, "(enemy turn)", metrics, color.RGBA{R: 150, G: 150, B: 150, A: 255})
		}
		return
	}

	pt := battle.PlayerTurn
	a := battlepkg.GetAbility(pt.SelectedAbilityID)

	targetStr := "-"
	switch pt.SelectedTarget.Kind {
	case battlepkg.TargetKindSelf:
		targetStr = "self"
	case battlepkg.TargetKindUnit:
		if tu := battle.Units[pt.SelectedTarget.UnitID]; tu != nil {
			targetStr = fmt.Sprintf("%s (#%d)", tu.Name(), tu.ID)
		} else {
			targetStr = fmt.Sprintf("unit #%d", pt.SelectedTarget.UnitID)
		}
	case battlepkg.TargetKindNone:
		targetStr = "none"
	}

	summaryRect := battleToRect(layout.ActionSummary)
	summaryLines := []string{
		fmt.Sprintf("Step: %s", pt.PhaseString()),
		fmt.Sprintf("Ability: %s", a.Name),
		fmt.Sprintf("Target: %s", targetStr),
	}

	// Preview as 4th line only when ActionSummary fits at least 4 lines; otherwise omit to keep v1 layout stable.
	if maxLinesForRect(metrics, summaryRect, 0, 0, metrics.LineH) >= 4 {
		req := battlepkg.ActionRequest{Actor: active.ID, Ability: pt.SelectedAbilityID, Target: pt.SelectedTarget}
		preview, v := battlepkg.PreviewAction(battle, req)
		if v.OK && (preview.HasDamage() || preview.HasHeal()) {
			var previewStr string
			if preview.HasDamage() {
				previewStr = fmt.Sprintf("Preview: dmg %d-%d", preview.DamageMin, preview.DamageMax)
			} else {
				previewStr = fmt.Sprintf("Preview: heal %d-%d", preview.HealMin, preview.HealMax)
			}
			summaryLines = append(summaryLines, previewStr)
		}
	}

	maxSummaryW := summaryRect.W
	for i := range summaryLines {
		summaryLines[i] = fitTextToWidth(hudFace, summaryLines[i], maxSummaryW)
	}
	_ = drawLinesInRect(screen, hudFace, summaryRect, summaryLines, metrics, color.White, 0)

	backR := battleToRect(layout.BackButton)
	confirmR := battleToRect(layout.ConfirmButton)

	drawButton := func(r rect, label string, enabled, hovered bool) {
		baseFill := color.RGBA{R: 40, G: 40, B: 40, A: 255}
		baseBorder := color.RGBA{R: 140, G: 140, B: 140, A: 255}
		textCol := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		if !enabled {
			baseFill = color.RGBA{R: 30, G: 30, B: 30, A: 255}
			baseBorder = color.RGBA{R: 90, G: 90, B: 90, A: 255}
			textCol = color.RGBA{R: 180, G: 180, B: 180, A: 255}
		}
		if enabled && hovered {
			baseFill = color.RGBA{R: 60, G: 80, B: 100, A: 255}
			baseBorder = color.RGBA{R: 200, G: 220, B: 255, A: 255}
		}
		vector.FillRect(screen, r.X, r.Y, r.W, r.H, baseFill, false)
		vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 2, baseBorder, false)

		drawSingleLineInRect(screen, hudFace, r, label, metrics, textCol)
	}

	canConfirm := pt.Phase == battlepkg.PlayerConfirmAction
	drawButton(backR, "Back", true, pt.HoverBackButton)
	drawButton(confirmR, "Confirm", canConfirm, pt.HoverConfirmButton && canConfirm)
}

func drawFooterPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, battleToRect(layout.FooterTitleRow), "COMBAT LOG", hudFace, metrics)
	if battle == nil {
		return
	}
	logRect := battleToRect(layout.CombatLog)
	controlsRect := battleToRect(layout.HintLine)

	if logRect.W > 0 && logRect.H > 0 {
		availableLines := int(logRect.H / metrics.LineH)
		if availableLines < 1 {
			availableLines = 1
		}
		availableLines = minInt(availableLines, 3)
		logLines := make([]string, 0, availableLines)
		start := 0
		if len(battle.BattleLog) > availableLines {
			start = len(battle.BattleLog) - availableLines
		}
		for i := start; i < len(battle.BattleLog); i++ {
			logLines = append(logLines, strings.TrimSpace(battle.BattleLog[i]))
		}
		for i := range logLines {
			logLines[i] = fitTextToWidth(hudFace, logLines[i], logRect.W)
		}
		_ = drawLinesInRect(screen, hudFace, logRect, logLines, metrics, color.RGBA{R: 220, G: 220, B: 220, A: 255}, 0)
	}

	if controlsRect.W > 0 && controlsRect.H > 0 {
		controls := "LMB/RMB: action | Esc: retreat"
		drawSingleLineInRect(screen, hudFace, controlsRect, controls, metrics, color.RGBA{R: 155, G: 155, B: 155, A: 255})
	}
}

// drawBattleScreenV2 рисует battle screen в стиле Disciples: центр = сцена, слева/справа ростеры, внизу панель команд, сверху минимальный TopBar.
func drawBattleScreenV2(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	rosterBg := color.RGBA{R: 28, G: 28, B: 32, A: 255}
	rosterBorder := color.RGBA{R: 75, G: 75, B: 85, A: 255}
	bottomPanelBg := color.RGBA{R: 32, G: 32, B: 38, A: 255}
	bottomPanelBorder := color.RGBA{R: 70, G: 70, B: 80, A: 255}
	topBarBg := color.RGBA{R: 26, G: 26, B: 30, A: 255}
	topBarBorder := color.RGBA{R: 55, G: 55, B: 62, A: 255}

	// 1) Battlefield (center) — сцена: тёмный фон + чуть светлее центр (placeholder под арт)
	bf := layout.Battlefield
	if bf.W > 0 && bf.H > 0 {
		vector.FillRect(screen, bf.X, bf.Y, bf.W, bf.H, color.RGBA{R: 18, G: 18, B: 24, A: 255}, false)
		insetPx := float32(24)
		if bf.W > insetPx*2 && bf.H > insetPx*2 {
			cx, cy := bf.X+insetPx, bf.Y+insetPx
			cw, ch := bf.W-insetPx*2, bf.H-insetPx*2
			vector.FillRect(screen, cx, cy, cw, ch, color.RGBA{R: 24, G: 24, B: 32, A: 255}, false)
		}
		vector.StrokeRect(screen, bf.X, bf.Y, bf.W, bf.H, 1, color.RGBA{R: 48, G: 48, B: 56, A: 255}, false)
	}

	// 2) Left / Right rosters — боковые панели, визуально отделены от сцены
	lr := layout.LeftRoster
	if lr.W > 0 && lr.H > 0 {
		vector.FillRect(screen, lr.X, lr.Y, lr.W, lr.H, rosterBg, false)
		vector.StrokeRect(screen, lr.X, lr.Y, lr.W, lr.H, 1, rosterBorder, false)
	}
	rr := layout.RightRoster
	if rr.W > 0 && rr.H > 0 {
		vector.FillRect(screen, rr.X, rr.Y, rr.W, rr.H, rosterBg, false)
		vector.StrokeRect(screen, rr.X, rr.Y, rr.W, rr.H, 1, rosterBorder, false)
	}

	// 3) Unit cards in rosters
	for id, hr := range layout.UnitRects {
		u := battle.Units[id]
		if u == nil {
			continue
		}
		r := battleToRect(hr)
		fill := color.RGBA{R: 42, G: 42, B: 48, A: 255}
		border := color.RGBA{R: 85, G: 85, B: 95, A: 255}
		textCol := color.RGBA{R: 228, G: 228, B: 228, A: 255}
		if !u.IsAlive() {
			fill = color.RGBA{R: 24, G: 24, B: 26, A: 255}
			textCol = color.RGBA{R: 110, G: 110, B: 110, A: 255}
		}
		active := battle.ActiveUnit()
		if active != nil && active.ID == u.ID {
			border = color.RGBA{R: 255, G: 210, B: 70, A: 255}
		}
		pt := &battle.PlayerTurn
		if pt.HoverTargetUnitID == u.ID {
			border = color.RGBA{R: 100, G: 170, B: 255, A: 255}
		}
		if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && pt.SelectedTarget.UnitID == u.ID {
			border = color.RGBA{R: 235, G: 70, B: 70, A: 255}
		}
		vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
		vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 2, border, false)
		name := u.Name()
		if len([]rune(name)) > 10 {
			rs := []rune(name)
			name = string(rs[:10])
		}
		row1 := rect{X: r.X + 8, Y: r.Y, W: r.W - 16, H: metrics.LineH}
		drawSingleLineInRect(screen, hudFace, row1, name, metrics, textCol)
		hpStr := "DEAD"
		if u.IsAlive() {
			hpStr = fmt.Sprintf("HP %d/%d", u.State.HP, u.MaxHP())
		}
		row2 := rect{X: r.X + 8, Y: r.Y + metrics.LineH, W: r.W - 16, H: metrics.LineH}
		drawSingleLineInRect(screen, hudFace, row2, hpStr, metrics, textCol)
	}

	// 4) Bottom panel — control panel: Active | Target → Abilities → Summary → Hint → Buttons
	bp := layout.BottomPanel
	if bp.W > 0 && bp.H > 0 {
		vector.FillRect(screen, bp.X, bp.Y, bp.W, bp.H, bottomPanelBg, false)
		vector.StrokeRect(screen, bp.X, bp.Y, bp.W, bp.H, 1, bottomPanelBorder, false)
	}
	activeR := battleToRect(layout.V2BottomActive)
	if activeR.W > 0 && activeR.H > 0 && battle != nil {
		active := battle.ActiveUnit()
		s := "—"
		if active != nil {
			s = active.Name()
		}
		drawSingleLineInRect(screen, hudFace, activeR, fitTextToWidth(hudFace, s, activeR.W), metrics, color.RGBA{R: 215, G: 215, B: 215, A: 255})
	}
	targetR := battleToRect(layout.V2BottomTarget)
	if targetR.W > 0 && targetR.H > 0 && battle != nil {
		pt := &battle.PlayerTurn
		s := "—"
		if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && battle.Units[pt.SelectedTarget.UnitID] != nil {
			s = battle.Units[pt.SelectedTarget.UnitID].Name()
		} else if pt.SelectedTarget.Kind == battlepkg.TargetKindSelf {
			s = "self"
		}
		drawSingleLineInRect(screen, hudFace, targetR, fitTextToWidth(hudFace, s, targetR.W), metrics, color.RGBA{R: 195, G: 195, B: 195, A: 255})
	}
	// Ability row (special abilities only; basic attack = default, click enemy)
	for i, hr := range layout.AbilityItemRects {
		rowRect := battleToRect(hr)
		abs := []battlepkg.AbilityID{}
		if active := battle.ActiveUnit(); active != nil {
			abs = battlepkg.SpecialAbilities(active)
		}
		if i >= len(abs) {
			break
		}
		a := battlepkg.GetAbility(abs[i])
		col := color.RGBA{R: 218, G: 218, B: 218, A: 255}
		bg := color.RGBA{R: 38, G: 38, B: 44, A: 255}
		brd := color.RGBA{R: 72, G: 72, B: 82, A: 255}
		if battle.PlayerTurn.SelectedAbilityID == abs[i] {
			col = color.RGBA{R: 255, G: 228, B: 95, A: 255}
			brd = color.RGBA{R: 165, G: 165, B: 85, A: 255}
		}
		if battle.PlayerTurn.HoverAbilityIndex == i {
			brd = color.RGBA{R: 130, G: 180, B: 255, A: 255}
		}
		vector.FillRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, bg, false)
		vector.StrokeRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, 1, brd, false)
		drawSingleLineInRect(screen, hudFace, rowRect, fitTextToWidth(hudFace, a.Name, rowRect.W-8), metrics, col)
	}
	// Summary — коротко: default attack = "Attack · click enemy"; special = ability → target | preview
	summaryR := battleToRect(layout.V2BottomSummary)
	if summaryR.W > 0 && summaryR.H > 0 && battle != nil {
		active := battle.ActiveUnit()
		lines := []string{}
		if active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
			pt := &battle.PlayerTurn
			if pt.SelectedAbilityID == battlepkg.AbilityBasicAttack {
				lines = append(lines, fitTextToWidth(hudFace, "Attack · click enemy", summaryR.W))
				if pt.HoverTargetUnitID != 0 && battle.Units[pt.HoverTargetUnitID] != nil {
					preview, v := battlepkg.PreviewAction(battle, battlepkg.ActionRequest{Actor: active.ID, Ability: battlepkg.AbilityBasicAttack, Target: battlepkg.UnitTarget(pt.HoverTargetUnitID)})
					if v.OK && preview.HasDamage() {
						lines = append(lines, fmt.Sprintf("dmg %d-%d", preview.DamageMin, preview.DamageMax))
					}
				}
			} else {
				a := battlepkg.GetAbility(pt.SelectedAbilityID)
				targetStr := "—"
				if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && battle.Units[pt.SelectedTarget.UnitID] != nil {
					targetStr = battle.Units[pt.SelectedTarget.UnitID].Name()
				} else if pt.SelectedTarget.Kind == battlepkg.TargetKindSelf {
					targetStr = "self"
				}
				line1 := fmt.Sprintf("%s → %s", a.Name, targetStr)
				lines = append(lines, fitTextToWidth(hudFace, line1, summaryR.W))
				req := battlepkg.ActionRequest{Actor: active.ID, Ability: pt.SelectedAbilityID, Target: pt.SelectedTarget}
				preview, v := battlepkg.PreviewAction(battle, req)
				if v.OK && (preview.HasDamage() || preview.HasHeal()) {
					if preview.HasDamage() {
						lines = append(lines, fmt.Sprintf("dmg %d-%d", preview.DamageMin, preview.DamageMax))
					} else {
						lines = append(lines, fmt.Sprintf("heal %d-%d", preview.HealMin, preview.HealMax))
					}
				}
			}
		} else {
			lines = append(lines, "(waiting)")
		}
		if len(lines) > 2 {
			lines = lines[:2]
		}
		_ = drawLinesInRect(screen, hudFace, summaryR, lines, metrics, color.RGBA{R: 240, G: 240, B: 240, A: 255}, 2)
	}
	// Hint — в default attack: "Click enemy to attack"; иначе последняя строка лога или управление
	logR := battleToRect(layout.V2BottomLog)
	if logR.W > 0 && logR.H > 0 && battle != nil {
		hint := ""
		active := battle.ActiveUnit()
		if active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction && battle.PlayerTurn.SelectedAbilityID == battlepkg.AbilityBasicAttack {
			hint = "Click enemy to attack · RMB / Esc cancel"
		} else {
			if len(battle.BattleLog) > 0 {
				hint = strings.TrimSpace(battle.BattleLog[len(battle.BattleLog)-1])
			}
			if hint == "" {
				hint = "LMB/RMB · Esc retreat"
			}
		}
		drawSingleLineInRect(screen, hudFace, logR, fitTextToWidth(hudFace, hint, logR.W), metrics, color.RGBA{R: 145, G: 145, B: 150, A: 255})
	}
	// Buttons
	backR := battleToRect(layout.BackButton)
	confirmR := battleToRect(layout.ConfirmButton)
	drawButton := func(r rect, label string, enabled, hovered bool) {
		fill := color.RGBA{R: 38, G: 38, B: 44, A: 255}
		brd := color.RGBA{R: 125, G: 125, B: 138, A: 255}
		tcol := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		if !enabled {
			fill = color.RGBA{R: 28, G: 28, B: 32, A: 255}
			tcol = color.RGBA{R: 170, G: 170, B: 170, A: 255}
		}
		if enabled && hovered {
			fill = color.RGBA{R: 55, G: 72, B: 95, A: 255}
			brd = color.RGBA{R: 185, G: 210, B: 255, A: 255}
		}
		vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
		vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 2, brd, false)
		drawSingleLineInRect(screen, hudFace, r, label, metrics, tcol)
	}
	pt := &battle.PlayerTurn
	drawButton(backR, "Back", true, pt.HoverBackButton)
	drawButton(confirmR, "Confirm", pt.Phase == battlepkg.PlayerConfirmAction, pt.HoverConfirmButton && pt.Phase == battlepkg.PlayerConfirmAction)

	// 5) TopBar — лёгкая status line, не перебивает сцену
	tb := layout.TopBar
	if tb.W > 0 && tb.H > 0 {
		vector.FillRect(screen, tb.X, tb.Y, tb.W, tb.H, topBarBg, false)
		vector.StrokeRect(screen, tb.X, tb.Y, tb.W, tb.H, 1, topBarBorder, false)
	}
	topLine := battleToRect(layout.V2TopBarLine)
	if topLine.W > 0 && topLine.H > 0 && battle != nil {
		s := fmt.Sprintf("Round %d · %s", battle.Round, battle.PhaseString())
		if battle.Result != battlepkg.ResultNone {
			s = battle.ResultString() + " · SPACE/ENTER"
		} else if active := battle.ActiveUnit(); active != nil {
			s = fmt.Sprintf("Round %d · %s · %s", battle.Round, battle.PhaseString(), active.Name())
		}
		drawSingleLineInRect(screen, hudFace, topLine, fitTextToWidth(hudFace, s, topLine.W), metrics, color.RGBA{R: 195, G: 195, B: 200, A: 255})
	}
}
