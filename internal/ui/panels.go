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

const (
	uiLineH       = float32(18)
	uiPad         = float32(12)
	uiGap         = float32(10)
	uiPanelBorder = float32(2)
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampF(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func inset(r rect, pad float32) rect {
	n := rect{X: r.X + pad, Y: r.Y + pad, W: r.W - pad*2, H: r.H - pad*2}
	if n.W < 0 {
		n.W = 0
	}
	if n.H < 0 {
		n.H = 0
	}
	return n
}

func splitH(r rect, leftW, gap float32) (rect, rect) {
	left := rect{X: r.X, Y: r.Y, W: leftW, H: r.H}
	right := rect{X: r.X + leftW + gap, Y: r.Y, W: r.W - leftW - gap, H: r.H}
	if right.W < 0 {
		right.W = 0
	}
	return left, right
}

func splitV(r rect, topH, gap float32) (rect, rect) {
	top := rect{X: r.X, Y: r.Y, W: r.W, H: topH}
	bot := rect{X: r.X, Y: r.Y + topH + gap, W: r.W, H: r.H - topH - gap}
	if bot.H < 0 {
		bot.H = 0
	}
	return top, bot
}

// drawBattleOverlayPanel рисует затемнённый фон и центральную панель боевого overlay.
func drawBattleOverlayPanel(screen *ebiten.Image, screenWidth, screenHeight int, layout battlepkg.BattleHUDLayout) rect {
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	panelColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	panelBorderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)

	ov := layout.Overlay
	vector.FillRect(screen, ov.X, ov.Y, ov.W, ov.H, panelColor, false)
	vector.StrokeRect(screen, ov.X, ov.Y, ov.W, ov.H, uiPanelBorder, panelBorderColor, false)
	return rect{X: ov.X, Y: ov.Y, W: ov.W, H: ov.H}
}

func toRect(hr battlepkg.HUDRect) rect {
	return rect{X: hr.X, Y: hr.Y, W: hr.W, H: hr.H}
}

// drawBattleOverlayText рисует все текстовые блоки боевого overlay.
func drawBattleOverlayText(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, layout battlepkg.BattleHUDLayout) {
	panel := toRect(layout.Overlay)
	panelX := panel.X
	panelY := panel.Y

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(panelX)+float64(uiPad), float64(panelY)+float64(uiPad)+float64(uiLineH))
	titleOp.ColorScale.ScaleWithColor(color.White)

	title := "Battle mode"
	if battle != nil && len(battle.Encounter.Enemies) > 0 {
		title = fmt.Sprintf("Battle mode: enemy #%d", battle.Encounter.Enemies[0].EnemyID)
	}

	text.Draw(screen, title, hudFace, titleOp)

	if battle == nil {
		return
	}

	if battle.Result != battlepkg.ResultNone {
		bannerY := float64(panelY) + float64(uiPad) + float64(uiLineH)*2
		bannerOp := &text.DrawOptions{}
		bannerOp.GeoM.Translate(float64(panelX)+float64(uiPad), bannerY)
		bannerOp.ColorScale.ScaleWithColor(color.White)
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
		text.Draw(screen, banner, hudFace, bannerOp)

		hintOp := &text.DrawOptions{}
		hintOp.GeoM.Translate(float64(panelX)+float64(uiPad), bannerY+float64(uiLineH))
		hintOp.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
		text.Draw(screen, "SPACE/ENTER: continue", hudFace, hintOp)
	}

	// Layout: container-based inside overlay, driven by shared layout.
	content := toRect(layout.Content)

	// Info line.
	infoOp := &text.DrawOptions{}
	infoOp.GeoM.Translate(float64(content.X), float64(content.Y+uiLineH))
	infoOp.ColorScale.ScaleWithColor(color.White)

	active := battle.ActiveUnit()
	activeLabel := "-"
	if active != nil {
		activeLabel = fmt.Sprintf("%s (#%d)", active.Name(), active.ID)
		if active.Side == battlepkg.TeamPlayer {
			activeLabel += " [PLAYER]"
		} else {
			activeLabel += " [ENEMY]"
		}
	}
	playerSub := ""
	if active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
		playerSub = fmt.Sprintf(" | player:%s", battle.PlayerTurn.PhaseString())
	}
	text.Draw(screen, fmt.Sprintf("Round %d | phase:%s%s | active: %s", battle.Round, battle.PhaseString(), playerSub, activeLabel), hudFace, infoOp)

	// Vertical packing: info row + formation + middle + footer — из layout.
	footerRect := toRect(layout.Footer)

	playerPanel := toRect(layout.PlayerFormation)
	enemyPanel := toRect(layout.EnemyFormation)

	drawFormationPanel(screen, hudFace, battle, playerPanel, battlepkg.BattleSidePlayer, "PLAYER", layout)
	drawFormationPanel(screen, hudFace, battle, enemyPanel, battlepkg.BattleSideEnemy, "ENEMY", layout)

	abilitiesRect := toRect(layout.Abilities)
	confirmRect := toRect(layout.Action)

	drawAbilityPanel(screen, hudFace, battle, abilitiesRect, layout)
	drawConfirmPanel(screen, hudFace, battle, confirmRect, layout)

	drawFooterPanel(screen, hudFace, battle, footerRect)
}

// drawHUDText рисует текстовые блоки HUD (счётчик собранных предметов и т.п.).
func drawHUDText(screen *ebiten.Image, pickupCount int, hudFace *text.GoTextFace) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, 20)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Pickups: %d", pickupCount), hudFace, op)
}

type rect struct {
	X, Y, W, H float32
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func drawPanelBox(screen *ebiten.Image, r rect, title string, hudFace *text.GoTextFace) {
	bg := color.RGBA{R: 28, G: 28, B: 28, A: 255}
	border := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, bg, false)
	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, border, false)

	if title == "" {
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(r.X)+float64(uiPad*0.6), float64(r.Y)+float64(uiLineH))
	op.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	text.Draw(screen, title, hudFace, op)
}

func drawFormationPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, side battlepkg.BattleSide, title string, layout battlepkg.BattleHUDLayout) {
	drawPanelBox(screen, r, title, hudFace)
	if battle == nil {
		return
	}
	inner := inset(r, uiPad*0.6)
	inner.Y += uiLineH
	inner.H -= uiLineH
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

	// Slot grid labels (approximate; per-unit rects come from shared layout).
	labelH := uiLineH
	drawRowLabel := func(label string, y float32) {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(inner.X), float64(y+labelH))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 170, G: 170, B: 170, A: 255})
		text.Draw(screen, label, hudFace, op)
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

		// Base style.
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

		// Visual priority: selected > hovered > valid > active > normal.
		if u != nil && u.ID == selectedTargetID {
			border = color.RGBA{R: 240, G: 80, B: 80, A: 255}
		} else if u != nil && u.ID == hoverTargetID {
			border = color.RGBA{R: 120, G: 190, B: 255, A: 255}
		} else if u != nil && validSet[u.ID] {
			border = color.RGBA{R: 80, G: 150, B: 255, A: 255}
		} else if active != nil && u != nil && u.ID == active.ID {
			border = color.RGBA{R: 255, G: 215, B: 80, A: 255}
		}

		// Use shared layout for unit-bearing slots to match mouse hit-areas.
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
			// Fallback to approximate grid when we don't have a unit rect (e.g. empty slot).
			w = (inner.W - uiGap*2) / 3
			h = clampF((inner.H-labelH*2-uiGap*0.6)/2, uiLineH*2.4, uiLineH*3.5)
		}
		vector.FillRect(screen, x, y, w, h, fill, false)
		vector.StrokeRect(screen, x, y, w, h, 2, border, false)

		// NOTE: avoid "\n" in a single string here: the current text renderer
		// does not reliably support multi-line strings without artifacts.
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

		// Two separate text draws to keep line spacing predictable.
		op1 := &text.DrawOptions{}
		op1.GeoM.Translate(float64(x)+6, float64(y)+float64(uiLineH*0.95))
		op1.ColorScale.ScaleWithColor(textCol)
		text.Draw(screen, line1, hudFace, op1)

		if line2 != "" {
			op2 := &text.DrawOptions{}
			op2.GeoM.Translate(float64(x)+6, float64(y)+float64(uiLineH*1.90))
			op2.ColorScale.ScaleWithColor(textCol)
			text.Draw(screen, line2, hudFace, op2)
		}
	}

	for i := 0; i < 3; i++ {
		x := inner.X + float32(i)*((inner.W - uiGap*2) / 3)
		drawSlot(battlepkg.BattleRowFront, i, x, frontSlotsY)
		drawSlot(battlepkg.BattleRowBack, i, x, backSlotsY)
	}
}

func drawAbilityPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	drawPanelBox(screen, r, "ABILITIES", hudFace)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X)+float64(uiPad*0.6), float64(r.Y)+float64(uiLineH*2))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 150, G: 150, B: 150, A: 255})
		text.Draw(screen, "(waiting)", hudFace, op)
		return
	}

	abs := active.Abilities()
	sel := battle.PlayerTurn.SelectedAbilityID
	hoverIdx := battle.PlayerTurn.HoverAbilityIndex

	var hoveredAbility *battlepkg.Ability
	for i, id := range abs {
		if i >= len(layout.AbilityItemRects) {
			break
		}
		rowRect := toRect(layout.AbilityItemRects[i])
		a := battlepkg.GetAbility(id)
		prefix := " "
		col := color.RGBA{R: 220, G: 220, B: 220, A: 255}
		bg := color.RGBA{R: 35, G: 35, B: 35, A: 255}
		if id == sel && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			prefix = "▶"
			bg = color.RGBA{R: 60, G: 60, B: 30, A: 255}
			col = color.RGBA{R: 255, G: 235, B: 120, A: 255}
		} else if hoverIdx == i && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			// Hover highlight (mouse-driven).
			bg = color.RGBA{R: 40, G: 55, B: 70, A: 255}
			col = color.RGBA{R: 180, G: 220, B: 255, A: 255}
			hoveredAbility = &a
		}
		rule := ""
		switch a.TargetRule {
		case battlepkg.TargetEnemySingle:
			rule = "enemy"
		case battlepkg.TargetAllySingle:
			rule = "ally"
		case battlepkg.TargetSelf:
			rule = "self"
		default:
			rule = "none"
		}
		vector.FillRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, bg, false)

		line := fmt.Sprintf("%s %s [%s]", prefix, a.Name, rule)
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(rowRect.X+4), float64(rowRect.Y+uiLineH*0.9))
		op.ColorScale.ScaleWithColor(col)
		text.Draw(screen, line, hudFace, op)
		// y increment is encoded into layout.AbilityItemRects; nothing to update here.
	}

	// Very simple tooltip for hovered ability (inside the panel, under list).
	if hoveredAbility != nil {
		inner := inset(r, uiPad*0.6)
		infoY := inner.Y + inner.H - uiLineH*3
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(inner.X), float64(infoY))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 180, G: 220, B: 255, A: 255})

		target := ""
		switch hoveredAbility.TargetRule {
		case battlepkg.TargetEnemySingle:
			target = "enemy"
		case battlepkg.TargetAllySingle:
			target = "ally"
		case battlepkg.TargetSelf:
			target = "self"
		default:
			target = "none"
		}
		rng := ""
		switch hoveredAbility.Range {
		case battlepkg.RangeMelee:
			rng = "melee"
		case battlepkg.RangeRanged:
			rng = "ranged"
		default:
			rng = "—"
		}
		line := fmt.Sprintf("Target: %s | Range: %s", target, rng)
		text.Draw(screen, line, hudFace, op)
	}
}

func drawConfirmPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	drawPanelBox(screen, r, "ACTION", hudFace)
	if battle == nil {
		return
	}

	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X)+float64(uiPad*0.6), float64(r.Y)+float64(uiLineH*2))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 150, G: 150, B: 150, A: 255})
		text.Draw(screen, "(enemy turn)", hudFace, op)
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

	lines := []string{}

	// STEP / actor summary.
	switch pt.Phase {
	case battlepkg.PlayerChooseAbility:
		lines = append(lines, "STEP: Choose ability")
	case battlepkg.PlayerChooseTarget:
		lines = append(lines, "STEP: Choose target")
	case battlepkg.PlayerConfirmAction:
		lines = append(lines, "STEP: Confirm action")
	default:
		lines = append(lines, fmt.Sprintf("STEP: %s", pt.PhaseString()))
	}

	actor := battle.ActiveUnit()
	if actor != nil {
		roleStr := fmt.Sprintf("%v", actor.Def.Role)
		atkKind := "melee"
		if actor.IsRanged() {
			atkKind = "ranged"
		}
		lines = append(lines, fmt.Sprintf("Actor: %s (%s, %s)", actor.Name(), roleStr, atkKind))
		lines = append(lines, fmt.Sprintf("HP: %d/%d", actor.State.HP, actor.MaxHP()))
	}

	lines = append(lines, fmt.Sprintf("Ability: %s", a.Name))
	lines = append(lines, fmt.Sprintf("Target: %s", targetStr))

	// Preview (UI only reads PreviewAction API).
	req := pt.Pending
	if req.Actor == 0 {
		req = battlepkg.ActionRequest{Actor: active.ID, Ability: pt.SelectedAbilityID, Target: pt.SelectedTarget}
	}
	if prev, v := battlepkg.PreviewAction(battle, req); v.OK {
		if prev.HasDamage() {
			lines = append(lines, fmt.Sprintf("damage: ~%d-%d", prev.DamageMin, prev.DamageMax))
		} else if prev.HasHeal() {
			lines = append(lines, fmt.Sprintf("heal: ~%d-%d", prev.HealMin, prev.HealMax))
		}
	}

	if pt.Phase == battlepkg.PlayerConfirmAction {
		lines = append(lines, "Hint: Click CONFIRM or RMB to go back")
	} else if pt.Phase == battlepkg.PlayerChooseTarget {
		lines = append(lines, fmt.Sprintf("Hint: Click a highlighted enemy; %d valid targets", len(pt.ValidTargets)))
	} else if pt.Phase == battlepkg.PlayerChooseAbility {
		lines = append(lines, "Hint: Left-click ability, then target")
	}

	inner := inset(r, uiPad*0.6)
	y := inner.Y + uiLineH*1.6
	for _, line := range lines {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(inner.X), float64(y))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, line, hudFace, op)
		y += uiLineH * 1.1
	}

	// Unit hover info block.
	hoverID := battle.PlayerTurn.HoverTargetUnitID
	if hoverID != 0 {
		if hu := battle.Units[hoverID]; hu != nil {
			infoY := inner.Y + inner.H - uiLineH*3.0
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(inner.X), float64(infoY))
			op.ColorScale.ScaleWithColor(color.RGBA{R: 180, G: 220, B: 255, A: 255})
			roleStr := fmt.Sprintf("%v", hu.Def.Role)
			atkKind := "melee"
			if hu.IsRanged() {
				atkKind = "ranged"
			}
			text.Draw(screen, fmt.Sprintf("Hover: %s (%s, %s)", hu.Name(), roleStr, atkKind), hudFace, op)

			op2 := &text.DrawOptions{}
			op2.GeoM.Translate(float64(inner.X), float64(infoY+uiLineH))
			op2.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 240, B: 255, A: 255})
			text.Draw(screen, fmt.Sprintf("HP: %d/%d", hu.State.HP, hu.MaxHP()), hudFace, op2)
		}
	}

	// Confirm / Back buttons: use shared layout rects so visuals match hit-areas.
	backR := toRect(layout.BackButton)
	confirmR := toRect(layout.ConfirmButton)

	// Draw Back button (always available while on player turn).
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

		op := &text.DrawOptions{}
		// Roughly center text.
		textX := r.X + r.W*0.5 - float32(len(label))*uiLineH*0.22
		textY := r.Y + r.H*0.5 + uiLineH*0.15
		op.GeoM.Translate(float64(textX), float64(textY))
		op.ColorScale.ScaleWithColor(textCol)
		text.Draw(screen, label, hudFace, op)
	}

	canConfirm := pt.Phase == battlepkg.PlayerConfirmAction
	drawButton(backR, "Back", true, pt.HoverBackButton)
	drawButton(confirmR, "Confirm", canConfirm, pt.HoverConfirmButton && canConfirm)
}

func drawFooterPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect) {
	drawPanelBox(screen, r, "COMBAT LOG", hudFace)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	isPlayerTurn := active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction

	controls := "Esc: retreat"
	if isPlayerTurn {
		switch battle.PlayerTurn.Phase {
		case battlepkg.PlayerChooseAbility:
			controls = "Mouse: ability/target/confirm | RMB: back | Esc: retreat"
		case battlepkg.PlayerChooseTarget:
			controls = "Mouse: target/confirm | RMB: back | Esc: retreat"
		case battlepkg.PlayerConfirmAction:
			controls = "Mouse: confirm/back | RMB: back | Esc: retreat"
		default:
			controls = "Mouse: select | Keyboard: still works | Esc: retreat"
		}
	}

	inner := inset(r, uiPad*0.6)
	titleH := uiLineH * 2
	controlsH := uiLineH
	logTop := inner.Y + titleH
	controlsPadBottom := uiPad * 0.65
	logBottom := inner.Y + inner.H - controlsH - controlsPadBottom
	if logBottom < logTop {
		logBottom = logTop
	}
	availableLines := int((logBottom - logTop) / uiLineH)
	if availableLines < 1 {
		availableLines = 1
	}
	// Visually cap the log height: the footer should not feel "too tall" because of log space.
	availableLines = minInt(availableLines, 4)

	// Log lines (last N that fits).
	y := logTop
	maxLines := availableLines
	start := 0
	if len(battle.BattleLog) > maxLines {
		start = len(battle.BattleLog) - maxLines
	}
	for i := start; i < len(battle.BattleLog); i++ {
		line := strings.TrimSpace(battle.BattleLog[i])
		if len([]rune(line)) > 80 {
			rs := []rune(line)
			line = string(rs[:77]) + "..."
		}
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(inner.X), float64(y))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 220, G: 220, B: 220, A: 255})
		text.Draw(screen, line, hudFace, op)
		y += uiLineH
	}

	op2 := &text.DrawOptions{}
	// Raise baseline so glyph descenders aren't clipped by the panel border.
	op2.GeoM.Translate(float64(inner.X), float64(inner.Y+inner.H-controlsPadBottom))
	op2.ColorScale.ScaleWithColor(color.RGBA{R: 170, G: 170, B: 170, A: 255})
	text.Draw(screen, controls, hudFace, op2)
}
