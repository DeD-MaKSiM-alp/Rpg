package game

import (
	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

// PromotionCostFromTargetTier — знаки обучения для повышения в шаблон с данным tier (цена = tier цели).
// Минимальная модель без таблиц в unitdata: tier 2 → 2, tier 3 → 3 и т.д.
func PromotionCostFromTargetTier(tier int) int {
	if tier < 1 {
		return 1
	}
	return tier
}

func promotionCostForTargetUnitID(targetUnitID string) (int, bool) {
	next, ok := unitdata.GetUnitTemplate(targetUnitID)
	if !ok {
		return 0, false
	}
	return PromotionCostFromTargetTier(next.Tier), true
}

// PromotionTrainingMarkCostForHeroTarget — цена для конкретной допустимой цели (ветка или линейный единственный шаг).
func PromotionTrainingMarkCostForHeroTarget(h *hero.Hero, targetUnitID string) (int, bool) {
	if h == nil || targetUnitID == "" {
		return 0, false
	}
	if err := hero.ValidatePromotionPathsExist(h); err != nil {
		return 0, false
	}
	targets, err := hero.PromotionTargetUnitIDs(h)
	if err != nil {
		return 0, false
	}
	ok := false
	for _, id := range targets {
		if id == targetUnitID {
			ok = true
			break
		}
	}
	if !ok {
		return 0, false
	}
	return promotionCostForTargetUnitID(targetUnitID)
}

// PromotionTrainingMarkCostForHero — только если ровно одна цель (линейный путь).
func PromotionTrainingMarkCostForHero(h *hero.Hero) (cost int, ok bool) {
	if h == nil {
		return 0, false
	}
	targets, err := hero.PromotionTargetUnitIDs(h)
	if err != nil || len(targets) != 1 {
		return 0, false
	}
	return promotionCostForTargetUnitID(targets[0])
}
