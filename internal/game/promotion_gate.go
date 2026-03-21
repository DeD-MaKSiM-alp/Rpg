package game

import (
	"fmt"

	"mygame/internal/hero"
)

// TrainingMarksPerVictory — начисление за победу в бою (см. applyVictoryTrainingMarks).
const TrainingMarksPerVictory = 1

// PromotionGateResult — оценка «можно ли применить promotion сейчас» (gameplay), без мутации героя.
type PromotionGateResult struct {
	Allowed bool
	// Message — текст баннера при отказе; при Allowed=true не используется.
	Message string
	// Cost — стоимость в знаках для выбранной цели; иначе 0.
	Cost int
}

// EvaluatePromotionGate: домен + лагерь + знаки. selectedTargetUnitID — цель для ветвления; при одной цели может быть "" (берётся единственная).
func EvaluatePromotionGate(h *hero.Hero, atCamp bool, trainingMarks int, selectedTargetUnitID string) PromotionGateResult {
	if err := hero.ValidatePromotionPathsExist(h); err != nil {
		return PromotionGateResult{Allowed: false, Message: hero.PromotionErrUserMessage(err)}
	}
	targets, err := hero.PromotionTargetUnitIDs(h)
	if err != nil || len(targets) == 0 {
		return PromotionGateResult{Allowed: false, Message: hero.PromotionErrUserMessage(hero.ErrPromotionNoPath)}
	}
	var effective string
	if len(targets) == 1 {
		effective = targets[0]
	} else {
		if selectedTargetUnitID == "" {
			return PromotionGateResult{
				Allowed: false,
				Message: "Повышение: сначала выберите ветку (←/→).",
				Cost:    0,
			}
		}
		ok := false
		for _, id := range targets {
			if id == selectedTargetUnitID {
				effective = id
				ok = true
				break
			}
		}
		if !ok {
			return PromotionGateResult{Allowed: false, Message: "Повышение: недопустимая ветка.", Cost: 0}
		}
	}
	cost, ok := promotionCostForTargetUnitID(effective)
	if !ok {
		return PromotionGateResult{
			Allowed: false,
			Message: "Повышение: не удалось определить стоимость (данные).",
			Cost:    0,
		}
	}
	if !atCamp {
		return PromotionGateResult{
			Allowed: false,
			Message: "Повышение: встаньте на лагерь (лазурный маркер).",
			Cost:    cost,
		}
	}
	if trainingMarks < cost {
		return PromotionGateResult{
			Allowed: false,
			Message: fmt.Sprintf(
				"Повышение: не хватает знаков обучения (%d/%d).",
				trainingMarks, cost,
			),
			Cost: cost,
		}
	}
	return PromotionGateResult{Allowed: true, Message: "", Cost: cost}
}
