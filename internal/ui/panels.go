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
func drawBattleOverlayPanel(screen *ebiten.Image, screenWidth, screenHeight int) rect {
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	panelColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	panelBorderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)

	sw := float32(screenWidth)
	sh := float32(screenHeight)

	// Overlay takes most of the screen with margins, centered.
	marginX := clampF(sw*0.08, 12, 80)
	marginY := clampF(sh*0.08, 12, 80)
	panelW := sw - marginX*2
	panelH := sh - marginY*2

	// Reasonable max size (keeps things readable if screen is large later).
	panelW = clampF(panelW, 520, 760)
	panelH = clampF(panelH, 360, 540)

	panelX := (sw - panelW) / 2
	panelY := (sh - panelH) / 2

	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelColor, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, uiPanelBorder, panelBorderColor, false)
	return rect{X: panelX, Y: panelY, W: panelW, H: panelH}
}

// drawBattleOverlayText рисует все текстовые блоки боевого overlay.
func drawBattleOverlayText(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, overlay rect) {
	panelX := overlay.X
	panelY := overlay.Y
	panelW := overlay.W
	panelH := overlay.H

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

	extraHeaderLines := float32(0)
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
		extraHeaderLines = 2
	}

	// Layout: container-based inside overlay.
	content := inset(rect{X: panelX, Y: panelY, W: panelW, H: panelH}, uiPad)
	content.Y += uiLineH // title line already used
	content.H -= uiLineH

	// Reserve extra header rows when result banner is shown.
	if extraHeaderLines > 0 {
		used := extraHeaderLines * uiLineH
		content.Y += used
		content.H -= used
	}
	if content.H < 0 {
		content.H = 0
	}

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

	// Vertical packing: info row + formation + middle + footer.
	afterInfo := rect{X: content.X, Y: content.Y + uiLineH + uiGap, W: content.W, H: content.H - uiLineH - uiGap}
	if afterInfo.H < 0 {
		afterInfo.H = 0
	}

	footerMin := uiLineH*4 + uiPad
	middleMin := uiLineH*5 + uiPad
	formationMin := uiLineH*7 + uiPad

	total := afterInfo.H
	// Footer is intentionally kept smaller; it should not dominate the HUD.
	footerH := clampF(total*0.22, footerMin, total)
	middleH := clampF(total*0.28, middleMin, total-footerH)
	formationH := total - footerH - middleH - uiGap*2
	if formationH < formationMin {
		deficit := formationMin - formationH
		take := clampF(deficit, 0, middleH-middleMin)
		middleH -= take
		deficit -= take
		if deficit > 0 {
			take2 := clampF(deficit, 0, footerH-footerMin)
			footerH -= take2
			deficit -= take2
		}
		formationH = total - footerH - middleH - uiGap*2
		if formationH < 0 {
			formationH = 0
		}
	}

	formationRect := rect{X: afterInfo.X, Y: afterInfo.Y, W: afterInfo.W, H: clampF(formationH, 0, afterInfo.H)}
	middleRect := rect{X: afterInfo.X, Y: formationRect.Y + formationRect.H + uiGap, W: afterInfo.W, H: clampF(middleH, 0, afterInfo.H)}
	footerRect := rect{X: afterInfo.X, Y: middleRect.Y + middleRect.H + uiGap, W: afterInfo.W, H: afterInfo.Y + afterInfo.H - (middleRect.Y + middleRect.H + uiGap)}
	if footerRect.H < 0 {
		footerRect.H = 0
	}

	colW := (formationRect.W - uiGap) / 2
	leftCol, rightCol := splitH(formationRect, colW, uiGap)

	playerPanel := leftCol
	enemyPanel := rightCol

	drawFormationPanel(screen, hudFace, battle, playerPanel, battlepkg.BattleSidePlayer, "PLAYER")
	drawFormationPanel(screen, hudFace, battle, enemyPanel, battlepkg.BattleSideEnemy, "ENEMY")

	// Ability panel (only meaningful on player turn).
	mColW := (middleRect.W - uiGap) / 2
	mLeft, mRight := splitH(middleRect, mColW, uiGap)
	abilitiesRect := mLeft
	confirmRect := mRight

	drawAbilityPanel(screen, hudFace, battle, abilitiesRect)
	drawConfirmPanel(screen, hudFace, battle, confirmRect)

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
	op.GeoM.Translate(float64(r.X)+float64(uiPad*0.6), float64(r.Y)+float64(uiLineH))
	op.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	text.Draw(screen, title, hudFace, op)
}

func drawFormationPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, side battlepkg.BattleSide, title string) {
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

	// Slot grid: 3 front + 3 back.
	cellW := (inner.W - uiGap*2) / 3
	rowGap := uiGap * 0.6
	labelH := uiLineH
	rowAreaH := (inner.H - labelH*2 - rowGap) / 2
	cellH := clampF(rowAreaH, uiLineH*2.4, uiLineH*3.5)

	drawRowLabel := func(label string, y float32) {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(inner.X), float64(y+labelH))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 170, G: 170, B: 170, A: 255})
		text.Draw(screen, label, hudFace, op)
	}

	frontLabelY := inner.Y
	frontSlotsY := frontLabelY + labelH
	backLabelY := frontSlotsY + cellH + rowGap
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

		w := cellW - 4
		vector.FillRect(screen, x, y, w, cellH, fill, false)
		vector.StrokeRect(screen, x, y, w, cellH, 2, border, false)

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
		x := inner.X + float32(i)*cellW
		drawSlot(battlepkg.BattleRowFront, i, x, frontSlotsY)
		drawSlot(battlepkg.BattleRowBack, i, x, backSlotsY)
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
		op.GeoM.Translate(float64(r.X)+float64(uiPad*0.6), float64(r.Y)+float64(uiLineH*2))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 150, G: 150, B: 150, A: 255})
		text.Draw(screen, "(waiting)", hudFace, op)
		return
	}

	abs := active.Abilities()
	sel := battle.PlayerTurn.SelectedAbilityID

	inner := inset(r, uiPad*0.6)
	y := inner.Y + uiLineH*2
	maxY := inner.Y + inner.H - uiLineH*0.5
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
		op.GeoM.Translate(float64(inner.X), float64(y))
		op.ColorScale.ScaleWithColor(col)
		text.Draw(screen, line, hudFace, op)
		y += uiLineH
		if y > maxY {
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

	inner := inset(r, uiPad*0.6)
	y := inner.Y + uiLineH*2
	for _, line := range lines {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(inner.X), float64(y))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, line, hudFace, op)
		y += uiLineH
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
