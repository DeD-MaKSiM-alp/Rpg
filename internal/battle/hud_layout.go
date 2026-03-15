package battle

// Battle HUD layout: single source of truth for both rendering and mouse hit-testing.
// All rectangles here are in screen coordinates.

type HUDRect struct {
	X, Y, W, H float32
}

// HUDMetrics describes adaptive HUD metrics derived from screen size.
type HUDMetrics struct {
	LineH    float32
	Pad      float32
	Gap      float32
	SmallGap float32
	ButtonH  float32
	TitleH   float32
}

// BattleHUDLayout describes all major areas of the battle HUD.
type BattleHUDLayout struct {
	// Adaptive metrics used to build all rects below.
	Metrics HUDMetrics

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

	// Ability panel: strict text grid.
	AbilitiesTitleRow HUDRect // one line for panel title
	AbilityList       HUDRect // content area (abilities list)
	AbilityHeader     HUDRect // deprecated: same as AbilitiesTitleRow
	AbilityTooltip    HUDRect // unused in simplified HUD

	// Action panel: strict text grid.
	ActionTitleRow   HUDRect // one line for panel title
	ActionSummary    HUDRect // content: step / ability / target
	ActionButtons    HUDRect // row with Back/Confirm
	ActionMain       HUDRect // deprecated: same as ActionSummary
	ActorInfo        HUDRect // unused
	HoverInfo        HUDRect // unused

	BackButton    HUDRect
	ConfirmButton HUDRect

	// Footer: strict text grid.
	FooterTitleRow HUDRect // one line for "COMBAT LOG"
	CombatLog      HUDRect // content area (log lines)
	HintLine       HUDRect // one line for controls

	// Formation panels: title row per side (one line each).
	PlayerFormationTitleRow HUDRect
	EnemyFormationTitleRow  HUDRect

	// Per-ability item rectangles (only meaningful on player turn).
	AbilityItemRects []HUDRect

	// Per-unit rectangles for formation slots that currently contain a unit.
	// Used for both rendering and hit-testing.
	UnitRects map[UnitID]HUDRect
}

// hudClamp clamps v into [lo, hi].

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

// computeHUDMetrics derives adaptive HUD metrics from the current screen size.
// It uses the smaller screen dimension to compute a scale factor and then
// clamps each metric into a sensible range so that the HUD remains legible
// on both small and large resolutions.
func computeHUDMetrics(screenW, screenH int) HUDMetrics {
	sw := float32(screenW)
	sh := float32(screenH)
	minDim := sw
	if sh < minDim {
		minDim = sh
	}

	// Baseline around a 720p-ish resolution.
	base := minDim / 720.0
	if base < 0.75 {
		base = 0.75
	}
	if base > 1.4 {
		base = 1.4
	}

	var m HUDMetrics

	// Line height: around 18 at baseline, clamped into [14..24].
	m.LineH = hudClamp(18*base, 14, 24)
	// General padding: around 12 at baseline, [8..18].
	m.Pad = hudClamp(12*base, 8, 18)
	// Standard gap between blocks: around 10 at baseline, [6..14].
	m.Gap = hudClamp(10*base, 6, 14)
	// Smaller gap for tighter vertical groupings.
	m.SmallGap = hudClamp(6*base, 4, 10)
	// Button height: around 26 at baseline, [20..36].
	m.ButtonH = hudClamp(26*base, 20, 36)
	// Title line height (overlay title / banner).
	m.TitleH = hudClamp(20*base, 16, 28)

	return m
}

// ComputeBattleHUDLayout builds a battle HUD layout for the given screen size and
// current battle state. It is the single source of truth for all major HUD rects.
func (b *BattleContext) ComputeBattleHUDLayout(screenW, screenH int) BattleHUDLayout {
	sw := float32(screenW)
	sh := float32(screenH)

	metrics := computeHUDMetrics(screenW, screenH)

	var layout BattleHUDLayout
	layout.Metrics = metrics

	// 1) Overlay panel.
	// Use a percentage of screen size, with gentle safety clamps, so that the
	// panel grows on large resolutions and shrinks on small ones, but never
	// touches the screen edges.
	targetPanelW := sw * 0.86
	targetPanelH := sh * 0.82
	minPanelW := hudClamp(sw*0.55, 440, sw-2*metrics.Pad-8)
	minPanelH := hudClamp(sh*0.50, 300, sh-2*metrics.Pad-8)
	maxPanelW := sw - 2*metrics.Pad
	maxPanelH := sh - 2*metrics.Pad

	panelW := hudClamp(targetPanelW, minPanelW, maxPanelW)
	panelH := hudClamp(targetPanelH, minPanelH, maxPanelH)

	panelX := (sw - panelW) / 2
	panelY := (sh - panelH) / 2
	layout.Overlay = HUDRect{X: panelX, Y: panelY, W: panelW, H: panelH}

	// 2) Inner content area, accounting for title and possible result banner.
	content := hudInset(layout.Overlay, metrics.Pad)
	content.Y += metrics.TitleH // title line
	content.H -= metrics.TitleH
	if content.H < 0 {
		content.H = 0
	}

	extraHeaderLines := float32(0)
	if b != nil && b.Result != ResultNone {
		// When battle is finished, we reserve 2 extra lines for banner + hint.
		extraHeaderLines = 2
	}
	if extraHeaderLines > 0 {
		used := extraHeaderLines * metrics.LineH
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
		H: metrics.LineH,
	}
	layout.TopInfoSecondary = HUDRect{
		X: content.X,
		Y: content.Y + metrics.LineH,
		W: content.W,
		H: metrics.LineH,
	}
	// Legacy combined rect for compatibility with any existing users.
	layout.InfoLine = HUDRect{
		X: content.X,
		Y: content.Y,
		W: content.W,
		H: metrics.LineH * 2,
	}

	// 4) Vertical packing: top info (2 lines) + formation + middle + footer.
	afterInfo := HUDRect{
		X: content.X,
		Y: content.Y + metrics.LineH*2 + metrics.Gap,
		W: content.W,
		H: content.H - metrics.LineH*2 - metrics.Gap,
	}
	if afterInfo.H < 0 {
		afterInfo.H = 0
	}

	// Minimal heights for key regions, driven by line height.
	footerMin := metrics.LineH*3 + metrics.Pad
	middleMin := metrics.LineH*4 + metrics.Pad
	formationMin := metrics.LineH*6 + metrics.Pad

	total := afterInfo.H
	footerH := hudClamp(total*0.22, footerMin, total)
	middleH := hudClamp(total*0.28, middleMin, total-footerH)
	formationH := total - footerH - middleH - metrics.Gap*2
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
		formationH = total - footerH - middleH - metrics.Gap*2
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
		Y: layout.Formation.Y + layout.Formation.H + metrics.Gap,
		W: afterInfo.W,
		H: hudClamp(middleH, 0, afterInfo.H),
	}
	layout.Footer = HUDRect{
		X: afterInfo.X,
		Y: layout.Middle.Y + layout.Middle.H + metrics.Gap,
		W: afterInfo.W,
		H: afterInfo.Y + afterInfo.H - (layout.Middle.Y + layout.Middle.H + metrics.Gap),
	}
	if layout.Footer.H < 0 {
		layout.Footer.H = 0
	}

	// 5) Split formation into player/enemy panels.
	colW := (layout.Formation.W - metrics.Gap) / 2
	layout.PlayerFormation = HUDRect{
		X: layout.Formation.X,
		Y: layout.Formation.Y,
		W: colW,
		H: layout.Formation.H,
	}
	layout.EnemyFormation = HUDRect{
		X: layout.Formation.X + colW + metrics.Gap,
		Y: layout.Formation.Y,
		W: layout.Formation.W - colW - metrics.Gap,
		H: layout.Formation.H,
	}
	if layout.EnemyFormation.W < 0 {
		layout.EnemyFormation.W = 0
	}
	titleRowH := metrics.LineH
	layout.PlayerFormationTitleRow = HUDRect{X: layout.PlayerFormation.X, Y: layout.PlayerFormation.Y, W: layout.PlayerFormation.W, H: titleRowH}
	layout.EnemyFormationTitleRow = HUDRect{X: layout.EnemyFormation.X, Y: layout.EnemyFormation.Y, W: layout.EnemyFormation.W, H: titleRowH}

	// 6) Middle split: abilities (left) + action (right).
	mColW := (layout.Middle.W - metrics.Gap) / 2
	layout.Abilities = HUDRect{
		X: layout.Middle.X,
		Y: layout.Middle.Y,
		W: mColW,
		H: layout.Middle.H,
	}
	layout.Action = HUDRect{
		X: layout.Middle.X + mColW + metrics.Gap,
		Y: layout.Middle.Y,
		W: layout.Middle.W - mColW - metrics.Gap,
		H: layout.Middle.H,
	}
	if layout.Action.W < 0 {
		layout.Action.W = 0
	}

	// 7) Footer: FooterTitleRow (one line) + CombatLog + HintLine (one line).
	if layout.Footer.H > 0 {
		inner := hudInset(layout.Footer, metrics.Pad*0.6)
		titleH := metrics.LineH
		controlsH := metrics.LineH
		logTop := inner.Y + titleH + metrics.SmallGap
		controlsPadBottom := metrics.Pad * 0.5
		logBottom := inner.Y + inner.H - controlsH - controlsPadBottom
		if logBottom < logTop {
			logBottom = logTop
		}
		layout.FooterTitleRow = HUDRect{X: inner.X, Y: inner.Y, W: inner.W, H: titleH}
		layout.CombatLog = HUDRect{X: inner.X, Y: logTop, W: inner.W, H: logBottom - logTop}
		layout.HintLine = HUDRect{
			X: inner.X,
			Y: inner.Y + inner.H - controlsPadBottom - controlsH,
			W: inner.W,
			H: controlsH,
		}
	}

	// 8) Ability panel: AbilitiesTitleRow (one line) + AbilityList.
	if layout.Abilities.W > 0 && layout.Abilities.H > 0 {
		inner := hudInset(layout.Abilities, metrics.Pad*0.6)
		titleH := metrics.LineH
		listTop := inner.Y + titleH + metrics.SmallGap*0.5
		listH := inner.Y + inner.H - listTop
		if listH < metrics.LineH*2 {
			listH = metrics.LineH * 2
		}
		if listH < 0 {
			listH = 0
		}
		layout.AbilitiesTitleRow = HUDRect{X: inner.X, Y: inner.Y, W: inner.W, H: titleH}
		layout.AbilityHeader = layout.AbilitiesTitleRow
		layout.AbilityList = HUDRect{X: inner.X, Y: listTop, W: inner.W, H: listH}
		layout.AbilityTooltip = HUDRect{}
	}

	// 9) Action panel: ActionTitleRow (one line) + ActionSummary + ActionButtons.
	if layout.Action.W > 0 && layout.Action.H > 0 {
		inner := hudInset(layout.Action, metrics.Pad*0.6)

		btnH := metrics.ButtonH
		buttonsGap := metrics.SmallGap
		buttonsY := inner.Y + inner.H - btnH
		layout.ActionButtons = HUDRect{X: inner.X, Y: buttonsY, W: inner.W, H: btnH}

		btnW := (inner.W - metrics.Gap) / 2
		if btnW < inner.W*0.3 {
			btnW = inner.W * 0.3
		}
		layout.BackButton = HUDRect{X: inner.X, Y: buttonsY, W: btnW, H: btnH}
		layout.ConfirmButton = HUDRect{X: inner.X + inner.W - btnW, Y: buttonsY, W: btnW, H: btnH}

		titleH := metrics.LineH
		summaryTop := inner.Y + titleH + metrics.SmallGap*0.5
		summaryBottom := buttonsY - buttonsGap
		summaryH := summaryBottom - summaryTop
		if summaryH < metrics.LineH*2 {
			summaryH = metrics.LineH * 2
		}
		if summaryH < 0 {
			summaryH = 0
		}
		layout.ActionTitleRow = HUDRect{X: inner.X, Y: inner.Y, W: inner.W, H: titleH}
		layout.ActionSummary = HUDRect{X: inner.X, Y: summaryTop, W: inner.W, H: summaryH}
		layout.ActionMain = layout.ActionSummary
		layout.ActorInfo = HUDRect{}
		layout.HoverInfo = HUDRect{}
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
				list = hudInset(layout.Abilities, metrics.Pad*0.6)
				list.Y += metrics.LineH * 2
				list.H -= metrics.LineH * 2
				if list.H < 0 {
					list.H = 0
				}
			}

			y := list.Y + metrics.LineH*0.1
			maxY := list.Y + list.H - metrics.LineH*1.3
			rects := make([]HUDRect, 0, len(abs))
			for range abs {
				if y > maxY {
					break
				}
				row := HUDRect{
					X: list.X,
					Y: y - metrics.LineH*0.2,
					W: list.W,
					H: metrics.LineH * 1.4,
				}
				rects = append(rects, row)
				y += metrics.LineH * 1.25
			}
			layout.AbilityItemRects = rects
		}
	}

	// 11) Unit rects in formation (only living units).
	layout.UnitRects = map[UnitID]HUDRect{}
	if b != nil {
		// Helper: compute slot rects inside a formation panel.
		computeUnitRectsForSide := func(panel HUDRect, side BattleSide) {
			inner := hudInset(panel, metrics.Pad*0.6)
			inner.Y += metrics.LineH
			inner.H -= metrics.LineH
			if inner.H < 0 {
				inner.H = 0
			}
			cellW := (inner.W - metrics.Gap*2) / 3
			rowGap := metrics.Gap * 0.6
			labelH := metrics.LineH
			rowAreaH := (inner.H - labelH*2 - rowGap) / 2
			cellH := hudClamp(rowAreaH, metrics.LineH*2.4, metrics.LineH*3.5)

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

