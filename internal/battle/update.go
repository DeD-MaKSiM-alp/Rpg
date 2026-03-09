package battle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const actionPauseFrames = 30

// Update обрабатывает один кадр боевого режима и возвращает итог.
func (b *BattleContext) Update() BattleOutcome {
	if b == nil {
		return BattleOutcomeNone
	}

	// PhaseFinishedWaitInput: ждём подтверждения (SPACE/ENTER) перед закрытием.
	if b.Phase == PhaseFinishedWaitInput {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
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
		if active.Team == TeamPlayer {
			if b.CanPlayerActNow() && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				action, ok := BuildFirstAvailablePlayerAction(b, active)
				if ok {
					result := ResolveAbility(b, action)
					b.ApplyActionResult(result)
					b.PauseFrames = actionPauseFrames
					b.Phase = PhaseActionPause
				}
			}
		} else {
			action, ok := BuildEnemyAction(b, active)
			if ok {
				result := ResolveAbility(b, action)
				b.ApplyActionResult(result)
			}
			b.PauseFrames = actionPauseFrames
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
