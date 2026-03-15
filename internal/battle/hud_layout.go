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
// Style: 0 = v1 table, 1 = v2 Disciples-like. When Style==1, v2 rects are filled and used for render + mouse.
type BattleHUDLayout struct {
	Metrics HUDMetrics
	Style   int // 0 = v1, 1 = v2

	// --- v2 layout (Disciples-like: center battlefield, side rosters, bottom panel) ---
	TopBar      HUDRect // minimal top: round, turn, phase
	LeftRoster  HUDRect // player unit column
	RightRoster HUDRect // enemy unit column
	Battlefield HUDRect // center scene (main visual)
	BottomPanel HUDRect // command/info strip at bottom

	V2TopBarLine   HUDRect // single line inside TopBar
	V2BottomActive HUDRect // active unit summary in BottomPanel
	V2BottomTarget HUDRect // target summary in BottomPanel
	V2BottomSummary HUDRect // ability/target/preview line in BottomPanel
	V2BottomLog    HUDRect // one-line log or hint in BottomPanel

	// --- v1 core layout (used by battle HUD v1 render when Style==0) ---
	Overlay  HUDRect // main panel
	Content  HUDRect // inner area after TitleRow and optional result banner
	TitleRow HUDRect // overlay title, 1 line (LineH)
	InfoRow1 HUDRect // Round | Phase
	InfoRow2 HUDRect // Active unit

	Formation HUDRect
	Middle    HUDRect
	Footer    HUDRect

	PlayerFormation HUDRect
	EnemyFormation  HUDRect

	Abilities HUDRect
	Action    HUDRect

	AbilitiesTitleRow HUDRect
	AbilityList       HUDRect

	ActionTitleRow   HUDRect
	ActionSummary    HUDRect
	ActionButtons    HUDRect

	FooterTitleRow HUDRect
	CombatLog      HUDRect
	HintLine       HUDRect

	PlayerFormationTitleRow HUDRect
	EnemyFormationTitleRow  HUDRect

	// --- shared interactive rects (render + mouse) ---
	BackButton    HUDRect
	ConfirmButton HUDRect
	AbilityItemRects []HUDRect
	UnitRects     map[UnitID]HUDRect

	// --- compatibility / legacy (not used by v1 render; set in one place at end of Compute) ---
	TopInfoPrimary   HUDRect // alias InfoRow1
	TopInfoSecondary HUDRect // alias InfoRow2
	InfoLine         HUDRect // InfoRow1+2 combined
	AbilityHeader    HUDRect // alias AbilitiesTitleRow
	AbilityTooltip   HUDRect // unused
	ActionMain       HUDRect // alias ActionSummary
	ActorInfo        HUDRect // unused
	HoverInfo        HUDRect // unused
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
	// TitleH kept for compatibility; v1 overlay uses LineH for title row only.
	m.TitleH = hudClamp(20*base, 16, 28)

	return m
}

// ComputeBattleHUDLayout builds a battle HUD layout for the given screen size and
// current battle state. It is the single source of truth for all major HUD rects.
// Uses b.LayoutStyle: 0 = v1 table, 1 = v2 Disciples-like.
func (b *BattleContext) ComputeBattleHUDLayout(screenW, screenH int) BattleHUDLayout {
	if b != nil && b.LayoutStyle == LayoutStyleV2Disciples {
		return b.computeLayoutV2(screenW, screenH)
	}
	return b.computeLayoutV1(screenW, screenH)
}

// computeLayoutV2 builds Disciples-like layout: TopBar, LeftRoster, RightRoster, Battlefield, BottomPanel.
// Proportions tuned so battlefield is the visual center; rosters and bottom panel frame it.
func (b *BattleContext) computeLayoutV2(screenW, screenH int) BattleHUDLayout {
	sw := float32(screenW)
	sh := float32(screenH)
	metrics := computeHUDMetrics(screenW, screenH)

	var layout BattleHUDLayout
	layout.Metrics = metrics
	layout.Style = LayoutStyleV2Disciples

	// TopBar: one status line only (minimal).
	topBarH := metrics.LineH + metrics.Pad*0.5
	if topBarH < metrics.LineH+4 {
		topBarH = metrics.LineH + 4
	}
	// BottomPanel: ~18% height — control panel, not a second table.
	bottomPanelH := hudClamp(sh*0.18, metrics.LineH*5, sh*0.22)
	// Rosters: ~16% width each so battlefield dominates.
	rosterW := hudClamp(sw*0.16, 130, sw*0.2)
	pad := metrics.Pad
	gap := metrics.Gap
	lineH := metrics.LineH
	btnH := metrics.ButtonH

	layout.TopBar = HUDRect{X: 0, Y: 0, W: sw, H: topBarH}
	layout.BottomPanel = HUDRect{X: 0, Y: sh - bottomPanelH, W: sw, H: bottomPanelH}
	layout.LeftRoster = HUDRect{X: 0, Y: topBarH, W: rosterW, H: sh - topBarH - bottomPanelH}
	layout.RightRoster = HUDRect{X: sw - rosterW, Y: topBarH, W: rosterW, H: sh - topBarH - bottomPanelH}
	layout.Battlefield = HUDRect{
		X: rosterW,
		Y: topBarH,
		W: sw - rosterW*2,
		H: sh - topBarH - bottomPanelH,
	}
	if layout.Battlefield.W < 0 {
		layout.Battlefield.W = 0
	}

	layout.V2TopBarLine = hudInset(layout.TopBar, pad*0.5)

	// BottomPanel: clear vertical order — Active|Target → Ability row → Summary → Hint → Buttons.
	inner := hudInset(layout.BottomPanel, pad)
	y := inner.Y

	// Row 1: Active (left) | Target (right), one line each
	infoH := lineH
	colW := (inner.W - gap) / 2
	layout.V2BottomActive = HUDRect{X: inner.X, Y: y, W: colW, H: infoH}
	layout.V2BottomTarget = HUDRect{X: inner.X + colW + gap, Y: y, W: inner.W - colW - gap, H: infoH}
	y += infoH + metrics.SmallGap

	// Ability row
	abilityRowH := lineH * 1.4
	layout.AbilityItemRects = nil
	if b != nil {
		active := b.ActiveUnit()
		if active != nil && active.Side == TeamPlayer && b.Phase == PhaseAwaitAction {
			abs := SpecialAbilities(active)
			if len(abs) > 0 {
				itemW := (inner.W - gap*float32(len(abs)-1)) / float32(len(abs))
				if itemW < 56 {
					itemW = 56
				}
				rects := make([]HUDRect, 0, len(abs))
				for i := range abs {
					x := inner.X + float32(i)*(itemW+gap)
					if x+itemW > inner.X+inner.W {
						break
					}
					rects = append(rects, HUDRect{X: x, Y: y, W: itemW, H: abilityRowH})
				}
				layout.AbilityItemRects = rects
			}
		}
	}
	y += abilityRowH + metrics.SmallGap

	// Summary: up to 2 lines (ability → target | preview)
	summaryH := lineH * 2
	if summaryH > inner.Y+inner.H-y-btnH-gap-lineH-gap {
		summaryH = lineH
	}
	layout.V2BottomSummary = HUDRect{X: inner.X, Y: y, W: inner.W, H: summaryH}
	y += summaryH + metrics.SmallGap

	// Hint line (secondary)
	layout.V2BottomLog = HUDRect{X: inner.X, Y: y, W: inner.W, H: lineH}
	y += lineH + gap

	// Buttons at bottom
	buttonsY := inner.Y + inner.H - btnH
	layout.BackButton = HUDRect{X: inner.X, Y: buttonsY, W: (inner.W - gap) / 2, H: btnH}
	layout.ConfirmButton = HUDRect{X: inner.X + inner.W - layout.BackButton.W, Y: buttonsY, W: layout.BackButton.W, H: btnH}

	// UnitRects: vertical cards with clearer spacing (Gap between cards).
	layout.UnitRects = map[UnitID]HUDRect{}
	if b != nil {
		cardPad := metrics.Pad
		for _, side := range []BattleSide{BattleSidePlayer, BattleSideEnemy} {
			roster := layout.LeftRoster
			if side == BattleSideEnemy {
				roster = layout.RightRoster
			}
			innerR := hudInset(roster, cardPad)
			cardGap := metrics.Gap
			nSlots := 6
			cardH := (innerR.H - float32(nSlots-1)*cardGap) / float32(nSlots)
			if cardH < metrics.LineH*2.2 {
				cardH = metrics.LineH * 2.2
			}
			idx := 0
			for _, row := range []BattleRow{BattleRowFront, BattleRowBack} {
				for i := 0; i < 3; i++ {
					slot := b.Slot(side, row, i)
					if slot == nil || slot.Occupied == 0 {
						continue
					}
					u := b.Units[slot.Occupied]
					if u == nil || !u.IsAlive() {
						continue
					}
					slotY := innerR.Y + float32(idx)*(cardH+cardGap)
					if slotY+cardH > innerR.Y+innerR.H {
						break
					}
					layout.UnitRects[u.ID] = HUDRect{X: innerR.X, Y: slotY, W: innerR.W, H: cardH}
					idx++
				}
			}
		}
	}

	// Legacy/alias placeholders so v1 code paths don't read garbage
	layout.Overlay = layout.Battlefield
	layout.Content = layout.Battlefield
	return layout
}

// computeLayoutV1 builds the original table-style layout (Overlay, Formation, Middle, Footer).
func (b *BattleContext) computeLayoutV1(screenW, screenH int) BattleHUDLayout {
	sw := float32(screenW)
	sh := float32(screenH)

	metrics := computeHUDMetrics(screenW, screenH)

	var layout BattleHUDLayout
	layout.Metrics = metrics
	layout.Style = LayoutStyleV1Table

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

	// 2) Rigid grid v1: TitleRow (1 line), small gap, then Content. No TitleH — use LineH everywhere.
	content := hudInset(layout.Overlay, metrics.Pad)
	layout.TitleRow = HUDRect{X: content.X, Y: content.Y, W: content.W, H: metrics.LineH}
	content.Y += metrics.LineH + metrics.SmallGap
	content.H -= metrics.LineH + metrics.SmallGap
	if content.H < 0 {
		content.H = 0
	}

	if b != nil && b.Result != ResultNone {
		used := metrics.LineH * 2
		content.Y += used
		content.H -= used
		if content.H < 0 {
			content.H = 0
		}
	}
	layout.Content = content

	// 3) Info rows: fixed LineH each.
	layout.TopInfoPrimary = HUDRect{X: content.X, Y: content.Y, W: content.W, H: metrics.LineH}
	layout.TopInfoSecondary = HUDRect{X: content.X, Y: content.Y + metrics.LineH, W: content.W, H: metrics.LineH}
	layout.InfoRow1 = layout.TopInfoPrimary
	layout.InfoRow2 = layout.TopInfoSecondary
	layout.InfoLine = HUDRect{X: content.X, Y: content.Y, W: content.W, H: metrics.LineH * 2}

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
		controlsPadBottom := metrics.SmallGap
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
		listTop := inner.Y + titleH + metrics.SmallGap
		listH := inner.Y + inner.H - listTop
		if listH < metrics.LineH*2 {
			listH = metrics.LineH * 2
		}
		if listH < 0 {
			listH = 0
		}
		layout.AbilitiesTitleRow = HUDRect{X: inner.X, Y: inner.Y, W: inner.W, H: titleH}
		layout.AbilityList = HUDRect{X: inner.X, Y: listTop, W: inner.W, H: listH}
	}

	// 9) Action panel: ActionTitleRow (one line) + ActionSummary + ActionButtons.
	if layout.Action.W > 0 && layout.Action.H > 0 {
		inner := hudInset(layout.Action, metrics.Pad*0.6)

		btnH := metrics.ButtonH
		summaryToButtonsGap := metrics.Gap
		buttonsY := inner.Y + inner.H - btnH
		layout.ActionButtons = HUDRect{X: inner.X, Y: buttonsY, W: inner.W, H: btnH}

		btnW := (inner.W - metrics.Gap) / 2
		if btnW < inner.W*0.3 {
			btnW = inner.W * 0.3
		}
		layout.BackButton = HUDRect{X: inner.X, Y: buttonsY, W: btnW, H: btnH}
		layout.ConfirmButton = HUDRect{X: inner.X + inner.W - btnW, Y: buttonsY, W: btnW, H: btnH}

		titleH := metrics.LineH
		summaryTop := inner.Y + titleH + metrics.SmallGap
		summaryBottom := buttonsY - summaryToButtonsGap
		summaryH := summaryBottom - summaryTop
		if summaryH < metrics.LineH*2 {
			summaryH = metrics.LineH * 2
		}
		if summaryH < 0 {
			summaryH = 0
		}
		layout.ActionTitleRow = HUDRect{X: inner.X, Y: inner.Y, W: inner.W, H: titleH}
		layout.ActionSummary = HUDRect{X: inner.X, Y: summaryTop, W: inner.W, H: summaryH}
	}

	// 10) Ability item rects (only meaningful on player turn).
	layout.AbilityItemRects = nil
	active := b.ActiveUnit()
	if active != nil && active.Side == TeamPlayer && b.Phase == PhaseAwaitAction {
		abs := SpecialAbilities(active)
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

	// Legacy compatibility: aliases and unused rects. v1 render does not use these.
	layout.AbilityHeader = layout.AbilitiesTitleRow
	layout.AbilityTooltip = HUDRect{}
	layout.ActionMain = layout.ActionSummary
	layout.ActorInfo = HUDRect{}
	layout.HoverInfo = HUDRect{}

	return layout
}

