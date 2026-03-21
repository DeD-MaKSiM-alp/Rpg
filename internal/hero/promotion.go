package hero

import (
	"errors"
	"fmt"

	battlepkg "mygame/internal/battle"
	"mygame/internal/unitdata"
)

// Ошибки promotion — проверяйте через errors.Is.
var (
	ErrPromotionNilHero          = errors.New("hero: nil hero")
	ErrPromotionNoUnitID         = errors.New("hero: promotion requires UnitID (legacy)")
	ErrPromotionUnknownCurrent   = errors.New("hero: current template not in registry")
	ErrPromotionNoPath           = errors.New("hero: no UpgradeToUnitID for this template")
	ErrPromotionTargetMissing    = errors.New("hero: upgrade target template not in registry")
	ErrPromotionSelfLoop         = errors.New("hero: UpgradeToUnitID points to self")
)

// preserveHPRatioOnPromotion сохраняет долю текущего HP при смене MaxHP (округление к ближайшему).
// 0 HP остаётся 0 (павший не «оживает» от promotion).
func preserveHPRatioOnPromotion(oldHP, oldMax, newMax int) int {
	if newMax <= 0 {
		return 0
	}
	if oldMax <= 0 {
		return newMax
	}
	if oldHP <= 0 {
		return 0
	}
	newHP := (oldHP*newMax + oldMax/2) / oldMax
	if newHP < 1 {
		newHP = 1
	}
	if newHP > newMax {
		newHP = newMax
	}
	return newHP
}

// validatePromotionDomain — только доменные условия шаблона (без лагеря / карты / UI).
func validatePromotionDomain(h *Hero) error {
	if h == nil {
		return ErrPromotionNilHero
	}
	if h.UnitID == "" {
		return ErrPromotionNoUnitID
	}
	cur, ok := unitdata.GetUnitTemplate(h.UnitID)
	if !ok {
		return ErrPromotionUnknownCurrent
	}
	if cur.UpgradeToUnitID == "" {
		return ErrPromotionNoPath
	}
	if cur.UpgradeToUnitID == h.UnitID {
		return ErrPromotionSelfLoop
	}
	if _, ok := unitdata.GetUnitTemplate(cur.UpgradeToUnitID); !ok {
		return fmt.Errorf("%w: %q", ErrPromotionTargetMissing, cur.UpgradeToUnitID)
	}
	return nil
}

// ValidatePromotionDomain экспортирует доменную проверку для gameplay-слоя (gating до TryPromoteHero).
func ValidatePromotionDomain(h *Hero) error {
	return validatePromotionDomain(h)
}

// TryPromoteHero переводит героя на следующий шаблон по UpgradeToUnitID.
// Сохраняет: CombatExperience, BasicAttackBonus (награды лидера), RecruitLabel.
// Пересобирает из целевого шаблона: статы, способности, UnitID.
// CurrentHP: доля от старого MaxHP переносится на новый MaxHP (см. preserveHPRatioOnPromotion).
// Политика «только в лагере» не проверяется здесь — вызывайте ValidatePromotionDomain + gating снаружи.
func TryPromoteHero(h *Hero) error {
	if err := validatePromotionDomain(h); err != nil {
		return err
	}
	cur, _ := unitdata.GetUnitTemplate(h.UnitID)
	next, _ := unitdata.GetUnitTemplate(cur.UpgradeToUnitID)

	oldHP, oldMax := h.CurrentHP, h.MaxHP
	xp := h.CombatExperience
	bonus := h.BasicAttackBonus
	label := h.RecruitLabel

	abils := make([]battlepkg.AbilityID, len(next.Abilities))
	copy(abils, next.Abilities)

	newMax := next.MaxHP
	newHP := preserveHPRatioOnPromotion(oldHP, oldMax, newMax)

	h.UnitID = next.UnitID
	h.MaxHP = newMax
	h.CurrentHP = newHP
	h.Attack = next.Attack
	h.Defense = next.Defense
	h.Initiative = next.Initiative
	h.HealPower = next.HealPower
	h.Abilities = abils
	h.CombatExperience = xp
	h.BasicAttackBonus = bonus
	h.RecruitLabel = label
	return nil
}

// PromotionUILine — строка для карточки: домен + флаг «стоим на лагере» (atCamp из world).
func PromotionUILine(h *Hero, atCamp bool) string {
	if err := validatePromotionDomain(h); err != nil {
		return PromotionErrUserMessage(err)
	}
	cur, _ := unitdata.GetUnitTemplate(h.UnitID)
	if !atCamp {
		return fmt.Sprintf("Повышение: только в лагере (лазурный маркер). Следующий: %s", cur.UpgradeToUnitID)
	}
	return fmt.Sprintf("Повышение: P — перейти к «%s»", cur.UpgradeToUnitID)
}

// PromotionErrUserMessage — короткий текст для баннера при ошибке promotion.
func PromotionErrUserMessage(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ErrPromotionNilHero):
		return "Повышение: нет героя."
	case errors.Is(err, ErrPromotionNoUnitID):
		return "Повышение: нет unit_id (legacy)."
	case errors.Is(err, ErrPromotionUnknownCurrent):
		return "Повышение: текущий шаблон не найден."
	case errors.Is(err, ErrPromotionNoPath):
		return "Повышение: нет следующего шага."
	case errors.Is(err, ErrPromotionSelfLoop):
		return "Повышение: ошибка данных (цикл)."
	case errors.Is(err, ErrPromotionTargetMissing):
		return "Повышение: целевой шаблон отсутствует в реестре."
	default:
		return err.Error()
	}
}

