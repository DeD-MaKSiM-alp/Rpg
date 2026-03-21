package hero

import (
	"errors"
	"fmt"

	battlepkg "mygame/internal/battle"
	"mygame/internal/unitdata"
)

// Ошибки promotion — проверяйте через errors.Is.
var (
	ErrPromotionNilHero                 = errors.New("hero: nil hero")
	ErrPromotionNoUnitID                = errors.New("hero: promotion requires UnitID (legacy)")
	ErrPromotionUnknownCurrent          = errors.New("hero: current template not in registry")
	ErrPromotionNoPath                  = errors.New("hero: no UpgradeToUnitID for this template")
	ErrPromotionTargetMissing           = errors.New("hero: upgrade target template not in registry")
	ErrPromotionSelfLoop                = errors.New("hero: UpgradeToUnitID points to self")
	ErrPromotionBranchChoiceRequired    = errors.New("hero: choose promotion branch (TryPromoteHeroTo)")
	ErrPromotionInvalidTarget           = errors.New("hero: empty promotion target")
	ErrPromotionTargetNotAllowed        = errors.New("hero: target not in allowed promotion options")
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

// promotionTargetsFromHero — нормализованные цели из текущего шаблона (линейно или UpgradeOptions).
func promotionTargetsFromHero(h *Hero) ([]string, error) {
	if h == nil {
		return nil, ErrPromotionNilHero
	}
	if h.UnitID == "" {
		return nil, ErrPromotionNoUnitID
	}
	cur, ok := unitdata.GetUnitTemplate(h.UnitID)
	if !ok {
		return nil, ErrPromotionUnknownCurrent
	}
	ids := unitdata.PromotionTargetUnitIDs(cur)
	if len(ids) == 0 {
		return nil, ErrPromotionNoPath
	}
	for _, id := range ids {
		if id == h.UnitID {
			return nil, ErrPromotionSelfLoop
		}
		if _, ok := unitdata.GetUnitTemplate(id); !ok {
			return nil, fmt.Errorf("%w: %q", ErrPromotionTargetMissing, id)
		}
	}
	return ids, nil
}

// PromotionTargetUnitIDs — публичный доступ к списку допустимых целей повышения (для UI / policy).
func PromotionTargetUnitIDs(h *Hero) ([]string, error) {
	return promotionTargetsFromHero(h)
}

// ValidatePromotionPathsExist — домен: есть ли хотя бы один допустимый шаг (включая ветвление).
func ValidatePromotionPathsExist(h *Hero) error {
	_, err := promotionTargetsFromHero(h)
	return err
}

// ValidatePromotionDomain — алиас для gameplay: «путь promotion существует» (старое имя).
func ValidatePromotionDomain(h *Hero) error {
	return ValidatePromotionPathsExist(h)
}

func applyPromotionTo(h *Hero, targetUnitID string) error {
	next, ok := unitdata.GetUnitTemplate(targetUnitID)
	if !ok {
		return fmt.Errorf("%w: %q", ErrPromotionTargetMissing, targetUnitID)
	}

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

// TryPromoteHero переводит героя на единственный возможный следующий шаблон.
// Если доступны две и более ветки — ErrPromotionBranchChoiceRequired; используйте TryPromoteHeroTo.
func TryPromoteHero(h *Hero) error {
	targets, err := promotionTargetsFromHero(h)
	if err != nil {
		return err
	}
	if len(targets) > 1 {
		return ErrPromotionBranchChoiceRequired
	}
	return applyPromotionTo(h, targets[0])
}

// TryPromoteHeroTo переводит героя в указанный target UnitID; target должен быть в списке допустимых для текущего шаблона.
func TryPromoteHeroTo(h *Hero, targetUnitID string) error {
	if h == nil {
		return ErrPromotionNilHero
	}
	if targetUnitID == "" {
		return ErrPromotionInvalidTarget
	}
	targets, err := promotionTargetsFromHero(h)
	if err != nil {
		return err
	}
	allowed := false
	for _, id := range targets {
		if id == targetUnitID {
			allowed = true
			break
		}
	}
	if !allowed {
		return ErrPromotionTargetNotAllowed
	}
	return applyPromotionTo(h, targetUnitID)
}

// PromotionUILine — строка для карточки: домен + флаг «стоим на лагере» (atCamp из world).
func PromotionUILine(h *Hero, atCamp bool) string {
	if err := ValidatePromotionPathsExist(h); err != nil {
		return PromotionErrUserMessage(err)
	}
	targets, err := PromotionTargetUnitIDs(h)
	if err != nil || len(targets) == 0 {
		return PromotionErrUserMessage(ErrPromotionNoPath)
	}
	if len(targets) > 1 {
		if !atCamp {
			return "Повышение: только в лагере. Две ветки — выберите ←/→, затем P."
		}
		return "Повышение: ←/→ выбор ветки · P — применить (в лагере)"
	}
	if !atCamp {
		return fmt.Sprintf("Повышение: только в лагере (лазурный маркер). Следующий: %s", targets[0])
	}
	return fmt.Sprintf("Повышение: P — перейти к «%s»", targets[0])
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
	case errors.Is(err, ErrPromotionBranchChoiceRequired):
		return "Повышение: выберите ветку (←/→), затем P."
	case errors.Is(err, ErrPromotionInvalidTarget):
		return "Повышение: не выбрана цель."
	case errors.Is(err, ErrPromotionTargetNotAllowed):
		return "Повышение: эта ветка недоступна."
	default:
		return err.Error()
	}
}
