package battle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// HUD layout constants duplicated from ui layer for mouse hit-testing.
const (
	hudLineH = float32(18)
	hudPad   = float32(12)
	hudGap   = float32(10)
)

type hudRect struct {
	X, Y, W, H float32
}

func hudClamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func hudInset(r hudRect, pad float32) hudRect {
	return hudRect{X: r.X + pad, Y: r.Y + pad, W: r.W - pad*2, H: r.H - pad*2}
}

func hudSplitH(r hudRect, leftW, gap float32) (hudRect, hudRect) {
	left := hudRect{X: r.X, Y: r.Y, W: leftW, H: r.H}
	right := hudRect{X: r.X + leftW + gap, Y: r.Y, W: r.W - leftW - gap, H: r.H}
	if right.W < 0 {
		right.W = 0
	}
	return left, right
}

// computeHUDLayout reproduces the battle HUD container layout for hit-testing.
func computeHUDLayout(screenW, screenH int) (overlay hudRect, formation hudRect, abilities hudRect, action hudRect) {
	sw := float32(screenW)
	sh := float32(screenH)

	marginX := hudClamp(sw*0.08, 12, 80)
	marginY := hudClamp(sh*0.08, 12, 80)
	panelW := sw - marginX*2
	panelH := sh - marginY*2
	panelW = hudClamp(panelW, 520, 760)
	panelH = hudClamp(panelH, 360, 540)

	panelX := (sw - panelW) / 2
	panelY := (sh - panelH) / 2
	overlay = hudRect{X: panelX, Y: panelY, W: panelW, H: panelH}

	content := hudInset(overlay, hudPad)
	content.Y += hudLineH // title
	content.H -= hudLineH
	if content.H < 0 {
		content.H = 0
	}

	afterInfo := hudRect{X: content.X, Y: content.Y + hudLineH + hudGap, W: content.W, H: content.H - hudLineH - hudGap}
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

	formation = hudRect{X: afterInfo.X, Y: afterInfo.Y, W: afterInfo.W, H: formationH}
	middle := hudRect{X: afterInfo.X, Y: formation.Y + formation.H + hudGap, W: afterInfo.W, H: middleH}
	// footer not needed for mouse right now

	colW := (formation.W - hudGap) / 2
	_, _ = hudSplitH(formation, colW, hudGap) // we only need overall formation rect for slots

	mColW := (middle.W - hudGap) / 2
	abilities, action = hudSplitH(middle, mColW, hudGap)
	return
}

func pointInRect(x, y float32, r hudRect) bool {
	return x >= r.X && y >= r.Y && x <= r.X+r.W && y <= r.Y+r.H
}

func (b *BattleContext) updatePlayerTurnMouse(actor *BattleUnit) (BattleAction, bool) {
	if b == nil || actor == nil || actor.Side != TeamPlayer {
		return BattleAction{}, false
	}

	// Reset hover each frame; will be filled by hit-tests below.
	b.PlayerTurn.HoverAbilityIndex = -1
	b.PlayerTurn.HoverTargetUnitID = 0
	b.PlayerTurn.HoverConfirmButton = false
	b.PlayerTurn.HoverBackButton = false

	// Only respond to mouse when it's player's turn and AwaitAction phase.
	if b.Phase != PhaseAwaitAction {
		return BattleAction{}, false
	}

	mx, my := ebiten.CursorPosition()
	mxf := float32(mx)
	myf := float32(my)
	leftClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

	screenW, screenH := ebiten.WindowSize()
	_, formationRect, abilitiesRect, actionRect := computeHUDLayout(screenW, screenH)

	action := BattleAction{}
	var performed bool

	// 1) Ability panel hit-test (only in PlayerChooseAbility; hover always).
	abilities := actor.Abilities()
	if len(abilities) > 0 {
		inner := hudInset(abilitiesRect, hudPad*0.6)
		y := inner.Y + hudLineH*2
		maxY := inner.Y + inner.H - hudLineH*0.5
		for i := 0; i < len(abilities) && y <= maxY; i++ {
			rowRect := hudRect{X: inner.X, Y: y - hudLineH*0.5, W: inner.W, H: hudLineH}
			if pointInRect(mxf, myf, rowRect) {
				b.PlayerTurn.HoverAbilityIndex = i
				if leftClick && b.PlayerTurn.Phase == PlayerChooseAbility {
					// Primary mouse flow: select ability and immediately advance subphase,
					// mirroring the keyboard confirm behavior.
					b.PlayerTurn.SelectedAbilityIndex = i
					b.PlayerTurn.SelectedAbilityID = abilities[i]

					ability := GetAbility(abilities[i])
					switch ability.TargetRule {
					case TargetEnemySingle, TargetAllySingle:
						targets, v := ListValidTargets(b, actor.ID, abilities[i])
						if !v.OK {
							b.AddBattleLog(v.Message)
							return BattleAction{}, false
						}
						if len(targets) == 0 {
							b.AddBattleLog("Нет валидных целей.")
							return BattleAction{}, false
						}
						b.PlayerTurn.ValidTargets = targets
						b.PlayerTurn.SelectedTargetIdx = 0
						b.PlayerTurn.SelectedTarget = targets[0]
						b.PlayerTurn.Pending = ActionRequest{} // pending будет сформирован на target step
						b.PlayerTurn.Phase = PlayerChooseTarget
						return BattleAction{}, false

					case TargetSelf:
						b.PlayerTurn.SelectedTarget = SelfTarget()
						b.PlayerTurn.Pending = ActionRequest{
							Actor:   actor.ID,
							Ability: abilities[i],
							Target:  b.PlayerTurn.SelectedTarget,
						}
						b.PlayerTurn.Phase = PlayerConfirmAction
						return BattleAction{}, false

					default:
						// No-target ability: immediately prepare pending and go to confirm.
						b.PlayerTurn.SelectedTarget = NoTarget()
						b.PlayerTurn.Pending = ActionRequest{
							Actor:   actor.ID,
							Ability: abilities[i],
							Target:  b.PlayerTurn.SelectedTarget,
						}
						b.PlayerTurn.Phase = PlayerConfirmAction
						return BattleAction{}, false
					}
				}
				break
			}
			y += hudLineH
		}
	}

	// 2) Formation slots hit-test for targets (only in target-related phases).
	pt := &b.PlayerTurn
	if pt.Phase == PlayerChooseTarget || pt.Phase == PlayerConfirmAction {
		inner := hudInset(formationRect, hudPad*0.6)
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

		// Precompute valid target set for quick lookup.
		validSet := map[UnitID]bool{}
		for _, td := range pt.ValidTargets {
			if td.Kind == TargetKindUnit {
				validSet[td.UnitID] = true
			}
		}

		// Determine which side's slots we are targeting:
		// enemy for enemy-target abilities, own side for ally-target.
		ability := GetAbility(pt.SelectedAbilityID)
		sideToScan := actor.Side
		if ability.TargetRule == TargetEnemySingle {
			sideToScan = b.EnemyTeam(actor.Side)
		}

		for rowIdx, row := range []BattleRow{BattleRowFront, BattleRowBack} {
			slotY := frontSlotsY
			if rowIdx == 1 {
				slotY = backSlotsY
			}
			for i := 0; i < 3; i++ {
				slot := b.Slot(sideToScan, row, i)
				if slot == nil || slot.Occupied == 0 {
					continue
				}
				u := b.Units[slot.Occupied]
				if u == nil {
					continue
				}
				x := inner.X + float32(i)*cellW
				r := hudRect{X: x, Y: slotY, W: cellW - 4, H: cellH}
				if pointInRect(mxf, myf, r) {
					pt.HoverTargetUnitID = u.ID
					if leftClick && validSet[u.ID] && pt.Phase == PlayerChooseTarget {
						// Select this target and move to confirm phase.
						pt.SelectedTarget = UnitTarget(u.ID)
						pt.Pending = ActionRequest{
							Actor:   actor.ID,
							Ability: pt.SelectedAbilityID,
							Target:  pt.SelectedTarget,
						}
						pt.Phase = PlayerConfirmAction
					}
					break
				}
			}
		}
	}

	// 3) Action panel buttons (Confirm / Back) hit-test.
	if pt.Phase == PlayerConfirmAction || pt.Phase == PlayerChooseTarget {
		inner := hudInset(actionRect, hudPad*0.6)
		btnH := hudLineH * 1.2
		btnW := inner.W * 0.45
		btnY := inner.Y + inner.H - btnH - hudPad*0.3

		backBtn := hudRect{X: inner.X, Y: btnY, W: btnW, H: btnH}
		confirmBtn := hudRect{X: inner.X + inner.W - btnW, Y: btnY, W: btnW, H: btnH}

		hasBack := pt.Phase == PlayerChooseTarget || pt.Phase == PlayerConfirmAction
		canConfirm := pt.Phase == PlayerConfirmAction

		if pointInRect(mxf, myf, backBtn) && hasBack {
			pt.HoverBackButton = true
		}
		if pointInRect(mxf, myf, confirmBtn) && canConfirm {
			pt.HoverConfirmButton = true
		}

		if leftClick {
			ability := GetAbility(pt.SelectedAbilityID)
			if hasBack && pointInRect(mxf, myf, backBtn) {
				// Back: same semantics as keyboard Back.
				if ability.TargetRule == TargetEnemySingle || ability.TargetRule == TargetAllySingle {
					pt.Phase = PlayerChooseTarget
				} else {
					pt.Phase = PlayerChooseAbility
				}
				pt.Pending = ActionRequest{}
			} else if canConfirm && pointInRect(mxf, myf, confirmBtn) {
				// Confirm: validate + ToBattleAction.
				v := ValidateAction(b, pt.Pending)
				if !v.OK {
					b.AddBattleLog(v.Message)
				} else {
					act, v2 := ToBattleAction(b, pt.Pending)
					if v2.OK {
						action = act
						performed = true
					} else if v2.Message != "" {
						b.AddBattleLog(v2.Message)
					}
				}
			}
		}
	}

	// 4) Right click = quick cancel/back.
	if rightClick {
		ability := GetAbility(pt.SelectedAbilityID)
		switch pt.Phase {
		case PlayerConfirmAction:
			if ability.TargetRule == TargetEnemySingle || ability.TargetRule == TargetAllySingle {
				pt.Phase = PlayerChooseTarget
			} else {
				pt.Phase = PlayerChooseAbility
			}
			pt.Pending = ActionRequest{}
		case PlayerChooseTarget:
			pt.Phase = PlayerChooseAbility
			pt.ValidTargets = nil
			pt.SelectedTargetIdx = 0
			pt.SelectedTarget = NoTarget()
			pt.Pending = ActionRequest{}
		default:
			// no-op for other phases
		}
	}

	return action, performed
}

