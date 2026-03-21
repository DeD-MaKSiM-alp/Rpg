package battle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const actionPauseFrames = 30

// justConfirmPressed: Space/Enter — execute current action (choose target or execute ability). No longer "confirm" semantics.
func justConfirmPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

func justBackPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyBackspace)
}

func justPrevPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft)
}

func justNextPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight)
}

func wrapIndex(idx, n int) int {
	if n <= 0 {
		return 0
	}
	for idx < 0 {
		idx += n
	}
	if idx >= n {
		idx %= n
	}
	return idx
}

func (b *BattleContext) ensurePlayerTurnInitialized(actor *BattleUnit) {
	if b == nil || actor == nil {
		return
	}
	if b.PlayerTurn.IsActiveFor(actor.ID) {
		return
	}
	b.PlayerTurn.Reset()
	b.PlayerTurn.Phase = PlayerChooseAbility
	b.PlayerTurn.Actor = actor.ID
	// Default = basic attack (click enemy to attack); special abilities only in the list.
	if HasBasicAttack(actor) {
		b.PlayerTurn.SelectedAbilityID = AbilityBasicAttack
		b.PlayerTurn.SelectedAbilityIndex = 0
	} else if abs := actor.Abilities(); len(abs) > 0 {
		b.PlayerTurn.SelectedAbilityID = abs[0]
		b.PlayerTurn.SelectedAbilityIndex = 0
	}
	b.PlayerTurn.ValidTargets = nil
	b.PlayerTurn.SelectedTargetIdx = 0
	b.PlayerTurn.SelectedTarget = NoTarget()
	b.PlayerTurn.Pending = ActionRequest{}
}

func (b *BattleContext) updatePlayerTurnStateMachine(actor *BattleUnit) (BattleAction, bool) {
	if b == nil || actor == nil || actor.Side != TeamPlayer {
		return BattleAction{}, false
	}
	if b.BlockPlayerInput {
		return BattleAction{}, false
	}
	b.ensurePlayerTurnInitialized(actor)

	p := &b.PlayerTurn
	abilities := actor.Abilities()
	if len(abilities) == 0 {
		b.AddBattleLog("У юнита нет способностей.")
		return BattleAction{}, false
	}

	// Keep selection in-bounds.
	p.SelectedAbilityIndex = wrapIndex(p.SelectedAbilityIndex, len(abilities))
	p.SelectedAbilityID = abilities[p.SelectedAbilityIndex]

	ability := GetAbility(p.SelectedAbilityID)

	switch p.Phase {
	case PlayerChooseAbility:
		if justBackPressed() {
			// Return to default attack mode (cancel special ability selection).
			if HasBasicAttack(actor) && p.SelectedAbilityID != AbilityBasicAttack {
				p.SelectedAbilityID = AbilityBasicAttack
				p.SelectedAbilityIndex = 0
				p.ValidTargets = nil
				p.SelectedTargetIdx = 0
				p.SelectedTarget = NoTarget()
				p.Pending = ActionRequest{}
			}
		}
		if justPrevPressed() {
			p.SelectedAbilityIndex = wrapIndex(p.SelectedAbilityIndex-1, len(abilities))
			p.SelectedAbilityID = abilities[p.SelectedAbilityIndex]
			p.ValidTargets = nil
			p.SelectedTargetIdx = 0
			p.SelectedTarget = NoTarget()
		}
		if justNextPressed() {
			p.SelectedAbilityIndex = wrapIndex(p.SelectedAbilityIndex+1, len(abilities))
			p.SelectedAbilityID = abilities[p.SelectedAbilityIndex]
			p.ValidTargets = nil
			p.SelectedTargetIdx = 0
			p.SelectedTarget = NoTarget()
		}

		if justConfirmPressed() {
			// If ability requires a target, go to target selection; otherwise execute immediately.
			switch ability.TargetRule {
			case TargetEnemySingle, TargetAllySingle:
				targets, v := ListValidTargets(b, actor.ID, p.SelectedAbilityID)
				if !v.OK {
					b.AddBattleLog(v.Message)
					return BattleAction{}, false
				}
				if len(targets) == 0 {
					b.AddBattleLog("Нет валидных целей.")
					return BattleAction{}, false
				}
				p.ValidTargets = targets
				p.SelectedTargetIdx = 0
				p.SelectedTarget = p.ValidTargets[0]
				p.Phase = PlayerChooseTarget
				return BattleAction{}, false

			case TargetSelf:
				p.SelectedTarget = SelfTarget()
				req := ActionRequest{Actor: actor.ID, Ability: p.SelectedAbilityID, Target: p.SelectedTarget}
				if v := ValidateAction(b, req); !v.OK {
					b.AddBattleLog(v.Message)
					return BattleAction{}, false
				}
				act, v2 := ToBattleAction(b, req)
				if !v2.OK {
					b.AddBattleLog(v2.Message)
					return BattleAction{}, false
				}
				p.Phase = PlayerResolveAction
				return act, true

			case TargetAllyTeam:
				// Массовое лечение: без выбора цели, сразу выполнение.
				fallthrough
			default:
				// No-target abilities: execute immediately.
				p.SelectedTarget = NoTarget()
				req := ActionRequest{Actor: actor.ID, Ability: p.SelectedAbilityID, Target: p.SelectedTarget}
				if v := ValidateAction(b, req); !v.OK {
					b.AddBattleLog(v.Message)
					return BattleAction{}, false
				}
				act, v2 := ToBattleAction(b, req)
				if !v2.OK {
					b.AddBattleLog(v2.Message)
					return BattleAction{}, false
				}
				p.Phase = PlayerResolveAction
				return act, true
			}
		}

	case PlayerChooseTarget:
		if justBackPressed() {
			p.Phase = PlayerChooseAbility
			if HasBasicAttack(actor) {
				p.SelectedAbilityID = AbilityBasicAttack
				p.SelectedAbilityIndex = 0
			}
			p.ValidTargets = nil
			p.SelectedTargetIdx = 0
			p.SelectedTarget = NoTarget()
			p.Pending = ActionRequest{}
			return BattleAction{}, false
		}
		if len(p.ValidTargets) == 0 {
			// Re-enumerate defensively (state may have changed).
			targets, _ := ListValidTargets(b, actor.ID, p.SelectedAbilityID)
			p.ValidTargets = targets
			p.SelectedTargetIdx = 0
			if len(p.ValidTargets) > 0 {
				p.SelectedTarget = p.ValidTargets[0]
			}
		}
		if len(p.ValidTargets) == 0 {
			b.AddBattleLog("Нет валидных целей.")
			p.Phase = PlayerChooseAbility
			return BattleAction{}, false
		}

		if justPrevPressed() {
			p.SelectedTargetIdx = wrapIndex(p.SelectedTargetIdx-1, len(p.ValidTargets))
			p.SelectedTarget = p.ValidTargets[p.SelectedTargetIdx]
		}
		if justNextPressed() {
			p.SelectedTargetIdx = wrapIndex(p.SelectedTargetIdx+1, len(p.ValidTargets))
			p.SelectedTarget = p.ValidTargets[p.SelectedTargetIdx]
		}
		if justConfirmPressed() {
			// Click on target = execute immediately (no Confirm phase).
			req := ActionRequest{Actor: actor.ID, Ability: p.SelectedAbilityID, Target: p.SelectedTarget}
			if v := ValidateAction(b, req); !v.OK {
				b.AddBattleLog(v.Message)
				return BattleAction{}, false
			}
			act, v2 := ToBattleAction(b, req)
			if !v2.OK {
				b.AddBattleLog(v2.Message)
				return BattleAction{}, false
			}
			p.Phase = PlayerResolveAction
			return act, true
		}

	case PlayerResolveAction:
		// This phase is consumed by battle loop; no input expected.
		return BattleAction{}, false

	default:
		p.Phase = PlayerChooseAbility
	}

	return BattleAction{}, false
}

// Update обрабатывает один кадр боевого режима и возвращает итог.
func (b *BattleContext) Update() BattleOutcome {
	if b == nil {
		return BattleOutcomeNone
	}
	b.tickFeedback()

	// PhaseFinishedWaitInput: ждём подтверждения (SPACE/ENTER) перед закрытием.
	if b.Phase == PhaseFinishedWaitInput {
		if justConfirmPressed() {
			return b.ToBattleOutcome()
		}
		return BattleOutcomeNone
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if b.SuppressEscThisFrame {
			return BattleOutcomeNone
		}
		// During player turn in special/target mode: Esc cancels to default attack. In default mode or other phases: retreat.
		if b.Phase == PhaseAwaitAction {
			if active := b.ActiveUnit(); active != nil && active.IsAlive() && active.Side == TeamPlayer {
				b.ensurePlayerTurnInitialized(active)
				pt := &b.PlayerTurn
				inSpecialOrTarget := pt.Phase == PlayerChooseTarget ||
					(pt.Phase == PlayerChooseAbility && HasBasicAttack(active) && pt.SelectedAbilityID != AbilityBasicAttack)
				if inSpecialOrTarget {
					pt.Phase = PlayerChooseAbility
					if HasBasicAttack(active) {
						pt.SelectedAbilityID = AbilityBasicAttack
						pt.SelectedAbilityIndex = 0
					}
					pt.ValidTargets = nil
					pt.SelectedTargetIdx = 0
					pt.SelectedTarget = NoTarget()
					pt.Pending = ActionRequest{}
					return BattleOutcomeNone
				}
			}
		}
		b.Result = ResultEscape
		b.AddBattleLog("Отступление.")
		b.Phase = PhaseFinishedWaitInput
		return BattleOutcomeNone
	}

	switch b.Phase {
	case PhaseStart:
		b.Phase = PhaseTurnStart
		return BattleOutcomeNone

	case PhaseTurnStart:
		b.UpdateResultIfFinished()
		if b.IsFinished() {
			b.Phase = PhaseFinishedWaitInput
			return BattleOutcomeNone
		}
		for b.TurnIndex < len(b.TurnOrder) {
			u := b.ActiveUnit()
			if u != nil && u.IsAlive() {
				b.Phase = PhaseAwaitAction
				return BattleOutcomeNone
			}
			b.TurnIndex++
		}
		b.Phase = PhaseRoundEnd
		return BattleOutcomeNone

	case PhaseAwaitAction:
		active := b.ActiveUnit()
		if active == nil || !active.IsAlive() {
			b.Phase = PhaseTurnEnd
			return BattleOutcomeNone
		}
		if active.Side == TeamPlayer {
			// Keyboard-driven state machine remains as fallback.
			action, ok := b.updatePlayerTurnStateMachine(active)
			if !ok {
				// Mouse layer can trigger an action as well; it only uses the same state machine fields.
				action, ok = b.updatePlayerTurnMouse(active)
			}
			if ok {
				result := ResolveAbility(b, action)
				b.ApplyActionResult(result)
				b.PauseFrames = actionPauseFrames
				b.PlayerTurn.Reset()
				b.Phase = PhaseActionPause
			}
		} else {
			action, ok := BuildEnemyAction(b, active)
			if ok {
				result := ResolveAbility(b, action)
				b.ApplyActionResult(result)
			}
			b.PauseFrames = actionPauseFrames
			b.PlayerTurn.Reset()
			b.Phase = PhaseActionPause
		}
		return BattleOutcomeNone

	case PhaseActionPause:
		b.PauseFrames--
		if b.PauseFrames <= 0 {
			if b.IsFinished() {
				b.Phase = PhaseFinishedWaitInput
			} else {
				b.Phase = PhaseTurnEnd
			}
		}
		return BattleOutcomeNone

	case PhaseTurnEnd:
		b.UpdateResultIfFinished()
		if b.IsFinished() {
			b.Phase = PhaseFinishedWaitInput
			return BattleOutcomeNone
		}
		b.AdvanceTurn()
		b.UpdateResultIfFinished()
		if b.IsFinished() {
			b.Phase = PhaseFinishedWaitInput
			return BattleOutcomeNone
		}
		b.Phase = PhaseTurnStart
		return BattleOutcomeNone

	case PhaseRoundEnd:
		b.Phase = PhaseTurnStart
		return BattleOutcomeNone
	}

	return BattleOutcomeNone
}
