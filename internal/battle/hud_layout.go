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

	// Top info area (two lines for readability).
	TopInfoPrimary   HUDRect // Round / Battle phase
	TopInfoSecondary HUDRect // Active unit / side / player turn subphase
	InfoLine         HUDRect // legacy: combined top info area (both lines)

	Formation HUDRect // combined formation area (player+enemy)
	Middle    HUDRect // abilities + action panel row
	Footer    HUDRect // combat log + controls

	PlayerFormation HUDRect
	EnemyFormation  HUDRect

	Abilities HUDRect
	Action    HUDRect

	// Ability panel sub-areas.
	AbilityHeader  HUDRect // area under the panel title and before the list
	AbilityList    HUDRect // scroll-free list area for abilities
	AbilityTooltip HUDRect // compact tooltip/info area at the bottom of the panel

	// Action panel sub-areas.
	ActionMain    HUDRect // step/ability/target/preview summary
	ActorInfo     HUDRect // compact current actor info
	HoverInfo     HUDRect // compact hovered/target unit info
	ActionButtons HUDRect // row that contains Back/Confirm buttons

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

	// 3) Top info area: two lines (primary / secondary).
	layout.TopInfoPrimary = HUDRect{
		X: content.X,
		Y: content.Y,
		W: content.W,
		H: hudLineH,
	}
	layout.TopInfoSecondary = HUDRect{
		X: content.X,
		Y: content.Y + hudLineH,
		W: content.W,
		H: hudLineH,
	}
	// Legacy combined rect for compatibility with any existing users.
	layout.InfoLine = HUDRect{
		X: content.X,
		Y: content.Y,
		W: content.W,
		H: hudLineH * 2,
	}

	// 4) Vertical packing: top info (2 lines) + formation + middle + footer.
	afterInfo := HUDRect{
		X: content.X,
		Y: content.Y + hudLineH*2 + hudGap,
		W: content.W,
		H: content.H - hudLineH*2 - hudGap,
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

	// 8) Ability panel sub-areas: header, list, tooltip.
	if layout.Abilities.W > 0 && layout.Abilities.H > 0 {
		inner := hudInset(layout.Abilities, hudPad*0.6)
		headerH := hudLineH * 1.6
		tooltipH := hudLineH * 2.4

		availableH := inner.H - headerH - tooltipH - hudGap*0.4
		if availableH < hudLineH*2 {
			availableH = hudLineH * 2
			// Allow tooltip to shrink a bit on very small panels.
			maxTooltip := inner.H - headerH - availableH - hudGap*0.4
			if maxTooltip < tooltipH && maxTooltip > hudLineH*1.4 {
				tooltipH = maxTooltip
			}
		}
		if availableH < 0 {
			availableH = 0
		}

		listTop := inner.Y + headerH
		listH := availableH
		tooltipY := listTop + listH + hudGap*0.2

		layout.AbilityHeader = HUDRect{X: inner.X, Y: inner.Y, W: inner.W, H: headerH}
		layout.AbilityList = HUDRect{X: inner.X, Y: listTop, W: inner.W, H: listH}
		layout.AbilityTooltip = HUDRect{X: inner.X, Y: tooltipY, W: inner.W, H: tooltipH}
	}

	// 9) Action panel: main content, compact info blocks, confirm/back buttons.
	if layout.Action.W > 0 && layout.Action.H > 0 {
		inner := hudInset(layout.Action, hudPad*0.6)

		// Buttons row at the bottom with a bit more presence.
		btnH := hudLineH * 1.4
		buttonsGap := hudPad * 0.4
		buttonsY := inner.Y + inner.H - btnH

		layout.ActionButtons = HUDRect{
			X: inner.X,
			Y: buttonsY,
			W: inner.W,
			H: btnH,
		}

		btnW := (inner.W - hudGap) / 2
		if btnW < inner.W*0.3 {
			btnW = inner.W * 0.3
		}
		layout.BackButton = HUDRect{X: inner.X, Y: buttonsY, W: btnW, H: btnH}
		layout.ConfirmButton = HUDRect{X: inner.X + inner.W - btnW, Y: buttonsY, W: btnW, H: btnH}

		// Above buttons: clean action summary + compact actor/hover info.
		topAreaBottom := buttonsY - buttonsGap
		topAreaH := topAreaBottom - inner.Y
		if topAreaH < hudLineH*3 {
			topAreaH = hudLineH * 3
		}

		actorH := hudLineH * 2.4
		hoverH := hudLineH * 2.4
		summaryH := topAreaH - actorH - hoverH - hudGap*0.4
		if summaryH < hudLineH*2 {
			summaryH = hudLineH * 2
			// Allow info blocks to shrink on small panels.
			actorH = hudLineH * 1.8
			hoverH = hudLineH * 1.8
		}

		layout.ActionMain = HUDRect{X: inner.X, Y: inner.Y, W: inner.W, H: summaryH}
		actorY := inner.Y + summaryH + hudGap*0.2
		layout.ActorInfo = HUDRect{X: inner.X, Y: actorY, W: inner.W, H: actorH}
		hoverY := actorY + actorH + hudGap*0.2
		layout.HoverInfo = HUDRect{X: inner.X, Y: hoverY, W: inner.W, H: hoverH}
	}

	// 10) Ability item rects (only meaningful on player turn).
	layout.AbilityItemRects = nil
	active := b.ActiveUnit()
	if active != nil && active.Side == TeamPlayer && b.Phase == PhaseAwaitAction {
		abs := active.Abilities()
		if len(abs) > 0 {
			list := layout.AbilityList
			if list.W <= 0 || list.H <= 0 {
				// Fallback: simple inset when sub-areas are not available.
				list = hudInset(layout.Abilities, hudPad*0.6)
				list.Y += hudLineH * 2
				list.H -= hudLineH * 2
				if list.H < 0 {
					list.H = 0
				}
			}

			y := list.Y + hudLineH*0.1
			maxY := list.Y + list.H - hudLineH*1.3
			rects := make([]HUDRect, 0, len(abs))
			for range abs {
				if y > maxY {
					break
				}
				row := HUDRect{
					X: list.X,
					Y: y - hudLineH*0.2,
					W: list.W,
					H: hudLineH * 1.4,
				}
				rects = append(rects, row)
				y += hudLineH * 1.25
			}
			layout.AbilityItemRects = rects
		}
	}

	// 11) Unit rects in formation (only living units).
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

