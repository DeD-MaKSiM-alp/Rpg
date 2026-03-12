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

	// Layout: container-based inside overlay, driven by shared layout.
	primary := toRect(layout.TopInfoPrimary)
	secondary := toRect(layout.TopInfoSecondary)

	// Info line 1: Round / Battle phase.
	infoOp1 := &text.DrawOptions{}
	infoOp1.GeoM.Translate(float64(primary.X)+float64(metrics.Pad*0.6), singleLineBaselineY(primary, metrics))
	infoOp1.ColorScale.ScaleWithColor(color.White)

	compactTop := isCompactForRect(metrics, primary)
	roundStr := fmt.Sprintf("Round %d", battle.Round)
	phaseStr := fmt.Sprintf("Phase: %s", battle.PhaseString())
	if compactTop {
		roundStr = fmt.Sprintf("R %d", battle.Round)
		phaseStr = fmt.Sprintf("Ph: %s", battle.PhaseString())
	}
	line1 := fmt.Sprintf("%s | %s", roundStr, phaseStr)
	line1 = fitTextToWidth(hudFace, line1, primary.W-metrics.Pad*1.2)
	text.Draw(screen, line1, hudFace, infoOp1)

	// Info line 2: Active unit / side / player turn subphase.
	infoOp2 := &text.DrawOptions{}
	infoOp2.GeoM.Translate(float64(secondary.X)+float64(metrics.Pad*0.6), singleLineBaselineY(secondary, metrics))
	infoOp2.ColorScale.ScaleWithColor(color.RGBA{R: 220, G: 220, B: 220, A: 255})

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
	line2 = fitTextToWidth(hudFace, line2, secondary.W-metrics.Pad*1.2)
	text.Draw(screen, line2, hudFace, infoOp2)

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

// singleLineBaselineY computes a safe baseline Y for a single-line label
// inside the given rect, using HUD metrics. It keeps the text comfortably
// within the rect without touching the bottom border.
func singleLineBaselineY(r rect, metrics battlepkg.HUDMetrics) float64 {
	h := metrics.LineH
	if h <= 0 || r.H <= 0 {
		return float64(r.Y + r.H*0.8)
	}
	// Start from a slightly above-centered baseline.
	base := r.Y + (r.H+h)*0.5 - h*0.15
	minY := r.Y + h*0.7
	maxY := r.Y + r.H - h*0.25
	if base < minY {
		base = minY
	}
	if base > maxY {
		base = maxY
	}
	return float64(base)
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

// drawLinesInRect draws up to maxLines (or as many as fit) from the given list
// of lines inside the rect, using a fixed vertical step. Returns the number of
// lines actually drawn.
func drawLinesInRect(screen *ebiten.Image, face *text.GoTextFace, r rect, lines []string, metrics battlepkg.HUDMetrics, col color.Color, maxLines int) int {
	if len(lines) == 0 || r.W <= 0 || r.H <= 0 || face == nil {
		return 0
	}
	lineStep := metrics.LineH * 1.05
	// Slightly smaller top padding so that compact blocks (like ActionMain /
	// ActorInfo / HoverInfo) can use more of their vertical space.
	topPad := metrics.LineH * 0.5
	bottomPad := float32(0)
	capacity := maxLinesForRect(metrics, r, topPad, bottomPad, lineStep)
	if maxLines > 0 && maxLines < capacity {
		capacity = maxLines
	}
	if capacity <= 0 {
		return 0
	}

	linesToDraw := capacity
	if linesToDraw > len(lines) {
		linesToDraw = len(lines)
	}

	y := r.Y + topPad
	drawn := 0
	for i := 0; i < linesToDraw; i++ {
		line := lines[i]
		if line == "" {
			y += lineStep
			continue
		}
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(r.X), float64(y))
		op.ColorScale.ScaleWithColor(col)
		text.Draw(screen, line, face, op)
		drawn++
		y += lineStep
	}
	return drawn
}

func drawPanelBox(screen *ebiten.Image, r rect, title string, hudFace *text.GoTextFace, metrics battlepkg.HUDMetrics) {
	bg := color.RGBA{R: 28, G: 28, B: 28, A: 255}
	border := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, bg, false)
	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, border, false)

	if title == "" {
		return
	}
	op := &text.DrawOptions{}
	maxW := r.W - metrics.Pad*1.2
	titleText := fitTextToWidth(hudFace, title, maxW)
	op.GeoM.Translate(float64(r.X)+float64(metrics.Pad*0.6), singleLineBaselineY(r, metrics))
	op.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	text.Draw(screen, titleText, hudFace, op)
}

func drawFormationPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, side battlepkg.BattleSide, title string, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, title, hudFace, metrics)
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

	// Slot grid labels (approximate; per-unit rects come from shared layout).
	labelH := metrics.LineH
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

		// Two separate text draws to keep line spacing predictable.
		op1 := &text.DrawOptions{}
		op1.GeoM.Translate(float64(x)+6, float64(y)+float64(metrics.LineH*0.95))
		op1.ColorScale.ScaleWithColor(textCol)
		text.Draw(screen, line1, hudFace, op1)

		if line2 != "" {
			op2 := &text.DrawOptions{}
			op2.GeoM.Translate(float64(x)+6, float64(y)+float64(metrics.LineH*1.90))
			op2.ColorScale.ScaleWithColor(textCol)
			text.Draw(screen, line2, hudFace, op2)
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
	drawPanelBox(screen, r, "ABILITIES", hudFace, metrics)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		// Draw a compact "(waiting)" marker inside the ability header area.
		headerRect := toRect(layout.AbilityHeader)
		if headerRect.W <= 0 || headerRect.H <= 0 {
			headerRect = inset(r, metrics.Pad*0.6)
		}
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(headerRect.X), singleLineBaselineY(headerRect, metrics))
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
			hoveredAbility = &a
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
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(rowRect.X+4), float64(rowRect.Y+metrics.LineH*0.9))
		op.ColorScale.ScaleWithColor(col)
		text.Draw(screen, line, hudFace, op)
		// y increment is encoded into layout.AbilityItemRects; nothing to update here.
	}

	// Very simple tooltip for hovered ability (inside the panel, under list).
	if hoveredAbility != nil {
		tipRect := toRect(layout.AbilityTooltip)
		if tipRect.W <= 0 || tipRect.H <= 0 {
			tipRect = inset(r, metrics.Pad*0.6)
		}
		infoY := tipRect.Y + metrics.LineH*0.9
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(tipRect.X), float64(infoY))
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
		compactTip := isCompactForRect(metrics, tipRect)
		line := ""
		if compactTip {
			line = fmt.Sprintf("T: %s | R: %s", target, rng)
		} else {
			line = fmt.Sprintf("Target: %s | Range: %s", target, rng)
		}
		line = fitTextToWidth(hudFace, line, tipRect.W)
		text.Draw(screen, line, hudFace, op)
	}
}

func drawConfirmPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, "ACTION", hudFace, metrics)
	if battle == nil {
		return
	}

	active := battle.ActiveUnit()
	if active == nil || active.Side != battlepkg.TeamPlayer || battle.Phase != battlepkg.PhaseAwaitAction {
		// Enemy turn: use the main action summary rect for a compact status label.
		mainRect := toRect(layout.ActionMain)
		if mainRect.W <= 0 || mainRect.H <= 0 {
			mainRect = inset(r, metrics.Pad*0.6)
		}
		label := "(enemy turn)"
		label = fitTextToWidth(hudFace, label, mainRect.W-metrics.Pad*0.4)
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(mainRect.X), singleLineBaselineY(mainRect, metrics))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 150, G: 150, B: 150, A: 255})
		text.Draw(screen, label, hudFace, op)
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

	// STEP / main summary.
	summaryLines := []string{}
	summaryRect := toRect(layout.ActionMain)
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

	// Preview (UI only reads PreviewAction API).
	req := pt.Pending
	if req.Actor == 0 {
		req = battlepkg.ActionRequest{Actor: active.ID, Ability: pt.SelectedAbilityID, Target: pt.SelectedTarget}
	}
	if prev, v := battlepkg.PreviewAction(battle, req); v.OK {
		if prev.HasDamage() {
			if summaryCompact {
				summaryLines = append(summaryLines, fmt.Sprintf("Dmg: %d-%d", prev.DamageMin, prev.DamageMax))
			} else {
				summaryLines = append(summaryLines, fmt.Sprintf("Preview: dmg ~%d-%d", prev.DamageMin, prev.DamageMax))
			}
		} else if prev.HasHeal() {
			if summaryCompact {
				summaryLines = append(summaryLines, fmt.Sprintf("Heal: %d-%d", prev.HealMin, prev.HealMax))
			} else {
				summaryLines = append(summaryLines, fmt.Sprintf("Preview: heal ~%d-%d", prev.HealMin, prev.HealMax))
			}
		}
	}

	if pt.Phase == battlepkg.PlayerConfirmAction {
		if summaryCompact {
			summaryLines = append(summaryLines, "Enter confirm | RMB back")
		} else {
			summaryLines = append(summaryLines, "Hint: Confirm or RMB to go back")
		}
	} else if pt.Phase == battlepkg.PlayerChooseTarget {
		if summaryCompact {
			summaryLines = append(summaryLines, fmt.Sprintf("Click target (%d)", len(pt.ValidTargets)))
		} else {
			summaryLines = append(summaryLines, fmt.Sprintf("Hint: Click highlighted target (%d options)", len(pt.ValidTargets)))
		}
	} else if pt.Phase == battlepkg.PlayerChooseAbility {
		if summaryCompact {
			summaryLines = append(summaryLines, "LMB: ability -> target")
		} else {
			summaryLines = append(summaryLines, "Hint: Left-click ability, then choose target")
		}
	}

	// Summary block: STEP / Ability / Target / Preview / Hint.
	summaryInner := inset(summaryRect, metrics.Pad*0.3)
	maxSummaryW := summaryInner.W
	for i := range summaryLines {
		summaryLines[i] = fitTextToWidth(hudFace, summaryLines[i], maxSummaryW)
	}
	// Respect vertical capacity: highest priority lines идут первыми, поэтому
	// при дефиците высоты будут отброшены Hint/Preview.
	_ = drawLinesInRect(screen, hudFace, summaryInner, summaryLines, metrics, color.White, 0)

	// Compact current actor info block.
	actorRect := toRect(layout.ActorInfo)
	actor := battle.ActiveUnit()
	if actor != nil && actorRect.W > 0 && actorRect.H > 0 {
		infoInner := inset(actorRect, metrics.Pad*0.3)
		roleStr := fmt.Sprintf("%v", actor.Def.Role)
		atkKind := "melee"
		if actor.IsRanged() {
			atkKind = "ranged"
		}
		nameLine := fmt.Sprintf("Actor: %s (%s, %s)", actor.Name(), roleStr, atkKind)
		hpLine := fmt.Sprintf("HP %d/%d", actor.State.HP, actor.MaxHP())
		nameLine = fitTextToWidth(hudFace, nameLine, infoInner.W)
		hpLine = fitTextToWidth(hudFace, hpLine, infoInner.W)

		lines := []string{nameLine, hpLine}
		// Prefer dropping HP line before name when there is not enough height.
		if maxLinesForRect(metrics, infoInner, metrics.LineH*0.8, 0, metrics.LineH*1.05) < 2 {
			lines = lines[:1]
		}
		_ = drawLinesInRect(screen, hudFace, infoInner, lines, metrics, color.RGBA{R: 220, G: 220, B: 220, A: 255}, 0)
	}

	// Compact hovered/target unit info block.
	hoverRect := toRect(layout.HoverInfo)
	if hoverRect.W > 0 && hoverRect.H > 0 {
		infoInner := inset(hoverRect, metrics.Pad*0.3)
		hoverID := battle.PlayerTurn.HoverTargetUnitID

		var hu *battlepkg.BattleUnit
		if hoverID != 0 {
			hu = battle.Units[hoverID]
		} else if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit {
			hu = battle.Units[pt.SelectedTarget.UnitID]
		}

		if hu != nil {
			roleStr := fmt.Sprintf("%v", hu.Def.Role)
			atkKind := "melee"
			if hu.IsRanged() {
				atkKind = "ranged"
			}

			label := "Target"
			if hoverID != 0 {
				label = "Hover"
			}

			nameLine := fmt.Sprintf("%s: %s (%s, %s)", label, hu.Name(), roleStr, atkKind)
			hpLine := fmt.Sprintf("HP %d/%d", hu.State.HP, hu.MaxHP())
			nameLine = fitTextToWidth(hudFace, nameLine, infoInner.W)
			hpLine = fitTextToWidth(hudFace, hpLine, infoInner.W)

			lines := []string{nameLine, hpLine}
			if maxLinesForRect(metrics, infoInner, metrics.LineH*0.8, 0, metrics.LineH*1.05) < 2 {
				lines = lines[:1]
			}
			_ = drawLinesInRect(screen, hudFace, infoInner, lines, metrics, color.RGBA{R: 180, G: 220, B: 255, A: 255}, 0)
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
		// Approximate centering based on adaptive line height and label length.
		textWApprox := float32(len(label)) * metrics.LineH * 0.4
		textX := r.X + (r.W-textWApprox)/2
		textY := r.Y + r.H*0.5 + metrics.LineH*0.15
		op.GeoM.Translate(float64(textX), float64(textY))
		op.ColorScale.ScaleWithColor(textCol)
		text.Draw(screen, label, hudFace, op)
	}

	canConfirm := pt.Phase == battlepkg.PlayerConfirmAction
	drawButton(backR, "Back", true, pt.HoverBackButton)
	drawButton(confirmR, "Confirm", canConfirm, pt.HoverConfirmButton && canConfirm)
}

func drawFooterPanel(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, r rect, layout battlepkg.BattleHUDLayout) {
	metrics := layout.Metrics
	drawPanelBox(screen, r, "COMBAT LOG", hudFace, metrics)
	if battle == nil {
		return
	}
	active := battle.ActiveUnit()
	isPlayerTurn := active != nil && active.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction

	// Combat log area and controls area come from shared layout.
	logRect := toRect(layout.CombatLog)
	controlsRect := toRect(layout.HintLine)

	if logRect.W <= 0 || logRect.H <= 0 {
		// Fallback to an inset region if combat log rect is unavailable.
		logRect = inset(r, metrics.Pad*0.6)
	}
	availableLines := int(logRect.H / metrics.LineH)
	if availableLines < 1 {
		availableLines = 1
	}
	// Visually cap the log height: the footer should not feel "too tall" because of log space.
	availableLines = minInt(availableLines, 4)

	// Log lines (last N that fits).
	y := logRect.Y + metrics.LineH*0.9
	maxLines := availableLines
	start := 0
	if len(battle.BattleLog) > maxLines {
		start = len(battle.BattleLog) - maxLines
	}
	for i := start; i < len(battle.BattleLog); i++ {
		line := strings.TrimSpace(battle.BattleLog[i])
		line = fitTextToWidth(hudFace, line, logRect.W)
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(logRect.X), float64(y))
		op.ColorScale.ScaleWithColor(color.RGBA{R: 220, G: 220, B: 220, A: 255})
		text.Draw(screen, line, hudFace, op)
		y += metrics.LineH
	}

	op2 := &text.DrawOptions{}
	// Place controls/hint text inside the reserved controls rect with
	// normal/compact wording based on available width.
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
	controlsText := fitTextToWidth(hudFace, controls, controlsRect.W)
	op2.GeoM.Translate(float64(controlsRect.X), singleLineBaselineY(controlsRect, metrics))
	op2.ColorScale.ScaleWithColor(color.RGBA{R: 170, G: 170, B: 170, A: 255})
	text.Draw(screen, controlsText, hudFace, op2)
}
