package game

import "mygame/internal/hero"

// PromotionGateResult — оценка «можно ли применить promotion сейчас» (gameplay), без мутации героя.
type PromotionGateResult struct {
	Allowed bool
	// Message — текст баннера при отказе; при Allowed=true не используется.
	Message string
}

// EvaluatePromotionGate: домен (как TryPromoteHero до применения) + политика «только в лагере наёмников».
// atCamp — игрок стоит на клетке с активным PickupKindRecruitCamp (лазурный маркер).
func EvaluatePromotionGate(h *hero.Hero, atCamp bool) PromotionGateResult {
	if err := hero.ValidatePromotionDomain(h); err != nil {
		return PromotionGateResult{Allowed: false, Message: hero.PromotionErrUserMessage(err)}
	}
	if !atCamp {
		return PromotionGateResult{Allowed: false, Message: "Повышение: встаньте на лагерь (лазурный маркер)."}
	}
	return PromotionGateResult{Allowed: true, Message: ""}
}
