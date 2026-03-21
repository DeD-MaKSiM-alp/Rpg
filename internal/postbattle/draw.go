package postbattle

import (
	"github.com/hajimehoshi/ebiten/v2"

	battlepkg "mygame/internal/battle"
	"mygame/internal/progression"
	"mygame/internal/ui"
)

// BuildPostBattleParams собирает параметры отрисовки overlay из текущего состояния flow.
func BuildPostBattleParams(f *Flow, screenW, screenH int) ui.PostBattleParams {
	var resultText string
	switch f.Outcome {
	case battlepkg.BattleOutcomeVictory:
		resultText = "Победа!"
	case battlepkg.BattleOutcomeDefeat:
		resultText = "Поражение"
	case battlepkg.BattleOutcomeRetreat:
		resultText = "Отступление"
	default:
		resultText = "Бой завершён"
	}
	isReward := f.Step == StepReward
	optN := len(f.RewardOffer)
	summaryN := len(f.VictorySummaryLines)
	if isReward {
		summaryN = 0
	}
	layout := ui.ComputePostBattleLayout(screenW, screenH, isReward, optN, summaryN)
	mx, my := ebiten.CursorPosition()

	params := ui.PostBattleParams{
		ResultText:            resultText,
		IsRewardStep:          isReward,
		SelectedIndex:         f.SelectedIndex,
		ScreenWidth:           screenW,
		ScreenHeight:          screenH,
		ConfirmRewardLabel:    "Подтвердить",
		VictorySummaryLines:   f.VictorySummaryLines,
		RewardPreambleLine:    "",
	}
	if f.Step == StepResult {
		if f.Outcome == battlepkg.BattleOutcomeVictory {
			params.ContinueButtonLabel = "Продолжить"
			params.ResultHintLine = "Пробел / Enter или кнопка — далее к награде лидеру"
		} else {
			params.ContinueButtonLabel = "В мир"
		}
		params.HoverContinue = layout.HitResultContinue(mx, my)
	}
	if isReward {
		params.RewardPreambleLine = "Награда только лидеру — отдельно от боевого опыта отряда."
		params.HoverRewardConfirm = layout.HitRewardConfirm(mx, my)
	}
	if params.IsRewardStep && len(f.RewardOffer) > 0 {
		params.OptionLabels = make([]string, len(f.RewardOffer))
		params.OptionDescs = make([]string, len(f.RewardOffer))
		for i := range f.RewardOffer {
			params.OptionLabels[i] = progression.RewardLabel(f.RewardOffer[i])
			params.OptionDescs[i] = progression.RewardDescription(f.RewardOffer[i])
		}
	}
	return params
}
