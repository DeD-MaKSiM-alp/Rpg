// Package postbattle orchestrates the post-battle flow: result screen → optional reward selection → end.
package postbattle

import (
	"github.com/hajimehoshi/ebiten/v2"

	battlepkg "mygame/internal/battle"
	"mygame/internal/party"
	"mygame/internal/progression"
	"mygame/internal/ui"
)

// Step — шаг post-battle flow (экран результата → выбор награды при победе).
type Step int

const (
	StepNone Step = iota
	StepResult
	StepReward
)

// Flow хранит состояние post-battle и обрабатывает ввод до возврата в explore.
type Flow struct {
	Step          Step
	Outcome       battlepkg.BattleOutcome
	RewardOffer   []progression.RewardKind
	SelectedIndex int
	// VictorySummaryLines — краткая сводка прогрессии (только победа, шаг результата).
	VictorySummaryLines []string
}

// Reset сбрасывает flow (при выходе из боя или старте нового).
func (f *Flow) Reset() {
	f.Step = StepNone
	f.Outcome = battlepkg.BattleOutcomeNone
	f.RewardOffer = nil
	f.SelectedIndex = 0
	f.VictorySummaryLines = nil
}

// Begin запускает post-battle после завершённого боя (показ результата).
// victorySummary — строки сводки прогрессии (только при победе); иначе nil.
func (f *Flow) Begin(outcome battlepkg.BattleOutcome, victorySummary []string) {
	f.Step = StepResult
	f.Outcome = outcome
	f.RewardOffer = nil
	f.SelectedIndex = 0
	f.VictorySummaryLines = nil
	if outcome == battlepkg.BattleOutcomeVictory && len(victorySummary) > 0 {
		f.VictorySummaryLines = victorySummary
	}
}

// IsActive возвращает true, пока игрок на экране результата/награды, а не в активном бою.
func (f *Flow) IsActive() bool {
	return f.Step != StepNone
}

func wrapRewardIndex(idx, n int) int {
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

// confirmResultStep — один канонический переход с экрана результата (победа → награда, иначе → в мир).
func (f *Flow) confirmResultStep(roster *party.Party) (endBattle bool) {
	if f.Step != StepResult {
		return false
	}
	leader := roster.Leader()
	if leader == nil {
		return true
	}
	if f.Outcome == battlepkg.BattleOutcomeVictory {
		f.RewardOffer = progression.GenerateRewardOffer(leader, progression.RewardOfferCount)
		if len(f.RewardOffer) == 0 {
			return true
		}
		f.Step = StepReward
		f.SelectedIndex = 0
		return false
	}
	return true
}

// confirmRewardSelection применяет выбранную награду и завершает post-battle.
func (f *Flow) confirmRewardSelection(roster *party.Party, rewardIndex int) (endBattle bool) {
	if f.Step != StepReward {
		return false
	}
	leader := roster.Leader()
	if leader == nil {
		return true
	}
	n := len(f.RewardOffer)
	if n <= 0 || rewardIndex < 0 || rewardIndex >= n {
		return false
	}
	progression.ApplyReward(leader, f.RewardOffer[rewardIndex])
	return true
}

// Update обрабатывает один кадр ввода в post-battle. Если нужно завершить бой и вернуться в мир — возвращает true.
// Выбор награды после победы применяется к лидеру (party.Leader); боевой опыт партии начисляется в game до этого flow.
func (f *Flow) Update(roster *party.Party, screenW, screenH int) (endBattle bool) {
	if f.Step == StepNone {
		return false
	}
	leader := roster.Leader()
	if leader == nil {
		return true
	}
	kbd := PollPostBattleKeyboardIntents()
	mb := pollPostBattleMouseButtons()

	n := len(f.RewardOffer)
	if n == 0 {
		n = 1
	}
	f.SelectedIndex = wrapRewardIndex(f.SelectedIndex, n)

	switch f.Step {
	case StepResult:
		if kbd.Confirm {
			return f.confirmResultStep(roster)
		}
		if mb.LeftJustPressed {
			mx, my := ebiten.CursorPosition()
			layout := ui.ComputePostBattleLayout(screenW, screenH, false, 0, len(f.VictorySummaryLines))
			if layout.HitResultContinue(mx, my) {
				return f.confirmResultStep(roster)
			}
		}
	case StepReward:
		n = len(f.RewardOffer)
		if kbd.Prev {
			f.SelectedIndex = wrapRewardIndex(f.SelectedIndex-1, n)
		}
		if kbd.Next {
			f.SelectedIndex = wrapRewardIndex(f.SelectedIndex+1, n)
		}
		if kbd.Confirm {
			return f.confirmRewardSelection(roster, f.SelectedIndex)
		}
		if mb.LeftJustPressed {
			mx, my := ebiten.CursorPosition()
			layout := ui.ComputePostBattleLayout(screenW, screenH, true, n, 0)
			if idx := layout.RewardOptionIndexAt(mx, my); idx >= 0 && idx < n {
				return f.confirmRewardSelection(roster, idx)
			}
			if layout.HitRewardConfirm(mx, my) {
				return f.confirmRewardSelection(roster, f.SelectedIndex)
			}
		}
	}
	return false
}
