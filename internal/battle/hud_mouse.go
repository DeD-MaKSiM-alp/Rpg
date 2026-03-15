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
	pt := &b.PlayerTurn

	// 0) Default attack mode: click on valid enemy = immediate basic attack (no ability choice, no confirm).
	if pt.Phase == PlayerChooseAbility && pt.SelectedAbilityID == AbilityBasicAttack && HasBasicAttack(actor) {
		targets, v := ListValidTargets(b, actor.ID, AbilityBasicAttack)
		if v.OK && len(targets) > 0 {
			validSet := map[UnitID]bool{}
			for _, td := range targets {
				if td.Kind == TargetKindUnit {
					validSet[td.UnitID] = true
				}
			}
			for id, r := range layout.UnitRects {
				if !validSet[id] {
					continue
				}
				u := b.Units[id]
				if u == nil || u.Side == actor.Side {
					continue
				}
				if pointInRect(mxf, myf, r) {
					pt.HoverTargetUnitID = id
					if leftClick {
						req := ActionRequest{Actor: actor.ID, Ability: AbilityBasicAttack, Target: UnitTarget(id)}
						if ValidateAction(b, req).OK {
							if act, v2 := ToBattleAction(b, req); v2.OK {
								pt.Pending = req
								pt.Phase = PlayerResolveAction
								return act, true
							}
						}
					}
					break
				}
			}
		}
	}

	// 1) Ability panel: only special abilities (list excludes basic attack).
	specialAbs := SpecialAbilities(actor)
	if len(specialAbs) > 0 && len(layout.AbilityItemRects) > 0 {
		for i := 0; i < len(specialAbs) && i < len(layout.AbilityItemRects); i++ {
			rowRect := layout.AbilityItemRects[i]
			if pointInRect(mxf, myf, rowRect) {
				pt.HoverAbilityIndex = i
				if leftClick && pt.Phase == PlayerChooseAbility {
					abilID := specialAbs[i]
					pt.SelectedAbilityID = abilID
					// Keep SelectedAbilityIndex in sync with full ability list for keyboard.
					if full := actor.Abilities(); len(full) > 0 {
						for j := range full {
							if full[j] == abilID {
								pt.SelectedAbilityIndex = j
								break
							}
						}
					}

					ability := GetAbility(abilID)
					switch ability.TargetRule {
					case TargetEnemySingle, TargetAllySingle:
						targets, v := ListValidTargets(b, actor.ID, abilID)
						if !v.OK {
							b.AddBattleLog(v.Message)
							return BattleAction{}, false
						}
						if len(targets) == 0 {
							b.AddBattleLog("Нет валидных целей.")
							return BattleAction{}, false
						}
						pt.ValidTargets = targets
						pt.SelectedTargetIdx = 0
						pt.SelectedTarget = targets[0]
						pt.Pending = ActionRequest{}
						pt.Phase = PlayerChooseTarget
						return BattleAction{}, false

					case TargetSelf:
						pt.SelectedTarget = SelfTarget()
						pt.Pending = ActionRequest{Actor: actor.ID, Ability: abilID, Target: pt.SelectedTarget}
						pt.Phase = PlayerConfirmAction
						return BattleAction{}, false

					default:
						pt.SelectedTarget = NoTarget()
						pt.Pending = ActionRequest{Actor: actor.ID, Ability: abilID, Target: pt.SelectedTarget}
						pt.Phase = PlayerConfirmAction
						return BattleAction{}, false
					}
				}
				break
			}
		}
	}

	// 2) Formation slots hit-test for targets (only in target-related phases; not used for default attack).
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
				if ability.TargetRule == TargetEnemySingle || ability.TargetRule == TargetAllySingle {
					pt.Phase = PlayerChooseTarget
				} else {
					pt.Phase = PlayerChooseAbility
					if HasBasicAttack(actor) {
						pt.SelectedAbilityID = AbilityBasicAttack
						pt.SelectedAbilityIndex = 0
					}
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

	// 4) Right click = quick cancel/back; return to default attack mode when going back to ChooseAbility.
	if rightClick {
		ability := GetAbility(pt.SelectedAbilityID)
		switch pt.Phase {
		case PlayerConfirmAction:
			if ability.TargetRule == TargetEnemySingle || ability.TargetRule == TargetAllySingle {
				pt.Phase = PlayerChooseTarget
			} else {
				pt.Phase = PlayerChooseAbility
				if HasBasicAttack(actor) {
					pt.SelectedAbilityID = AbilityBasicAttack
					pt.SelectedAbilityIndex = 0
				}
			}
			pt.Pending = ActionRequest{}
		case PlayerChooseTarget:
			pt.Phase = PlayerChooseAbility
			if HasBasicAttack(actor) {
				pt.SelectedAbilityID = AbilityBasicAttack
				pt.SelectedAbilityIndex = 0
			}
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

