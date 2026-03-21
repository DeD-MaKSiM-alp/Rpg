// Package postbattle orchestrates the post-battle flow: result screen → optional reward selection → end.
package postbattle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
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
}

// Reset сбрасывает flow (при выходе из боя или старте нового).
func (f *Flow) Reset() {
	f.Step = StepNone
	f.Outcome = battlepkg.BattleOutcomeNone
	f.RewardOffer = nil
	f.SelectedIndex = 0
}

// Begin запускает post-battle после завершённого боя (показ результата).
func (f *Flow) Begin(outcome battlepkg.BattleOutcome) {
	f.Step = StepResult
	f.Outcome = outcome
	f.RewardOffer = nil
	f.SelectedIndex = 0
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

// Update обрабатывает один кадр ввода в post-battle. Если нужно завершить бой и вернуться в мир — возвращает true.
func (f *Flow) Update(leader *hero.Hero, screenW, screenH int) (endBattle bool) {
	if f.Step == StepNone {
		return false
	}
	n := len(f.RewardOffer)
	if n == 0 {
		n = 1
	}
	f.SelectedIndex = wrapRewardIndex(f.SelectedIndex, n)

	switch f.Step {
	case StepResult:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if f.Outcome == battlepkg.BattleOutcomeVictory {
				f.RewardOffer = progression.GenerateRewardOffer(leader, progression.RewardOfferCount)
				if len(f.RewardOffer) == 0 {
					return true
				}
				f.Step = StepReward
				f.SelectedIndex = 0
			} else {
				return true
			}
		}
	case StepReward:
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			f.SelectedIndex = wrapRewardIndex(f.SelectedIndex-1, n)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			f.SelectedIndex = wrapRewardIndex(f.SelectedIndex+1, n)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			progression.ApplyReward(leader, f.RewardOffer[f.SelectedIndex])
			return true
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if idx := f.rewardOptionAtCursor(screenW, screenH); idx >= 0 && idx < n {
				progression.ApplyReward(leader, f.RewardOffer[idx])
				return true
			}
		}
	}
	return false
}

func (f *Flow) rewardOptionAtCursor(screenW, screenH int) int {
	mx, my := ebiten.CursorPosition()
	layout := ui.ComputePostBattleLayout(screenW, screenH, true, len(f.RewardOffer))
	return layout.RewardOptionIndexAt(mx, my)
}
