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
	metrics := layout.Metrics
	panelX := panel.X
	panelY := panel.Y

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(panelX)+float64(metrics.Pad), float64(panelY)+float64(metrics.Pad)+float64(metrics.LineH))
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
		bannerY := float64(panelY) + float64(metrics.Pad) + float64(metrics.LineH)*2
		bannerOp := &text.DrawOptions{}
		bannerOp.GeoM.Translate(float64(panelX)+float64(metrics.Pad), bannerY)
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
		hintOp.GeoM.Translate(float64(panelX)+float64(metrics.Pad), bannerY+float64(metrics.LineH))
		hintOp.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
		text.Draw(screen, "SPACE/ENTER: continue", hudFace, hintOp)
	}

	// Top info: strict rects from layout.
	primary := toRect(layout.TopInfoPrimary)
	secondary := toRect(layout.TopInfoSecondary)

	compactTop := isCompactForRect(metrics, primary)
	roundStr := fmt.Sprintf("Round %d", battle.Round)
	phaseStr := fmt.Sprintf("Phase: %s", battle.PhaseString())
	if compactTop {
		roundStr = fmt.Sprintf("R %d", battle.Round)
		phaseStr = fmt.Sprintf("Ph: %s", battle.PhaseString())
	}
	line1 := fmt.Sprintf("%s | %s", roundStr, phaseStr)
	drawSingleLineInRect(screen, hudFace, primary, line1, metrics, color.White)

	active := battle.ActiveUnit()
	activeLabel := "-"
	sideLabel := ""
	if active != nil {
		activeLabel = fmt.Sprintf("%s (#%d)", active.Name(), active.ID)
		if active.Side == battlepkg.TeamPlayer {
			if compactTop {
				sideLabel = "P"
			} else {
				sideLabel = "PLAYER"
			}
		} else {
			if compactTop {
				sideLabel = "E"
			} else {
				sideLabel = "ENEMY"
			}
		}
	}
	playerSub := ""
	if active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction {
		if compactTop {
			playerSub = fmt.Sprintf("Pl: %s", battle.PlayerTurn.PhaseString())
		} else {
			playerSub = fmt.Sprintf("Player: %s", battle.PlayerTurn.PhaseString())
		}
	}
	line2 := fmt.Sprintf("Active: %s", activeLabel)
	if sideLabel != "" {
		line2 = fmt.Sprintf("%s [%s]", line2, sideLabel)
	}
	if playerSub != "" {
		line2 = fmt.Sprintf("%s | %s", line2, playerSub)
	}
	drawSingleLineInRect(screen, hudFace, secondary, line2, metrics, color.RGBA{R: 220, G: 220, B: 220, A: 255})

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

	drawFooterPanel(screen, hudFace, battle, footerRect, layout)
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

// measureTextWidth returns the rendered width of a string for the given face.
func measureTextWidth(face *text.GoTextFace, s string) float32 {
	if s == "" || face == nil {
		return 0
	}
	adv := text.Advance(s, face)
	return float32(adv)
}

// trimTextToWidth returns a single-line string that fits into maxW pixels,
// appending "..." when trimming is required.
func trimTextToWidth(face *text.GoTextFace, s string, maxW float32) string {
	if maxW <= 0 || s == "" || face == nil {
		return ""
	}
	if measureTextWidth(face, s) <= maxW {
		return s
	}

	const ellipsis = "..."
	ellW := measureTextWidth(face, ellipsis)
	if ellW >= maxW {
		return ""
	}

	rs := []rune(s)
	lo, hi := 0, len(rs)
	best := ""
	for lo <= hi {
		mid := (lo + hi) / 2
		cand := string(rs[:mid])
		if measureTextWidth(face, cand)+ellW <= maxW {
			best = cand
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	if best == "" {
		return ellipsis
	}
	return best + ellipsis
}

// fitTextToWidth is a convenience alias for single-line trimming.
func fitTextToWidth(face *text.GoTextFace, s string, maxW float32) string {
	return trimTextToWidth(face, s, maxW)
}

// baselineYForLineInRect returns the Y position (top of line) for a single line of text
// so that a line of height metrics.LineH is vertically centered in the rect.
// Used by drawSingleLineInRect; no manual baseline math elsewhere.
func baselineYForLineInRect(r rect, metrics battlepkg.HUDMetrics) float32 {
	if metrics.LineH <= 0 || r.H <= 0 {
		return r.Y
	}
	off := (r.H - metrics.LineH) * 0.5
	if off < 0 {
		off = 0
	}
	return r.Y + off
}

// drawSingleLineInRect draws one line of text inside r, using the shared baseline helper.
// Text is trimmed to fit r.W and vertically placed within r. Does not draw if r has no area.
func drawSingleLineInRect(screen *ebiten.Image, face *text.GoTextFace, r rect, line string, metrics battlepkg.HUDMetrics, col color.Color) {
	if r.W <= 0 || r.H <= 0 || face == nil {
		return
	}
	line = fitTextToWidth(face, line, r.W)
	if line == "" {
		return
	}
	y := baselineYForLineInRect(r, metrics)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(r.X), float64(y))
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, line, face, op)
}

// isCompactForRect decides whether a given rect should use compact wording
// based on its width and the current HUD metrics.
func isCompactForRect(metrics battlepkg.HUDMetrics, r rect) bool {
	// Treat narrow panels or small line heights as candidates for compact text.
	if r.W <= 0 {
		return false
	}
	if r.W < 260 {
		return true
	}
	if metrics.LineH <= 16 {
		return true
	}
	return false
}

// maxLinesForRect returns how many lines (at the given lineStep) can fit
// vertically into the rect, respecting top and bottom padding.
func maxLinesForRect(metrics battlepkg.HUDMetrics, r rect, topPad, bottomPad, lineStep float32) int {
	usableH := r.H - topPad - bottomPad
	if usableH <= 0 || lineStep <= 0 {
		return 0
	}
	n := int(usableH / lineStep)
	if n < 0 {
		return 0
	}
	return n
}

// drawLinesInRect draws lines inside r on a strict grid: lineStep = metrics.LineH,
// first line at r.Y. Does not draw outside r. maxLines caps count; 0 = use capacity.
func drawLinesInRect(screen *ebiten.Image, face *text.GoTextFace, r rect, lines []string, metrics battlepkg.HUDMetrics, col color.Color, maxLines int) int {
	if len(lines) == 0 || r.W <= 0 || r.H <= 0 || face == nil {
		return 0
	}
	lineStep := metrics.LineH
	capacity := int(r.H / lineStep)
	if capacity <= 0 {
		return 0
	}
	if maxLines > 0 && maxLines < capacity {
		capacity = maxLines
	}
	linesToDraw := capacity
	if linesToDraw > len(lines) {
		linesToDraw = len(lines)
	}

	drawn := 0
	for i := 0; i < linesToDraw; i++ {
		y := r.Y + float32(i)*lineStep
		if y+lineStep > r.Y+r.H {
			break
		}
		line := lines[i]
		row := rect{X: r.X, Y: y, W: r.W, H: lineStep}
		drawSingleLineInRect(screen, face, row, line, metrics, col)
		drawn++
	}
	return drawn
}

// drawPanelBox draws panel background and border; title only in titleRow (strict text grid).
func drawPanelBox(screen *ebiten.Image, panelRect rect, titleRow rect, title string, hudFace *text.GoTextFace, metrics battlepkg.HUDMetrics) {
	bg := color.RGBA{R: 28, G: 28, B: 28, A: 255}
	border := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	vector.FillRect(screen, panelRect.X, panelRect.Y, panelRect.W, panelRect.H, bg, false)
	vector.StrokeRect(screen, panelRect.X, panelRect.Y, panelRect.W, panelRect.H, 1, border, false)

	if title != "" && titleRow.W > 0 && titleRow.H > 0 {
		drawSingleLineInRect(screen, hudFace, titleRow, title, metrics, color.RGBA{R: 200, G: 200, B: 200, A: 255})
	}
}

func drawFormationPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, side battlepkg.BattleSide, title string, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	titleRow := toRect(layout.PlayerFormationTitleRow)
	if side == battlepkg.BattleSideEnemy {
		titleRow = toRect(layout.EnemyFormationTitleRow)
	}
	drawPanelBox(screen, r, titleRow, title, hudFace, metrics)
	if battle == nil {
		return
	}
	inner := inset(r, metrics.Pad*0.6)
	inner.Y += metrics.LineH
	inner.H -= metrics.LineH
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
			w = (inner.W - metrics.Gap*2) / 3
			h = clampF((inner.H-labelH*2-metrics.Gap*0.6)/2, metrics.LineH*2.4, metrics.LineH*3.5)
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
	drawPanelBox(screen, r, toRect(layout.AbilitiesTitleRow), "ABILITIES", hudFace, metrics)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		titleRow := toRect(layout.AbilitiesTitleRow)
		if titleRow.W > 0 && titleRow.H > 0 {
			drawSingleLineInRect(screen, hudFace, titleRow, "(waiting)", metrics, color.RGBA{R: 150, G: 150, B: 150, A: 255})
		}
		return
	}

	abs := active.Abilities()
	sel := battle.PlayerTurn.SelectedAbilityID
	hoverIdx := battle.PlayerTurn.HoverAbilityIndex

	for i, id := range abs {
		if i >= len(layout.AbilityItemRects) {
			break
		}
		rowRect := toRect(layout.AbilityItemRects[i])
		a := battlepkg.GetAbility(id)
		prefix := " "
		col := color.RGBA{R: 220, G: 220, B: 220, A: 255}
		bg := color.RGBA{R: 35, G: 35, B: 35, A: 255}
		border := color.RGBA{R: 70, G: 70, B: 70, A: 255}
		if id == sel && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			prefix = "▶"
			bg = color.RGBA{R: 60, G: 60, B: 30, A: 255}
			col = color.RGBA{R: 255, G: 235, B: 120, A: 255}
			border = color.RGBA{R: 200, G: 200, B: 120, A: 255}
		} else if hoverIdx == i && battle.PlayerTurn.Phase == battlepkg.PlayerChooseAbility {
			// Hover highlight (mouse-driven).
			bg = color.RGBA{R: 40, G: 55, B: 70, A: 255}
			col = color.RGBA{R: 180, G: 220, B: 255, A: 255}
			border = color.RGBA{R: 140, G: 190, B: 255, A: 255}
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
		vector.StrokeRect(screen, rowRect.X, rowRect.Y, rowRect.W, rowRect.H, 1, border, false)

		compactRow := isCompactForRect(metrics, rowRect)
		line := ""
		if compactRow {
			line = fmt.Sprintf("%s %s", prefix, a.Name)
		} else {
			line = fmt.Sprintf("%s %s [%s]", prefix, a.Name, rule)
		}
		line = fitTextToWidth(hudFace, line, rowRect.W-8)
		textRow := rect{X: rowRect.X + 4, Y: rowRect.Y, W: rowRect.W - 8, H: rowRect.H}
		drawSingleLineInRect(screen, hudFace, textRow, line, metrics, col)
		// y increment is encoded into layout.AbilityItemRects; nothing to update here.
	}

	// Simplified HUD: no ability tooltip (removed for stability).
}

func drawConfirmPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, toRect(layout.ActionTitleRow), "ACTION", hudFace, metrics)
	if battle == nil {
		return
	}

	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		summaryRect := toRect(layout.ActionSummary)
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

	// STEP / main summary in ActionSummary content area.
	summaryLines := []string{}
	summaryRect := toRect(layout.ActionSummary)
	summaryCompact := isCompactForRect(metrics, summaryRect)
	switch pt.Phase {
	case battlepkg.PlayerChooseAbility:
		if summaryCompact {
			summaryLines = append(summaryLines, "Step: ability")
		} else {
			summaryLines = append(summaryLines, "STEP: Choose ability")
		}
	case battlepkg.PlayerChooseTarget:
		if summaryCompact {
			summaryLines = append(summaryLines, "Step: target")
		} else {
			summaryLines = append(summaryLines, "STEP: Choose target")
		}
	case battlepkg.PlayerConfirmAction:
		if summaryCompact {
			summaryLines = append(summaryLines, "Step: confirm")
		} else {
			summaryLines = append(summaryLines, "STEP: Confirm action")
		}
	default:
		if summaryCompact {
			summaryLines = append(summaryLines, fmt.Sprintf("Step: %s", pt.PhaseString()))
		} else {
			summaryLines = append(summaryLines, fmt.Sprintf("STEP: %s", pt.PhaseString()))
		}
	}

	// Ability / target.
	abilityLine := ""
	if summaryCompact {
		abilityLine = fmt.Sprintf("Abil: %s", a.Name)
	} else {
		abilityLine = fmt.Sprintf("Ability: %s", a.Name)
	}
	summaryLines = append(summaryLines, abilityLine)
	summaryLines = append(summaryLines, fmt.Sprintf("Target: %s", targetStr))

	// Draw summary lines strictly inside ActionSummary rect.
	maxSummaryW := summaryRect.W
	for i := range summaryLines {
		summaryLines[i] = fitTextToWidth(hudFace, summaryLines[i], maxSummaryW)
	}
	_ = drawLinesInRect(screen, hudFace, summaryRect, summaryLines, metrics, color.White, 0)

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

		drawSingleLineInRect(screen, hudFace, r, label, metrics, textCol)
	}

	canConfirm := pt.Phase == battlepkg.PlayerConfirmAction
	drawButton(backR, "Back", true, pt.HoverBackButton)
	drawButton(confirmR, "Confirm", canConfirm, pt.HoverConfirmButton && canConfirm)
}

func drawFooterPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, toRect(layout.FooterTitleRow), "COMBAT LOG", hudFace, metrics)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	isPlayerTurn := active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction

	logRect := toRect(layout.CombatLog)
	controlsRect := toRect(layout.HintLine)

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
		compactControls := isCompactForRect(metrics, controlsRect)
		controls := "Esc: retreat"
		if isPlayerTurn {
			switch battle.PlayerTurn.Phase {
			case battlepkg.PlayerChooseAbility:
				if compactControls {
					controls = "LMB ability/target | RMB back | Esc retreat"
				} else {
					controls = "Mouse: ability/target/confirm | RMB: back | Esc: retreat"
				}
			case battlepkg.PlayerChooseTarget:
				if compactControls {
					controls = "LMB target | RMB back | Esc retreat"
				} else {
					controls = "Mouse: target/confirm | RMB: back | Esc: retreat"
				}
			case battlepkg.PlayerConfirmAction:
				if compactControls {
					controls = "LMB confirm | RMB back | Esc retreat"
				} else {
					controls = "Mouse: confirm/back | RMB: back | Esc: retreat"
				}
			default:
				if compactControls {
					controls = "LMB select | Esc retreat"
				} else {
					controls = "Mouse: select | Keyboard: still works | Esc: retreat"
				}
			}
		}
		drawSingleLineInRect(screen, hudFace, controlsRect, controls, metrics, color.RGBA{R: 170, G: 170, B: 170, A: 255})
	}
}
