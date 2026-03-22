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

func battleHUDTier(screenW, screenH int) ResolutionTier {
	return TierFromScreen(screenW, screenH)
}

// battleCompactUnavailableHintRU сжимает типовые RU-подсказки gate для small tier (доменная логика в battle не меняется).
func battleCompactUnavailableHintRU(tier ResolutionTier, msg string) string {
	if tier != TierSmall || msg == "" {
		return msg
	}
	msg = strings.TrimSpace(msg)
	var rem int
	if _, err := fmt.Sscanf(msg, "КД: ещё %d р.", &rem); err == nil {
		return fmt.Sprintf("КД %dр", rem)
	}
	need, have := 0, 0
	if n, err := fmt.Sscanf(msg, "Мана: нужно %d, сейчас %d", &need, &have); err == nil && n == 2 {
		return fmt.Sprintf("мана %d/%d", have, need)
	}
	if n, err := fmt.Sscanf(msg, "Энергия: нужно %d, сейчас %d", &need, &have); err == nil && n == 2 {
		return fmt.Sprintf("энерг. %d/%d", have, need)
	}
	return msg
}

func battleV1FooterControlsHintRU(battle *battlepkg.BattleContext, tier ResolutionTier) string {
	if tier == TierSmall {
		if active := battle.ActiveUnit(); active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
			if battle.PlayerTurn.SelectedAbilityID == battlepkg.AbilityBasicAttack {
				return "клик/Enter · Esc · ПКМ"
			}
			if battle.PlayerTurn.Phase == battlepkg.PlayerChooseTarget {
				return "цель · Enter · Esc · ПКМ"
			}
			return "способн. · Enter · Esc · ПКМ"
		}
		return "Esc · ПКМ"
	}
	controls := "ЛКМ/ПКМ · Esc: отступить · ПКМ по юниту — сведения"
	if active := battle.ActiveUnit(); active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
		if battle.PlayerTurn.SelectedAbilityID == battlepkg.AbilityBasicAttack {
			controls = "Стрелки+Enter или клик по врагу · Esc: отступить · ПКМ по юниту — сведения"
		} else if battle.PlayerTurn.Phase == battlepkg.PlayerChooseTarget {
			controls = "Стрелки+Enter или клик по цели · Назад/Esc: отмена · ПКМ по юниту — сведения"
		} else {
			controls = "Стрелки+Enter или клик по способности · Назад/Esc: отмена · ПКМ по юниту — сведения"
		}
	}
	return controls
}

func battleV2BottomHintRU(battle *battlepkg.BattleContext, tier ResolutionTier) string {
	if tier == TierSmall {
		active := battle.ActiveUnit()
		pt := &battle.PlayerTurn
		isDefaultAttack := active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction && pt.SelectedAbilityID == battlepkg.AbilityBasicAttack
		if isDefaultAttack {
			return "Enter / Esc · ПКМ"
		}
		if active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
			switch pt.Phase {
			case battlepkg.PlayerChooseTarget:
				return "←/→ · Enter · Esc · ПКМ"
			default:
				return "←/→ · Enter · Esc · ПКМ"
			}
		}
		return "Esc · ПКМ"
	}
	active := battle.ActiveUnit()
	pt := &battle.PlayerTurn
	isDefaultAttack := active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction && pt.SelectedAbilityID == battlepkg.AbilityBasicAttack
	if isDefaultAttack {
		return "Enter: выбор цели · стрелки · Enter: атака · Esc: отступить · ПКМ по юниту — сведения"
	}
	if active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
		switch pt.Phase {
		case battlepkg.PlayerChooseTarget:
			return "Стрелки: цель · Enter: выполнить · Назад/Esc: отмена · ПКМ по юниту — сведения"
		default:
			return "Стрелки: способность · Enter: выбрать · Назад/Esc: отмена · ПКМ по юниту — сведения"
		}
	}
	hint := ""
	if len(battle.BattleLog) > 0 {
		hint = strings.TrimSpace(battle.BattleLog[len(battle.BattleLog)-1])
	}
	if hint == "" {
		hint = "Esc: отступить"
	}
	return hint + " · ПКМ по юниту — сведения"
}

// abilityUnavailableStrokeColor — цвет рамки строки способности при блокировке (КД / мана / энергия).
func abilityUnavailableStrokeColor(code battlepkg.ValidationCode) color.RGBA {
	switch code {
	case battlepkg.ErrAbilityOnCooldown:
		return Theme.AbilityBlockCooldownBrd
	case battlepkg.ErrInsufficientMana:
		return Theme.AbilityBlockManaBrd
	case battlepkg.ErrInsufficientEnergy:
		return Theme.AbilityBlockEnergyBrd
	default:
		return Theme.AbilityBorder
	}
}

// drawBattleOverlayPanel рисует затемнённый фон и центральную панель боевого overlay.
func drawBattleOverlayPanel(screen *ebiten.Image, screenWidth, screenHeight int, layout battlepkg.BattleHUDLayout) rect {
	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), Theme.OverlayDim, false)

	ov := layout.Overlay
	vector.FillRect(screen, ov.X, ov.Y, ov.W, ov.H, Theme.PanelBG, false)
	vector.StrokeRect(screen, ov.X, ov.Y, ov.W, ov.H, battlePanelBorder, Theme.PanelBorder, false)
	return rect{X: ov.X, Y: ov.Y, W: ov.W, H: ov.H}
}

// drawBattleOverlayText рисует battle HUD v1: жёсткая сетка, только drawSingleLineInRect / drawLinesInRect.
func drawBattleOverlayText(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, layout battlepkg.BattleHUDLayout, screenW, screenH int) {
	metrics := layout.Metrics
	tier := battleHUDTier(screenW, screenH)

	// Top block hierarchy: title primary, info rows secondary.
	titleRow := battleToRect(layout.TitleRow)
	if titleRow.W > 0 && titleRow.H > 0 {
		title := "Бой"
		if battle != nil && len(battle.Encounter.Enemies) > 0 {
			title = fmt.Sprintf("Бой: враг #%d", battle.Encounter.Enemies[0].EnemyID)
		}
		drawSingleLineInRect(screen, hudFace, titleRow, title, metrics, Theme.TextPrimary)
	}

	if battle == nil {
		return
	}

	if battle.Result != battlepkg.ResultNone {
		info1 := battleToRect(layout.InfoRow1)
		var banner string
		switch battle.Result {
		case battlepkg.ResultVictory:
			banner = "ПОБЕДА"
		case battlepkg.ResultDefeat:
			banner = "ПОРАЖЕНИЕ"
		case battlepkg.ResultEscape:
			banner = "ОТСТУПЛЕНИЕ"
		default:
			banner = battle.ResultString()
		}
		drawSingleLineInRect(screen, hudFace, info1, banner, metrics, Theme.TextSecondary)
		info2 := battleToRect(layout.InfoRow2)
		drawSingleLineInRect(screen, hudFace, info2, "Пробел / Enter — продолжить", metrics, Theme.TextMuted)
	} else {
		info1 := battleToRect(layout.InfoRow1)
		line1 := fmt.Sprintf("Раунд %d · фаза: %s", battle.Round, battle.PhaseLabelRU())
		drawSingleLineInRect(screen, hudFace, info1, line1, metrics, Theme.TextSecondary)

		info2 := battleToRect(layout.InfoRow2)
		activeStr := battle.DisplayPhaseLabel()
		if active := battle.ActiveUnit(); active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
			activeStr = fmt.Sprintf("%s | %s", activeStr, battle.PlayerTurn.PhaseLabelRU())
			activeStr = fmt.Sprintf("%s | %s", activeStr, battlepkg.ActorResourceLineRU(active))
		}
		var info2Line string
		if tier == TierSmall {
			phaseOnly := battle.DisplayPhaseLabel()
			if active := battle.ActiveUnit(); active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
				phaseOnly = fmt.Sprintf("%s | %s", phaseOnly, battle.PlayerTurn.PhaseLabelRU())
			}
			info2Line = CompactLine(hudFace, activeStr, phaseOnly, info2.W)
		} else {
			info2Line = fitTextToWidth(hudFace, activeStr, info2.W)
		}
		drawSingleLineInRect(screen, hudFace, info2, info2Line, metrics, Theme.TextMuted)
	}

	footerRect := battleToRect(layout.Footer)
	playerPanel := battleToRect(layout.PlayerFormation)
	enemyPanel := battleToRect(layout.EnemyFormation)

	drawFormationPanel(screen, hudFace, battle, playerPanel, battlepkg.BattleSidePlayer, "СОЮЗНИКИ", layout)
	drawFormationPanel(screen, hudFace, battle, enemyPanel, battlepkg.BattleSideEnemy, "ВРАГИ", layout)

	abilitiesRect := battleToRect(layout.Abilities)
	confirmRect := battleToRect(layout.Action)

	drawAbilityPanel(screen, hudFace, battle, abilitiesRect, layout, tier)
	drawConfirmPanel(screen, hudFace, battle, confirmRect, layout, tier)

	drawFooterPanel(screen, hudFace, battle, footerRect, layout, tier)

	DrawBattleFeedbackFloats(screen, hudFace, battle, layout, metrics)
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
	if isPlayerTurn && pt.Phase == battlepkg.PlayerChooseTarget {
		for _, td := range pt.ValidTargets {
			if td.Kind == battlepkg.TargetKindUnit {
				validSet[td.UnitID] = true
			}
		}
	}

	selectedTargetID := battlepkg.UnitID(0)
	if isPlayerTurn && pt.Phase == battlepkg.PlayerChooseTarget && pt.SelectedTarget.Kind == battlepkg.TargetKindUnit {
		selectedTargetID = pt.SelectedTarget.UnitID
	}
	hoverTargetID := battlepkg.UnitID(0)
	if isPlayerTurn && pt.Phase == battlepkg.PlayerChooseTarget {
		hoverTargetID = pt.HoverTargetUnitID
	}

	labelH := metrics.LineH
	drawRowLabel := func(label string, y float32) {
		row := rect{X: inner.X, Y: y, W: inner.W, H: labelH}
		drawSingleLineInRect(screen, hudFace, row, label, metrics, Theme.TextMuted)
	}

	frontLabelY := inner.Y
	frontSlotsY := frontLabelY + labelH
	backLabelY := inner.Y + (inner.H-labelH*2)*0.5
	backSlotsY := backLabelY + labelH

	drawRowLabel("ПЕРЕД", frontLabelY)
	drawRowLabel("ЗАД", backLabelY)

	enemySide := side == battlepkg.BattleSideEnemy
	drawSlot := func(row battlepkg.BattleRow, idx int, x, y float32) {
		slot := battle.Slot(side, row, idx)
		var u *battlepkg.BattleUnit
		if slot != nil {
			u = battle.UnitInSlot(slot)
		}

		fill := Theme.PanelBGDeep
		border := Theme.AllyAccent
		if enemySide {
			border = Theme.EnemyAccent
		}
		textCol := Theme.TextPrimary

		if u == nil {
			fill = Theme.EmptySlot
			textCol = Theme.TextMuted
		} else if !u.IsAlive() {
			fill = Theme.DeadFill
			textCol = Theme.DeadText
		}

		if u != nil && u.ID == selectedTargetID {
			border = Theme.SelectedKill
		} else if u != nil && u.ID == hoverTargetID {
			border = Theme.HoverTarget
		} else if u != nil && validSet[u.ID] {
			border = Theme.ValidTarget
		} else if active != nil && u != nil && u.ID == active.ID {
			border = Theme.ActiveTurn
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
		if u != nil {
			if k, in := battle.FeedbackFlashIntensity(u.ID); k >= 0 && in > 0 {
				drawFeedbackOverlayRect(screen, rect{X: x, Y: y, W: w, H: h}, k, in)
			}
		}

		line1 := "ПУСТО"
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
				line2 = "погиб"
			} else {
				line1 = name
				line2 = "ОЗ " + hp
			}
		}

		row1 := rect{X: x + 6, Y: y, W: w - 6, H: metrics.LineH}
		drawSingleLineInRect(screen, hudFace, row1, line1, metrics, textCol)
		if line2 != "" {
			row2 := rect{X: x + 6, Y: y + metrics.LineH, W: w - 6, H: metrics.LineH}
			drawSingleLineInRect(screen, hudFace, row2, line2, metrics, textCol)
		}
		if u != nil && u.IsAlive() && w > 8 && h > metrics.LineH*2+4 {
			barY := y + h - 6
			DrawHPBarMicro(screen, x+4, barY, w-8, 4, u.State.HP, u.MaxHP(), true, enemySide)
		}
	}

	for i := 0; i < 3; i++ {
		x := inner.X + float32(i)*((inner.W-metrics.Gap*2)/3)
		drawSlot(battlepkg.BattleRowFront, i, x, frontSlotsY)
		drawSlot(battlepkg.BattleRowBack, i, x, backSlotsY)
	}
}

func drawAbilityPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout, tier ResolutionTier) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, battleToRect(layout.AbilitiesTitleRow), "СПОСОБНОСТИ", hudFace, metrics)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		titleRow := battleToRect(layout.AbilitiesTitleRow)
		if titleRow.W > 0 && titleRow.H > 0 {
			drawSingleLineInRect(screen, hudFace, titleRow, "(ожидание)", metrics, Theme.TextMuted)
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
		prefix := " "
		gate := battlepkg.AbilityResourceGate(battle, active, id)
		col := Theme.TextPrimary
		colMuted := Theme.TextMuted
		bg := Theme.AbilityBG
		border := Theme.AbilityBorder
		if !gate.OK {
			bg = Theme.PanelBGDeep
			col = Theme.TextSecondary
			colMuted = Theme.TextMuted
		}
		if id == sel && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			prefix = "▶"
			bg = Theme.AbilitySelectedBG
			col = Theme.TextPrimary
			colMuted = Theme.TextSecondary
			border = Theme.AbilitySelectedBrd
		} else if hoverIdx == i && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			bg = Theme.AbilityHoverBG
			if gate.OK {
				col = Theme.TextSecondary
			}
			border = Theme.HoverTarget
		}
		vector.FillRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, bg, false)
		strokeW := float32(1)
		strokeCol := border
		if !gate.OK {
			strokeW = 2
			strokeCol = abilityUnavailableStrokeColor(gate.Code)
		}
		vector.StrokeRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, strokeW, strokeCol, false)

		nameLine := fmt.Sprintf("%s %s", prefix, battlepkg.PlayerAbilityLabelRU(id))
		nameLine = PrimaryLine(hudFace, nameLine, rowRect.W-12)
		lineH := metrics.LineH
		topRow := rect{X: rowRect.X + 6, Y: rowRect.Y + 2, W: rowRect.W - 12, H: lineH}
		drawSingleLineInRect(screen, hudFace, topRow, nameLine, metrics, col)
		if gate.OK {
			if cost := battlepkg.AbilityCostLinePlayerRU(battle, active, id); cost != "" {
				cost = fitTextToWidth(hudFace, cost, rowRect.W-12)
				botRow := rect{X: rowRect.X + 6, Y: rowRect.Y + lineH + 2, W: rowRect.W - 12, H: lineH}
				drawSingleLineInRect(screen, hudFace, botRow, cost, metrics, Theme.TextSecondary)
			}
		} else if msg := battlepkg.AbilityUnavailableHintRU(battle, active, id); msg != "" {
			msg = battleCompactUnavailableHintRU(tier, msg)
			msg = PrimaryLine(hudFace, msg, rowRect.W-12)
			botRow := rect{X: rowRect.X + 6, Y: rowRect.Y + lineH + 2, W: rowRect.W - 12, H: lineH}
			drawSingleLineInRect(screen, hudFace, botRow, msg, metrics, colMuted)
		}
	}
}

func drawConfirmPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout, tier ResolutionTier) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, battleToRect(layout.ActionTitleRow), "ХОД", hudFace, metrics)
	if battle == nil {
		return
	}

	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		summaryRect := battleToRect(layout.ActionSummary)
		if summaryRect.W > 0 && summaryRect.H > 0 {
			drawSingleLineInRect(screen, hudFace, summaryRect, "(ход врага)", metrics, Theme.TextMuted)
		}
		return
	}

	pt := battle.PlayerTurn

	targetStr := battleActionTargetLabelRU(&pt, battle, tier)

	summaryRect := battleToRect(layout.ActionSummary)
	// Стоимость способности — только в списке способностей; ресурсы актёра — в верхней строке InfoRow2 (не дублируем).
	var summaryLines []string
	if tier == TierSmall {
		summaryLines = []string{
			fmt.Sprintf("Способность: %s", battlepkg.PlayerAbilityLabelRU(pt.SelectedAbilityID)),
			fmt.Sprintf("Цель: %s", targetStr),
		}
	} else {
		summaryLines = []string{
			fmt.Sprintf("Шаг: %s", pt.PhaseLabelRU()),
			fmt.Sprintf("Способность: %s", battlepkg.PlayerAbilityLabelRU(pt.SelectedAbilityID)),
			fmt.Sprintf("Цель: %s", targetStr),
		}
	}

	// Превью урона/лечения — если под панелью остаётся строка (после базового блока).
	if maxLinesForRect(metrics, summaryRect, 0, 0, metrics.LineH) >= len(summaryLines)+1 {
		req := battlepkg.ActionRequest{Actor: active.ID, Ability: pt.SelectedAbilityID, Target: pt.SelectedTarget}
		preview, v := battlepkg.PreviewAction(battle, req)
		if v.OK && (preview.HasDamage() || preview.HasHeal()) {
			var previewStr string
			if preview.HasDamage() {
				previewStr = fmt.Sprintf("Вид: урон %d–%d", preview.DamageMin, preview.DamageMax)
			} else {
				previewStr = fmt.Sprintf("Вид: лечение %d–%d", preview.HealMin, preview.HealMax)
			}
			summaryLines = append(summaryLines, previewStr)
		}
	}

	maxSummaryW := summaryRect.W
	for i := range summaryLines {
		summaryLines[i] = fitTextToWidth(hudFace, summaryLines[i], maxSummaryW)
	}
	_ = drawLinesInRect(screen, hudFace, summaryRect, summaryLines, metrics, Theme.TextPrimary, 0)

	backR := battleToRect(layout.BackButton)
	drawButton := func(r rect, label string, enabled, hovered bool) {
		baseFill := Theme.ButtonBG
		baseBorder := Theme.ButtonBorder
		textCol := Theme.TextPrimary
		if !enabled {
			baseFill = Theme.PanelBGDeep
			baseBorder = Theme.PanelBorder
			textCol = Theme.DisabledFG
		}
		if enabled && hovered {
			baseFill = Theme.ButtonHoverBG
			baseBorder = Theme.ButtonHoverBorder
		}
		vector.FillRect(screen, r.X, r.Y, r.W, r.H, baseFill, false)
		vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 2, baseBorder, false)

		drawSingleLineInRect(screen, hudFace, r, label, metrics, textCol)
	}
	// Back only (Confirm removed from battle UX)
	if backR.W > 0 && backR.H > 0 {
		drawButton(backR, "Назад", true, pt.HoverBackButton)
	}
}

func drawFooterPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout, tier ResolutionTier) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, battleToRect(layout.FooterTitleRow), "ЛОГ БОЯ", hudFace, metrics)
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
		_ = drawLinesInRect(screen, hudFace, logRect, logLines, metrics, Theme.TextSecondary, 0)
	}

	if controlsRect.W > 0 && controlsRect.H > 0 {
		controls := battleV1FooterControlsHintRU(battle, tier)
		drawSingleLineInRect(screen, hudFace, controlsRect, fitTextToWidth(hudFace, controls, controlsRect.W), metrics, Theme.TextMuted)
	}
}

// drawBattleScreenV2 рисует battle screen в стиле Disciples: центр = сцена, слева/справа ростеры, внизу панель команд, сверху минимальный TopBar.
func drawBattleScreenV2(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, layout battlepkg.BattleHUDLayout, inspectOpenID battlepkg.UnitID, inspectOpen bool, screenW, screenH int) {
	metrics := layout.Metrics
	tier := battleHUDTier(screenW, screenH)

	// 1) Battlefield (center) — сцена: рамка в духе inspect/roster, внутренняя подложка, лёгкое затемнение по краям.
	bf := layout.Battlefield
	if bf.W > 0 && bf.H > 0 {
		vector.FillRect(screen, bf.X, bf.Y, bf.W, bf.H, Theme.SceneTint, false)
		insetPx := float32(14)
		if bf.W > insetPx*2+16 && bf.H > insetPx*2+16 {
			cx, cy := bf.X+insetPx, bf.Y+insetPx
			cw, ch := bf.W-insetPx*2, bf.H-insetPx*2
			vector.FillRect(screen, cx, cy, cw, ch, Theme.PanelBGDeep, false)
			vector.StrokeRect(screen, cx, cy, cw, ch, 1, Theme.RosterCardInnerStroke, false)
			DrawThinAccentLine(screen, cx+4, cy+4, cw-8)
		}
		vector.StrokeRect(screen, bf.X, bf.Y, bf.W, bf.H, 2, Theme.PostBattleBorder, false)
		vector.StrokeRect(screen, bf.X+3, bf.Y+3, bf.W-6, bf.H-6, 1, Theme.PanelBorder, false)
		v := float32(7)
		if bf.W > v*2+4 && bf.H > v*2+4 {
			vector.FillRect(screen, bf.X, bf.Y, bf.W, v, Theme.BattlefieldSceneVignette, false)
			vector.FillRect(screen, bf.X, bf.Y+bf.H-v, bf.W, v, Theme.BattlefieldSceneVignette, false)
			vector.FillRect(screen, bf.X, bf.Y+v, v, bf.H-v*2, Theme.BattlefieldSceneVignette, false)
			vector.FillRect(screen, bf.X+bf.W-v, bf.Y+v, v, bf.H-v*2, Theme.BattlefieldSceneVignette, false)
		}
	}

	DrawBattlefieldV2Scene(screen, hudFace, battle, layout, inspectOpenID, inspectOpen)

	// 2) Left / Right rosters — боковые панели, визуально отделены от сцены
	lr := layout.LeftRoster
	if lr.W > 0 && lr.H > 0 {
		vector.FillRect(screen, lr.X, lr.Y, lr.W, lr.H, Theme.PanelBG, false)
		vector.StrokeRect(screen, lr.X, lr.Y, lr.W, lr.H, 1, Theme.PanelBorder, false)
		DrawThinAccentLine(screen, lr.X+4, lr.Y+4, lr.W-8)
		lab := rect{X: lr.X + 6, Y: lr.Y + 8, W: lr.W - 12, H: metrics.LineH * 0.95}
		drawSingleLineInRect(screen, hudFace, lab, fitTextToWidth(hudFace, "СОЮЗНИКИ · перед→зад", lab.W), metrics, Theme.TextSecondary)
	}
	rr := layout.RightRoster
	if rr.W > 0 && rr.H > 0 {
		vector.FillRect(screen, rr.X, rr.Y, rr.W, rr.H, Theme.PanelBG, false)
		vector.StrokeRect(screen, rr.X, rr.Y, rr.W, rr.H, 1, Theme.PanelBorder, false)
		DrawThinAccentLine(screen, rr.X+4, rr.Y+4, rr.W-8)
		lab := rect{X: rr.X + 6, Y: rr.Y + 8, W: rr.W - 12, H: metrics.LineH * 0.95}
		drawSingleLineInRect(screen, hudFace, lab, fitTextToWidth(hudFace, "ВРАГИ · перед→зад", lab.W), metrics, Theme.TextSecondary)
	}

	// 3) Unit cards in rosters
	for id, hr := range layout.UnitRects {
		u := battle.Units[id]
		if u == nil {
			continue
		}
		drawBattleRosterUnitCard(screen, hudFace, battle, u, hr, metrics, inspectOpenID, inspectOpen)
	}

	// 4) Bottom panel — control panel: Active | Target → Abilities → Summary → Hint → Buttons
	bp := layout.BottomPanel
	if bp.W > 0 && bp.H > 0 {
		vector.FillRect(screen, bp.X, bp.Y, bp.W, bp.H, Theme.PanelBG, false)
		vector.StrokeRect(screen, bp.X, bp.Y, bp.W, bp.H, 1, Theme.PanelBorder, false)
	}
	activeR := battleToRect(layout.V2BottomActive)
	if activeR.W > 0 && activeR.H > 0 && battle != nil {
		active := battle.ActiveUnit()
		if active != nil && active.Side == battlepkg.TeamPlayer {
			l1 := fmt.Sprintf("Ваш ход: %s%s", active.Name(), battlepkg.PlayerTemplateIdentitySuffix(active))
			if tier == TierSmall {
				l1 = fmt.Sprintf("▶ %s%s", active.Name(), battlepkg.PlayerTemplateIdentitySuffix(active))
			}
			row1 := rect{X: activeR.X, Y: activeR.Y, W: activeR.W, H: metrics.LineH}
			drawSingleLineInRect(screen, hudFace, row1, fitTextToWidth(hudFace, l1, activeR.W), metrics, Theme.TextPrimary)
			barY := activeR.Y + metrics.LineH + 2
			if tier != TierSmall {
				row2 := rect{X: activeR.X, Y: activeR.Y + metrics.LineH, W: activeR.W, H: metrics.LineH}
				drawSingleLineInRect(screen, hudFace, row2, fitTextToWidth(hudFace, battlepkg.ActorResourceLineRU(active), activeR.W), metrics, Theme.TextSecondary)
				barY = activeR.Y + metrics.LineH*2 + 2
			}
			bw := activeR.W - 8
			if bw > 24 {
				DrawResourceBarMicro(screen, activeR.X+4, barY, bw, 3, active.State.Mana, active.State.MaxMana, Theme.ResourceManaFill)
				DrawResourceBarMicro(screen, activeR.X+4, barY+5, bw, 3, active.State.Energy, active.State.MaxEnergy, Theme.ResourceEnergyFill)
			}
		} else if active != nil {
			s := fmt.Sprintf("Ход врага: %s", active.Name())
			drawSingleLineInRect(screen, hudFace, activeR, fitTextToWidth(hudFace, s, activeR.W), metrics, Theme.TextPrimary)
		} else {
			drawSingleLineInRect(screen, hudFace, activeR, "—", metrics, Theme.TextPrimary)
		}
	}
	targetR := battleToRect(layout.V2BottomTarget)
	if targetR.W > 0 && targetR.H > 0 && battle != nil {
		pt := &battle.PlayerTurn
		isDefaultAttack := pt.Phase == battlepkg.PlayerChooseAbility && pt.SelectedAbilityID == battlepkg.AbilityBasicAttack
		s := "—"
		if isDefaultAttack && pt.HoverTargetUnitID != 0 && battle.Units[pt.HoverTargetUnitID] != nil {
			s = battle.Units[pt.HoverTargetUnitID].Name()
		} else if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && battle.Units[pt.SelectedTarget.UnitID] != nil {
			s = battle.Units[pt.SelectedTarget.UnitID].Name()
		} else if pt.SelectedTarget.Kind == battlepkg.TargetKindSelf {
			s = "себя"
		} else if pt.Phase == battlepkg.PlayerChooseAbility && pt.SelectedAbilityID == battlepkg.AbilityGroupHeal {
			s = "все союзники"
		}
		drawSingleLineInRect(screen, hudFace, targetR, fitTextToWidth(hudFace, s, targetR.W), metrics, Theme.TextSecondary)
	}
	// Ability row (special abilities; стоимость и блокировка)
	activeUnit := battle.ActiveUnit()
	for i, hr := range layout.AbilityItemRects {
		rowRect := battleToRect(hr)
		abs := []battlepkg.AbilityID{}
		if activeUnit != nil {
			abs = battlepkg.SpecialAbilities(activeUnit)
		}
		if i >= len(abs) {
			break
		}
		id := abs[i]
		gate := battlepkg.AbilityResourceGate(battle, activeUnit, id)
		col := Theme.TextPrimary
		col2 := Theme.TextMuted
		bg := Theme.AbilityBG
		brd := Theme.AbilityBorder
		if !gate.OK {
			bg = Theme.PanelBGDeep
			col = Theme.TextSecondary
		}
		if battle.PlayerTurn.SelectedAbilityID == id {
			col = Theme.TextPrimary
			col2 = Theme.TextSecondary
			bg = Theme.AbilitySelectedBG
			brd = Theme.AbilitySelectedBrd
		}
		if battle.PlayerTurn.HoverAbilityIndex == i {
			brd = Theme.HoverTarget
		}
		vector.FillRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, bg, false)
		strokeW := float32(1)
		strokeCol := brd
		if !gate.OK {
			strokeW = 2
			strokeCol = abilityUnavailableStrokeColor(gate.Code)
		}
		vector.StrokeRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, strokeW, strokeCol, false)
		lh := metrics.LineH
		label := PrimaryLine(hudFace, battlepkg.PlayerAbilityLabelRU(id), rowRect.W-8)
		r1 := rect{X: rowRect.X + 4, Y: rowRect.Y + 2, W: rowRect.W - 8, H: lh}
		drawSingleLineInRect(screen, hudFace, r1, label, metrics, col)
		if gate.OK {
			if c := battlepkg.AbilityCostLinePlayerRU(battle, activeUnit, id); c != "" {
				r2 := rect{X: rowRect.X + 4, Y: rowRect.Y + lh + 2, W: rowRect.W - 8, H: lh}
				drawSingleLineInRect(screen, hudFace, r2, fitTextToWidth(hudFace, c, rowRect.W-8), metrics, Theme.TextSecondary)
			}
		} else if msg := battlepkg.AbilityUnavailableHintRU(battle, activeUnit, id); msg != "" {
			msg = battleCompactUnavailableHintRU(tier, msg)
			r2 := rect{X: rowRect.X + 4, Y: rowRect.Y + lh + 2, W: rowRect.W - 8, H: lh}
			drawSingleLineInRect(screen, hudFace, r2, fitTextToWidth(hudFace, msg, rowRect.W-8), metrics, col2)
		}
	}
	// Summary — default attack: краткая подсказка; special: способность (и «→ цель» если не дублирует колонку цели) + превью; стоимость — только в ряду способностей.
	summaryR := battleToRect(layout.V2BottomSummary)
	if summaryR.W > 0 && summaryR.H > 0 && battle != nil {
		active := battle.ActiveUnit()
		lines := []string{}
		if active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
			pt := &battle.PlayerTurn
			if pt.SelectedAbilityID == battlepkg.AbilityBasicAttack {
				if tier == TierSmall {
					lines = append(lines, PrimaryLine(hudFace, fmt.Sprintf("%s · клик", battlepkg.PlayerAbilityLabelRU(battlepkg.AbilityBasicAttack)), summaryR.W))
				} else {
					lines = append(lines, fitTextToWidth(hudFace, fmt.Sprintf("%s · клик по врагу", battlepkg.PlayerAbilityLabelRU(battlepkg.AbilityBasicAttack)), summaryR.W))
				}
				if pt.HoverTargetUnitID != 0 && battle.Units[pt.HoverTargetUnitID] != nil {
					preview, v := battlepkg.PreviewAction(battle, battlepkg.ActionRequest{Actor: active.ID, Ability: battlepkg.AbilityBasicAttack, Target: battlepkg.UnitTarget(pt.HoverTargetUnitID)})
					if v.OK && preview.HasDamage() {
						lines = append(lines, fmt.Sprintf("урон %d–%d", preview.DamageMin, preview.DamageMax))
					}
				}
			} else {
				targetStr := "—"
				if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && battle.Units[pt.SelectedTarget.UnitID] != nil {
					targetStr = battle.Units[pt.SelectedTarget.UnitID].Name()
				} else if pt.SelectedTarget.Kind == battlepkg.TargetKindSelf {
					targetStr = "себя"
				} else if pt.SelectedTarget.Kind == battlepkg.TargetKindNone && pt.SelectedAbilityID == battlepkg.AbilityGroupHeal {
					targetStr = "все союзники"
				}
				showArrow := targetStr != "—"
				if tier == TierSmall {
					switch pt.SelectedTarget.Kind {
					case battlepkg.TargetKindUnit, battlepkg.TargetKindSelf:
						showArrow = false
					case battlepkg.TargetKindNone:
						if pt.SelectedAbilityID == battlepkg.AbilityGroupHeal {
							showArrow = false
						}
					}
				}
				line1 := ""
				if showArrow {
					line1 = fmt.Sprintf("%s → %s", battlepkg.PlayerAbilityLabelRU(pt.SelectedAbilityID), targetStr)
				} else {
					line1 = battlepkg.PlayerAbilityLabelRU(pt.SelectedAbilityID)
				}
				lines = append(lines, fitTextToWidth(hudFace, line1, summaryR.W))
				req := battlepkg.ActionRequest{Actor: active.ID, Ability: pt.SelectedAbilityID, Target: pt.SelectedTarget}
				preview, v := battlepkg.PreviewAction(battle, req)
				if v.OK && (preview.HasDamage() || preview.HasHeal()) {
					if preview.HasDamage() {
						lines = append(lines, fmt.Sprintf("урон %d–%d", preview.DamageMin, preview.DamageMax))
					} else {
						lines = append(lines, fmt.Sprintf("лечение %d–%d", preview.HealMin, preview.HealMax))
					}
				}
			}
		} else {
			// Пауза анимации, ход врага, TurnStart и т.д. — одна строка из battle (имя текущего юнита).
			lines = append(lines, fitTextToWidth(hudFace, battle.DisplayPhaseLabel(), summaryR.W))
		}
		if len(lines) > 2 {
			lines = lines[:2]
		}
		_ = drawLinesInRect(screen, hudFace, summaryR, lines, metrics, Theme.TextPrimary, 2)
	}
	// Hint — tier-aware (полные подсказки на medium/large).
	logR := battleToRect(layout.V2BottomLog)
	if logR.W > 0 && logR.H > 0 && battle != nil {
		hint := battleV2BottomHintRU(battle, tier)
		drawSingleLineInRect(screen, hudFace, logR, fitTextToWidth(hudFace, hint, logR.W), metrics, Theme.TextMuted)
	}
	// Buttons — Back only (Confirm removed from battle UX)
	backR := battleToRect(layout.BackButton)
	pt := &battle.PlayerTurn
	drawButton := func(r rect, label string, enabled, hovered bool) {
		fill := Theme.ButtonBG
		brd := Theme.ButtonBorder
		tcol := Theme.TextPrimary
		if !enabled {
			fill = Theme.PanelBGDeep
			tcol = Theme.DisabledFG
		}
		if enabled && hovered {
			fill = Theme.ButtonHoverBG
			brd = Theme.ButtonHoverBorder
		}
		vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
		vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 2, brd, false)
		drawSingleLineInRect(screen, hudFace, r, label, metrics, tcol)
	}
	if backR.W > 0 && backR.H > 0 {
		drawButton(backR, "Назад", true, pt.HoverBackButton)
	}

	// 5) TopBar — лёгкая status line, не перебивает сцену
	tb := layout.TopBar
	if tb.W > 0 && tb.H > 0 {
		vector.FillRect(screen, tb.X, tb.Y, tb.W, tb.H, Theme.PanelBGDeep, false)
		vector.StrokeRect(screen, tb.X, tb.Y, tb.W, tb.H, 1, Theme.PanelBorder, false)
	}
	topLine := battleToRect(layout.V2TopBarLine)
	if topLine.W > 0 && topLine.H > 0 && battle != nil {
		var s string
		if battle.Result != battlepkg.ResultNone {
			s = battle.ResultString() + " · Пробел/Enter"
		} else {
			s = fmt.Sprintf("Раунд %d · %s", battle.Round, battle.DisplayPhaseLabel())
		}
		drawSingleLineInRect(screen, hudFace, topLine, fitTextToWidth(hudFace, s, topLine.W), metrics, Theme.TextSecondary)
	}
}

// battleActionTargetLabelRU — краткая подпись цели для панели «Ход» (v1 summary).
func battleActionTargetLabelRU(pt *battlepkg.PlayerTurnState, bctx *battlepkg.BattleContext, tier ResolutionTier) string {
	if pt == nil || bctx == nil {
		return "—"
	}
	switch pt.SelectedTarget.Kind {
	case battlepkg.TargetKindSelf:
		return "себя"
	case battlepkg.TargetKindUnit:
		if tu := bctx.Units[pt.SelectedTarget.UnitID]; tu != nil {
			if tier == TierSmall {
				return tu.Name()
			}
			return fmt.Sprintf("%s (#%d)", tu.Name(), tu.ID)
		}
		return fmt.Sprintf("юнит #%d", pt.SelectedTarget.UnitID)
	case battlepkg.TargetKindNone:
		return "нет"
	default:
		return "—"
	}
}
