package postbattle

import (
	battlepkg "mygame/internal/battle"
	"mygame/internal/progression"
	"mygame/internal/ui"
)

// BuildPostBattleParams собирает параметры отрисовки overlay из текущего состояния flow.
func BuildPostBattleParams(f *Flow, screenW, screenH int) ui.PostBattleParams {
	var resultText string
	switch f.Outcome {
	case battlepkg.BattleOutcomeVictory:
		resultText = "Victory!"
	case battlepkg.BattleOutcomeDefeat:
		resultText = "Defeat"
	case battlepkg.BattleOutcomeRetreat:
		resultText = "Escaped"
	default:
		resultText = "Battle ended"
	}
	params := ui.PostBattleParams{
		ResultText:    resultText,
		IsRewardStep:  f.Step == StepReward,
		SelectedIndex: f.SelectedIndex,
		ScreenWidth:   screenW,
		ScreenHeight:  screenH,
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
