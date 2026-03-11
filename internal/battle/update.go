package battle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const actionPauseFrames = 30

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
	b.PlayerTurn.SelectedAbilityIndex = 0
	if abs := actor.Abilities(); len(abs) > 0 {
		b.PlayerTurn.SelectedAbilityID = abs[0]
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
	b.ensurePlayerTurnInitialized(actor)

	p := &b.PlayerTurn
	abilities := actor.Abilities()
	if len(abilities) == 0 {
		b.LastMessage = "У юнита нет способностей."
		return BattleAction{}, false
	}

	// Keep selection in-bounds.
	p.SelectedAbilityIndex = wrapIndex(p.SelectedAbilityIndex, len(abilities))
	p.SelectedAbilityID = abilities[p.SelectedAbilityIndex]

	ability := GetAbility(p.SelectedAbilityID)

	switch p.Phase {
	case PlayerChooseAbility:
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
			// If ability requires a target, enumerate.
			switch ability.TargetRule {
			case TargetEnemySingle, TargetAllySingle:
				targets, v := ListValidTargets(b, actor.ID, p.SelectedAbilityID)
				if !v.OK {
					b.LastMessage = v.Message
					return BattleAction{}, false
				}
				if len(targets) == 0 {
					b.LastMessage = "Нет валидных целей."
					return BattleAction{}, false
				}
				p.ValidTargets = targets
				p.SelectedTargetIdx = 0
				p.SelectedTarget = p.ValidTargets[0]
				p.Phase = PlayerChooseTarget
				return BattleAction{}, false

			case TargetSelf:
				p.SelectedTarget = SelfTarget()
				p.Pending = ActionRequest{Actor: actor.ID, Ability: p.SelectedAbilityID, Target: p.SelectedTarget}
				p.Phase = PlayerConfirmAction
				return BattleAction{}, false

			default:
				// Groundwork for no-target abilities.
				p.SelectedTarget = NoTarget()
				p.Pending = ActionRequest{Actor: actor.ID, Ability: p.SelectedAbilityID, Target: p.SelectedTarget}
				p.Phase = PlayerConfirmAction
				return BattleAction{}, false
			}
		}

	case PlayerChooseTarget:
		if justBackPressed() {
			p.Phase = PlayerChooseAbility
			p.ValidTargets = nil
			p.SelectedTargetIdx = 0
			p.SelectedTarget = NoTarget()
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
			b.LastMessage = "Нет валидных целей."
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
			p.Pending = ActionRequest{Actor: actor.ID, Ability: p.SelectedAbilityID, Target: p.SelectedTarget}
			p.Phase = PlayerConfirmAction
			return BattleAction{}, false
		}

	case PlayerConfirmAction:
		if justBackPressed() {
			// Back to previous step.
			if ability.TargetRule == TargetEnemySingle || ability.TargetRule == TargetAllySingle {
				p.Phase = PlayerChooseTarget
			} else {
				p.Phase = PlayerChooseAbility
			}
			p.Pending = ActionRequest{}
			return BattleAction{}, false
		}
		if justConfirmPressed() {
			// Re-validate on confirm.
			v := ValidateAction(b, p.Pending)
			if !v.OK {
				b.LastMessage = v.Message
				// Send user back to a recoverable step.
				if ability.TargetRule == TargetEnemySingle || ability.TargetRule == TargetAllySingle {
					p.Phase = PlayerChooseTarget
				} else {
					p.Phase = PlayerChooseAbility
				}
				return BattleAction{}, false
			}
			act, v2 := ToBattleAction(b, p.Pending)
			if !v2.OK {
				b.LastMessage = v2.Message
				p.Phase = PlayerChooseAbility
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

	// PhaseFinishedWaitInput: ждём подтверждения (SPACE/ENTER) перед закрытием.
	if b.Phase == PhaseFinishedWaitInput {
		if justConfirmPressed() {
			return b.ToBattleOutcome()
		}
		return BattleOutcomeNone
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		b.Result = ResultEscape
		b.LastMessage = "Игрок покинул бой."
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
			action, ok := b.updatePlayerTurnStateMachine(active)
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
