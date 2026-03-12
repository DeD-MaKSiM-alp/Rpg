package battle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func pointInRect(x, y float32, r HUDRect) bool {
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
	layout := b.ComputeBattleHUDLayout(screenW, screenH)

	action := BattleAction{}
	var performed bool

	// 1) Ability panel hit-test (only in PlayerChooseAbility; hover always).
	abilities := actor.Abilities()
	if len(abilities) > 0 && len(layout.AbilityItemRects) > 0 {
		for i := 0; i < len(abilities) && i < len(layout.AbilityItemRects); i++ {
			rowRect := layout.AbilityItemRects[i]
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
		}
	}

	// 2) Formation slots hit-test for targets (only in target-related phases).
	pt := &b.PlayerTurn
	if pt.Phase == PlayerChooseTarget || pt.Phase == PlayerConfirmAction {
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

		for _, row := range []BattleRow{BattleRowFront, BattleRowBack} {
			for i := 0; i < 3; i++ {
				slot := b.Slot(sideToScan, row, i)
				if slot == nil || slot.Occupied == 0 {
					continue
				}
				u := b.Units[slot.Occupied]
				if u == nil {
					continue
				}
				r, ok := layout.UnitRects[u.ID]
				if !ok {
					continue
				}
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
		backBtn := layout.BackButton
		confirmBtn := layout.ConfirmButton

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
				// Confirm: mirror keyboard confirm behavior with defensive Pending handling.
				req := pt.Pending
				if req.Actor == 0 || req.Ability == 0 {
					// Defensive fallback: reconstruct from current selection.
					req = ActionRequest{
						Actor:   actor.ID,
						Ability: pt.SelectedAbilityID,
						Target:  pt.SelectedTarget,
					}
				}

				v := ValidateAction(b, req)
				if !v.OK {
					if v.Message != "" {
						b.AddBattleLog("Cannot confirm action: " + v.Message)
					} else {
						b.AddBattleLog("Cannot confirm action: invalid action")
					}

					// Send the player back to a recoverable step, same as keyboard path.
					if ability.TargetRule == TargetEnemySingle || ability.TargetRule == TargetAllySingle {
						pt.Phase = PlayerChooseTarget
					} else {
						pt.Phase = PlayerChooseAbility
					}
					pt.Pending = ActionRequest{}
				} else {
					act, v2 := ToBattleAction(b, req)
					if v2.OK {
						action = act
						performed = true
						pt.Pending = req
						pt.Phase = PlayerResolveAction
					} else if v2.Message != "" {
						b.AddBattleLog("Cannot confirm action: " + v2.Message)
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

