package battle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func pointInRect(x, y float32, r HUDRect) bool {
	return x >= r.X && y >= r.Y && x <= r.X+r.W && y <= r.Y+r.H
}

// pointHitsUnit — roster-карточка или токен на поле (v2), единый hit-test для одного UnitID.
func (l BattleHUDLayout) pointHitsUnit(id UnitID, mxf, myf float32) bool {
	if r, ok := l.UnitRects[id]; ok && pointInRect(mxf, myf, r) {
		return true
	}
	if l.BattlefieldTokens != nil {
		if r, ok := l.BattlefieldTokens[id]; ok && pointInRect(mxf, myf, r) {
			return true
		}
	}
	return false
}

func (b *BattleContext) updatePlayerTurnMouse(actor *BattleUnit) (BattleAction, bool) {
	if b == nil || actor == nil || actor.Side != TeamPlayer {
		return BattleAction{}, false
	}

	// Reset hover each frame; will be filled by hit-tests below.
	b.PlayerTurn.HoverAbilityIndex = -1
	b.PlayerTurn.HoverTargetUnitID = 0
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
			for id := range validSet {
				u := b.Units[id]
				if u == nil || u.Side == actor.Side {
					continue
				}
				if layout.pointHitsUnit(id, mxf, myf) {
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
						req := ActionRequest{Actor: actor.ID, Ability: abilID, Target: pt.SelectedTarget}
						if ValidateAction(b, req).OK {
							if act, v2 := ToBattleAction(b, req); v2.OK {
								pt.Phase = PlayerResolveAction
								return act, true
							}
						}
						return BattleAction{}, false

					default:
						pt.SelectedTarget = NoTarget()
						req := ActionRequest{Actor: actor.ID, Ability: abilID, Target: pt.SelectedTarget}
						if ValidateAction(b, req).OK {
							if act, v2 := ToBattleAction(b, req); v2.OK {
								pt.Phase = PlayerResolveAction
								return act, true
							}
						}
						return BattleAction{}, false
					}
				}
				break
			}
		}
	}

	// 2) Formation slots hit-test for targets (only when choosing target for special ability).
	if pt.Phase == PlayerChooseTarget {
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

	slotLoop:
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
				if !layout.pointHitsUnit(u.ID, mxf, myf) {
					continue
				}
				pt.HoverTargetUnitID = u.ID
				if leftClick && validSet[u.ID] && pt.Phase == PlayerChooseTarget {
					// Click on valid target = execute immediately (no Confirm phase).
					pt.SelectedTarget = UnitTarget(u.ID)
					req := ActionRequest{
						Actor:   actor.ID,
						Ability: pt.SelectedAbilityID,
						Target:  pt.SelectedTarget,
					}
					if ValidateAction(b, req).OK {
						if act, v2 := ToBattleAction(b, req); v2.OK {
							pt.Phase = PlayerResolveAction
							return act, true
						}
					}
				}
				break slotLoop
			}
		}
	}

	// 3) Action panel: Back button only (Confirm removed from UX). Back = cancel special ability / return to default attack.
	backBtn := layout.BackButton
	if backBtn.W > 0 && backBtn.H > 0 && (pt.Phase == PlayerChooseTarget || (pt.Phase == PlayerChooseAbility && pt.SelectedAbilityID != AbilityBasicAttack)) {
		if pointInRect(mxf, myf, backBtn) {
			pt.HoverBackButton = true
		}
		if leftClick && pointInRect(mxf, myf, backBtn) {
			pt.Phase = PlayerChooseAbility
			if HasBasicAttack(actor) {
				pt.SelectedAbilityID = AbilityBasicAttack
				pt.SelectedAbilityIndex = 0
			}
			pt.ValidTargets = nil
			pt.SelectedTargetIdx = 0
			pt.SelectedTarget = NoTarget()
			pt.Pending = ActionRequest{}
		}
	}

	// 4) Right click = quick cancel/back; return to default attack mode when going back to ChooseAbility.
	if rightClick {
		switch pt.Phase {
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
