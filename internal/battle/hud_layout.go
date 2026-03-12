package battle

// Battle HUD layout: single source of truth for both rendering and mouse hit-testing.
// All rectangles here are in screen coordinates.

type HUDRect struct {
	X, Y, W, H float32
}

// BattleHUDLayout describes all major areas of the battle HUD.
type BattleHUDLayout struct {
	// High-level containers.
	Overlay HUDRect // main battle panel within the darkened screen
	Content HUDRect // inner content area inside overlay (after title/banner)

	InfoLine HUDRect // one-line "Round / phase / active" text area

	Formation HUDRect // combined formation area (player+enemy)
	Middle    HUDRect // abilities + action panel row
	Footer    HUDRect // combat log + controls

	PlayerFormation HUDRect
	EnemyFormation  HUDRect

	Abilities HUDRect
	Action    HUDRect

	BackButton    HUDRect
	ConfirmButton HUDRect

	// Optional / more detailed areas.
	CombatLog HUDRect
	HintLine  HUDRect

	// Per-ability item rectangles (only meaningful on player turn).
	AbilityItemRects []HUDRect

	// Per-unit rectangles for formation slots that currently contain a unit.
	// Used for both rendering and hit-testing.
	UnitRects map[UnitID]HUDRect
}

// HUD layout constants, shared between render and mouse logic.
const (
	hudLineH = float32(18)
	hudPad   = float32(12)
	hudGap   = float32(10)
)

func hudClamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func hudInset(r HUDRect, pad float32) HUDRect {
	n := HUDRect{X: r.X + pad, Y: r.Y + pad, W: r.W - pad*2, H: r.H - pad*2}
	if n.W < 0 {
		n.W = 0
	}
	if n.H < 0 {
		n.H = 0
	}
	return n
}

// ComputeBattleHUDLayout builds a battle HUD layout for the given screen size and
// current battle state. It is the single source of truth for all major HUD rects.
func (b *BattleContext) ComputeBattleHUDLayout(screenW, screenH int) BattleHUDLayout {
	sw := float32(screenW)
	sh := float32(screenH)

	var layout BattleHUDLayout

	// 1) Overlay panel.
	marginX := hudClamp(sw*0.08, 12, 80)
	marginY := hudClamp(sh*0.08, 12, 80)
	panelW := sw - marginX*2
	panelH := sh - marginY*2
	panelW = hudClamp(panelW, 520, 760)
	panelH = hudClamp(panelH, 360, 540)

	panelX := (sw - panelW) / 2
	panelY := (sh - panelH) / 2
	layout.Overlay = HUDRect{X: panelX, Y: panelY, W: panelW, H: panelH}

	// 2) Inner content area, accounting for title and possible result banner.
	content := hudInset(layout.Overlay, hudPad)
	content.Y += hudLineH // title line
	content.H -= hudLineH
	if content.H < 0 {
		content.H = 0
	}

	extraHeaderLines := float32(0)
	if b != nil && b.Result != ResultNone {
		// When battle is finished, we reserve 2 extra lines for banner + hint.
		extraHeaderLines = 2
	}
	if extraHeaderLines > 0 {
		used := extraHeaderLines * hudLineH
		content.Y += used
		content.H -= used
		if content.H < 0 {
			content.H = 0
		}
	}
	layout.Content = content

	// 3) Info line at top of content.
	layout.InfoLine = HUDRect{
		X: content.X,
		Y: content.Y,
		W: content.W,
		H: hudLineH,
	}

	// 4) Vertical packing: info row + formation + middle + footer.
	afterInfo := HUDRect{
		X: content.X,
		Y: content.Y + hudLineH + hudGap,
		W: content.W,
		H: content.H - hudLineH - hudGap,
	}
	if afterInfo.H < 0 {
		afterInfo.H = 0
	}

	footerMin := hudLineH*4 + hudPad
	middleMin := hudLineH*5 + hudPad
	formationMin := hudLineH*7 + hudPad

	total := afterInfo.H
	footerH := hudClamp(total*0.22, footerMin, total)
	middleH := hudClamp(total*0.28, middleMin, total-footerH)
	formationH := total - footerH - middleH - hudGap*2
	if formationH < formationMin {
		deficit := formationMin - formationH
		take := hudClamp(deficit, 0, middleH-middleMin)
		middleH -= take
		deficit -= take
		if deficit > 0 {
			take2 := hudClamp(deficit, 0, footerH-footerMin)
			footerH -= take2
			deficit -= take2
		}
		formationH = total - footerH - middleH - hudGap*2
		if formationH < 0 {
			formationH = 0
		}
	}

	layout.Formation = HUDRect{
		X: afterInfo.X,
		Y: afterInfo.Y,
		W: afterInfo.W,
		H: hudClamp(formationH, 0, afterInfo.H),
	}
	layout.Middle = HUDRect{
		X: afterInfo.X,
		Y: layout.Formation.Y + layout.Formation.H + hudGap,
		W: afterInfo.W,
		H: hudClamp(middleH, 0, afterInfo.H),
	}
	layout.Footer = HUDRect{
		X: afterInfo.X,
		Y: layout.Middle.Y + layout.Middle.H + hudGap,
		W: afterInfo.W,
		H: afterInfo.Y + afterInfo.H - (layout.Middle.Y + layout.Middle.H + hudGap),
	}
	if layout.Footer.H < 0 {
		layout.Footer.H = 0
	}

	// 5) Split formation into player/enemy panels.
	colW := (layout.Formation.W - hudGap) / 2
	layout.PlayerFormation = HUDRect{
		X: layout.Formation.X,
		Y: layout.Formation.Y,
		W: colW,
		H: layout.Formation.H,
	}
	layout.EnemyFormation = HUDRect{
		X: layout.Formation.X + colW + hudGap,
		Y: layout.Formation.Y,
		W: layout.Formation.W - colW - hudGap,
		H: layout.Formation.H,
	}
	if layout.EnemyFormation.W < 0 {
		layout.EnemyFormation.W = 0
	}

	// 6) Middle split: abilities (left) + action (right).
	mColW := (layout.Middle.W - hudGap) / 2
	layout.Abilities = HUDRect{
		X: layout.Middle.X,
		Y: layout.Middle.Y,
		W: mColW,
		H: layout.Middle.H,
	}
	layout.Action = HUDRect{
		X: layout.Middle.X + mColW + hudGap,
		Y: layout.Middle.Y,
		W: layout.Middle.W - mColW - hudGap,
		H: layout.Middle.H,
	}
	if layout.Action.W < 0 {
		layout.Action.W = 0
	}

	// 7) Footer sub-areas: combat log + hint line.
	if layout.Footer.H > 0 {
		inner := hudInset(layout.Footer, hudPad*0.6)
		titleH := hudLineH * 2
		controlsH := hudLineH
		logTop := inner.Y + titleH
		controlsPadBottom := hudPad * 0.65
		logBottom := inner.Y + inner.H - controlsH - controlsPadBottom
		if logBottom < logTop {
			logBottom = logTop
		}
		layout.CombatLog = HUDRect{X: inner.X, Y: logTop, W: inner.W, H: logBottom - logTop}
		layout.HintLine = HUDRect{
			X: inner.X,
			Y: inner.Y + inner.H - controlsPadBottom - controlsH,
			W: inner.W,
			H: controlsH,
		}
	}

	// 8) Action panel: confirm/back buttons.
	if layout.Action.W > 0 && layout.Action.H > 0 {
		inner := hudInset(layout.Action, hudPad*0.6)
		btnH := hudLineH * 1.2
		btnW := inner.W * 0.45
		btnY := inner.Y + inner.H - btnH - hudPad*0.3
		layout.BackButton = HUDRect{X: inner.X, Y: btnY, W: btnW, H: btnH}
		layout.ConfirmButton = HUDRect{X: inner.X + inner.W - btnW, Y: btnY, W: btnW, H: btnH}
	}

	// 9) Ability item rects (only meaningful on player turn).
	layout.AbilityItemRects = nil
	active := b.ActiveUnit()
	if active != nil && active.Side == TeamPlayer && b.Phase == PhaseAwaitAction {
		abs := active.Abilities()
		if len(abs) > 0 {
			inner := hudInset(layout.Abilities, hudPad*0.6)
			y := inner.Y + hudLineH*2
			maxY := inner.Y + inner.H - hudLineH*2.5
			rects := make([]HUDRect, 0, len(abs))
			for range abs {
				if y > maxY {
					break
				}
				row := HUDRect{
					X: inner.X,
					Y: y - hudLineH*0.2,
					W: inner.W,
					H: hudLineH * 1.4,
				}
				rects = append(rects, row)
				y += hudLineH * 1.3
			}
			layout.AbilityItemRects = rects
		}
	}

	// 10) Unit rects in formation (only living units).
	layout.UnitRects = map[UnitID]HUDRect{}
	if b != nil {
		// Helper: compute slot rects inside a formation panel.
		computeUnitRectsForSide := func(panel HUDRect, side BattleSide) {
			inner := hudInset(panel, hudPad*0.6)
			inner.Y += hudLineH
			inner.H -= hudLineH
			if inner.H < 0 {
				inner.H = 0
			}
			cellW := (inner.W - hudGap*2) / 3
			rowGap := hudGap * 0.6
			labelH := hudLineH
			rowAreaH := (inner.H - labelH*2 - rowGap) / 2
			cellH := hudClamp(rowAreaH, hudLineH*2.4, hudLineH*3.5)

			frontLabelY := inner.Y
			frontSlotsY := frontLabelY + labelH
			backLabelY := frontSlotsY + cellH + rowGap
			backSlotsY := backLabelY + labelH

			for rowIdx, row := range []BattleRow{BattleRowFront, BattleRowBack} {
				slotY := frontSlotsY
				if rowIdx == 1 {
					slotY = backSlotsY
				}
				for i := 0; i < 3; i++ {
					slot := b.Slot(side, row, i)
					if slot == nil || slot.Occupied == 0 {
						continue
					}
					u := b.Units[slot.Occupied]
					if u == nil || !u.IsAlive() {
						continue
					}
					x := inner.X + float32(i)*cellW
					r := HUDRect{X: x, Y: slotY, W: cellW - 4, H: cellH}
					layout.UnitRects[u.ID] = r
				}
			}
		}

		computeUnitRectsForSide(layout.PlayerFormation, BattleSidePlayer)
		computeUnitRectsForSide(layout.EnemyFormation, BattleSideEnemy)
	}

	return layout
}

