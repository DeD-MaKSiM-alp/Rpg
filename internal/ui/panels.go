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

// drawBattleOverlayPanel рисует затемнённый фон и центральную панель боевого overlay.
func drawBattleOverlayPanel(screen *ebiten.Image, screenWidth, screenHeight int) {
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	panelColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	panelBorderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)

	panelX := float32(80)
	panelY := float32(80)
	panelW := float32(640)
	panelH := float32(360)

	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelColor, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2, panelBorderColor, false)
}

// drawBattleOverlayText рисует все текстовые блоки боевого overlay.
func drawBattleOverlayText(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext) {
	panelX := float32(80)
	panelY := float32(80)
	panelW := float32(640)
	panelH := float32(360)

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(panelX)+20, float64(panelY)+35)
	titleOp.ColorScale.ScaleWithColor(color.White)

	title := "Battle mode"
	if battle != nil && len(battle.Encounter.Enemies) > 0 {
		title = fmt.Sprintf("Battle mode: enemy #%d", battle.Encounter.Enemies[0].EnemyID)
	}

	text.Draw(screen, title, hudFace, titleOp)

	if battle == nil {
		return
	}

	offsetY := 0.0
	if battle.Result != battlepkg.ResultNone {
		bannerY := float64(panelY) + 70
		bannerOp := &text.DrawOptions{}
		bannerOp.GeoM.Translate(float64(panelX)+20, bannerY)
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
		hintOp.GeoM.Translate(float64(panelX)+20, bannerY+25)
		hintOp.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
		text.Draw(screen, "SPACE/ENTER: continue", hudFace, hintOp)
		offsetY = 55
	}

	// Layout: top info line + two formation panels + ability panel + message/controls.
	contentTop := float32(panelY) + 60 + float32(offsetY)
	contentLeft := panelX + 20
	contentRight := panelX + panelW - 20

	// Info line.
	infoOp := &text.DrawOptions{}
	infoOp.GeoM.Translate(float64(contentLeft), float64(contentTop))
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

	// Formation panels.
	formationTop := contentTop + 18
	formationH := float32(160)
	gap := float32(12)
	colW := (contentRight - contentLeft - gap) / 2

	playerPanel := rect{X: contentLeft, Y: formationTop, W: colW, H: formationH}
	enemyPanel := rect{X: contentLeft + colW + gap, Y: formationTop, W: colW, H: formationH}

	drawFormationPanel(screen, hudFace, battle, playerPanel, battlepkg.BattleSidePlayer, "PLAYER")
	drawFormationPanel(screen, hudFace, battle, enemyPanel, battlepkg.BattleSideEnemy, "ENEMY")

	// Ability panel (only meaningful on player turn).
	abilitiesTop := formationTop + formationH + 10
	abilitiesH := float32(110)
	abilitiesRect := rect{X: contentLeft, Y: abilitiesTop, W: colW, H: abilitiesH}
	confirmRect := rect{X: contentLeft + colW + gap, Y: abilitiesTop, W: colW, H: abilitiesH}

	drawAbilityPanel(screen, hudFace, battle, abilitiesRect)
	drawConfirmPanel(screen, hudFace, battle, confirmRect)

	// Combat log + controls.
	footerTop := abilitiesTop + abilitiesH + 10
	footerRect := rect{X: contentLeft, Y: footerTop, W: contentRight - contentLeft, H: (panelY + panelH - 20) - footerTop}
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

func drawPanelBox(screen *ebiten.Image, r rect, title string, hudFace *text.GoTextFace) {
	bg := color.RGBA{R: 28, G: 28, B: 28, A: 255}
	border := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, bg, false)
	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, border, false)

	if title == "" {
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(r.X)+8, float64(r.Y)+14)
	op.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	text.Draw(screen, title, hudFace, op)
}

func drawFormationPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, side battlepkg.BattleSide, title string) {
	drawPanelBox(screen, r, title, hudFace)
	if battle == nil {
		return
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

	// Slot grid: 3 front + 3 back.
	cellW := (r.W - 16) / 3
	cellH := float32(44)
	rowGap := float32(10)

	drawRowLabel := func(label string, y float32) {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X)+8, float64(y)+14)
		op.ColorScale.ScaleWithColor(color.RGBA{R: 170, G: 170, B: 170, A: 255})
		text.Draw(screen, label, hudFace, op)
	}

	frontY := r.Y + 22
	backY := frontY + cellH + rowGap + 16

	drawRowLabel("FRONT", frontY)
	drawRowLabel("BACK", backY)

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

		// Valid/selected/active highlights (read from battle state; no rule logic here).
		if u != nil && validSet[u.ID] {
			border = color.RGBA{R: 80, G: 150, B: 255, A: 255}
		}
		if u != nil && u.ID == selectedTargetID {
			border = color.RGBA{R: 255, G: 80, B: 80, A: 255}
		}
		if active != nil && u != nil && u.ID == active.ID {
			border = color.RGBA{R: 255, G: 215, B: 80, A: 255}
		}

		vector.FillRect(screen, x, y, cellW-4, cellH, fill, false)
		vector.StrokeRect(screen, x, y, cellW-4, cellH, 2, border, false)

		label := fmt.Sprintf("(%d)", idx)
		if u != nil {
			hp := fmt.Sprintf("%d/%d", u.State.HP, u.MaxHP())
			name := u.Name()
			if len([]rune(name)) > 10 {
				rs := []rune(name)
				name = string(rs[:10])
			}
			label = fmt.Sprintf("%s\nHP %s", name, hp)
			if !u.IsAlive() {
				label = fmt.Sprintf("%s\nDEAD", name)
			}
		} else {
			label = "EMPTY"
		}

		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(x)+6, float64(y)+14)
		op.ColorScale.ScaleWithColor(textCol)
		text.Draw(screen, label, hudFace, op)
	}

	for i := 0; i < 3; i++ {
		x := r.X + 8 + float32(i)*cellW
		drawSlot(battlepkg.BattleRowFront, i, x, frontY+12)
		drawSlot(battlepkg.BattleRowBack, i, x, backY+12)
	}
}

func drawAbilityPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect) {
	drawPanelBox(screen, r, "ABILITIES", hudFace)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X)+8, float64(r.Y)+34)
		op.ColorScale.ScaleWithColor(color.RGBA{R: 150, G: 150, B: 150, A: 255})
		text.Draw(screen, "(waiting)", hudFace, op)
		return
	}

	abs := active.Abilities()
	sel := battle.PlayerTurn.SelectedAbilityID

	y := r.Y + 34
	for i, id := range abs {
		a := battlepkg.GetAbility(id)
		prefix := "  "
		col := color.RGBA{R: 220, G: 220, B: 220, A: 255}
		if id == sel && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			prefix = "> "
			col = color.RGBA{R: 255, G: 215, B: 80, A: 255}
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
		line := fmt.Sprintf("%s%d) %s [%s]", prefix, i+1, a.Name, rule)
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X)+8, float64(y))
		op.ColorScale.ScaleWithColor(col)
		text.Draw(screen, line, hudFace, op)
		y += 16
		if y > r.Y+r.H-10 {
			break
		}
	}
}

func drawConfirmPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect) {
	drawPanelBox(screen, r, "ACTION", hudFace)
	if battle == nil {
		return
	}

	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X)+8, float64(r.Y)+34)
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

	lines := []string{
		fmt.Sprintf("phase: %s", pt.PhaseString()),
		fmt.Sprintf("action: %s → %s", a.Name, targetStr),
	}

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
		lines = append(lines, "READY: confirm")
	} else if pt.Phase == battlepkg.PlayerChooseTarget {
		lines = append(lines, fmt.Sprintf("valid targets: %d", len(pt.ValidTargets)))
	}

	y := r.Y + 34
	for _, line := range lines {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X)+8, float64(y))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, line, hudFace, op)
		y += 16
	}
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
			controls = "Arrows: ability | Space/Enter: choose | Backspace: (noop) | Esc: retreat"
		case battlepkg.PlayerChooseTarget:
			controls = "Arrows: target | Space/Enter: choose | Backspace: back | Esc: retreat"
		case battlepkg.PlayerConfirmAction:
			controls = "Space/Enter: confirm | Backspace: back | Esc: retreat"
		default:
			controls = "Arrows: select | Space/Enter: confirm | Backspace: back | Esc: retreat"
		}
	}

	// Log lines (last N).
	y := r.Y + 34
	maxLines := 6
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
		op.GeoM.Translate(float64(r.X)+8, float64(y))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 220, G: 220, B: 220, A: 255})
		text.Draw(screen, line, hudFace, op)
		y += 16
	}

	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(float64(r.X)+8, float64(r.Y+r.H-10))
	op2.ColorScale.ScaleWithColor(color.RGBA{R: 170, G: 170, B: 170, A: 255})
	text.Draw(screen, controls, hudFace, op2)
}
