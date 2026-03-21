package progression

import (
	"math/rand"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
)

// RewardKind — тип награды после победы.
type RewardKind int

const (
	RewardMaxHP RewardKind = iota
	RewardAttack
	RewardDefense
	RewardInitiative
	RewardAbilityHeal
	RewardAbilityRanged
	RewardHealUpgrade
	RewardBasicAttackUpgrade
)

// rewardPool — полный пул наград (порядок для генерации оффера).
var rewardPool = []RewardKind{
	RewardMaxHP,
	RewardAttack,
	RewardDefense,
	RewardInitiative,
	RewardAbilityHeal,
	RewardAbilityRanged,
	RewardHealUpgrade,
	RewardBasicAttackUpgrade,
}

// RewardOfferCount — сколько вариантов награды показывать после победы.
const RewardOfferCount = 3

// CanOffer возвращает true, если награду r имеет смысл предлагать при текущей прогрессии (нет дублей/бесполезных).
func CanOffer(h *hero.Hero, r RewardKind) bool {
	switch r {
	case RewardAbilityHeal:
		return !hasAbility(h.Abilities, battlepkg.AbilityHeal)
	case RewardAbilityRanged:
		return !hasAbility(h.Abilities, battlepkg.AbilityRangedAttack)
	case RewardHealUpgrade:
		return hasAbility(h.Abilities, battlepkg.AbilityHeal)
	default:
		return true
	}
}

// GenerateRewardOffer возвращает награды из пула, применимых к текущей прогрессии (без дублей unlock и т.п.).
func GenerateRewardOffer(h *hero.Hero, count int) []RewardKind {
	var valid []RewardKind
	for _, r := range rewardPool {
		if CanOffer(h, r) {
			valid = append(valid, r)
		}
	}
	if len(valid) == 0 {
		return nil
	}
	if count > len(valid) {
		count = len(valid)
	}
	shuffled := make([]RewardKind, len(valid))
	copy(shuffled, valid)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return shuffled[:count]
}

// ApplyReward применяет выбранную награду к лидеру (канонический герой выбора после победы).
// Не дублирует party-wide боевой опыт (CombatExperience) — это отдельный слой.
func ApplyReward(h *hero.Hero, r RewardKind) {
	switch r {
	case RewardMaxHP:
		h.MaxHP += 2
		h.CurrentHP += 2
	case RewardAttack:
		h.Attack += 1
	case RewardDefense:
		h.Defense += 1
	case RewardInitiative:
		h.Initiative += 1
	case RewardAbilityHeal:
		if !hasAbility(h.Abilities, battlepkg.AbilityHeal) {
			h.Abilities = append(h.Abilities, battlepkg.AbilityHeal)
		}
	case RewardAbilityRanged:
		if !hasAbility(h.Abilities, battlepkg.AbilityRangedAttack) {
			h.Abilities = append(h.Abilities, battlepkg.AbilityRangedAttack)
		}
	case RewardHealUpgrade:
		h.HealPower += 2
	case RewardBasicAttackUpgrade:
		h.BasicAttackBonus += 1
	}
}

func hasAbility(abils []battlepkg.AbilityID, id battlepkg.AbilityID) bool {
	for _, a := range abils {
		if a == id {
			return true
		}
	}
	return false
}

// RewardLabel возвращает короткое название награды для UI.
func RewardLabel(r RewardKind) string {
	switch r {
	case RewardMaxHP:
		return "+2 макс. ОЗ"
	case RewardAttack:
		return "+1 атака"
	case RewardDefense:
		return "+1 защита"
	case RewardInitiative:
		return "+1 инициатива"
	case RewardAbilityHeal:
		return "Открыть: Лечение"
	case RewardAbilityRanged:
		return "Открыть: Выстрел"
	case RewardHealUpgrade:
		return "Лечение +2"
	case RewardBasicAttackUpgrade:
		return "Базовая атака +1"
	default:
		return "?"
	}
}

// RewardDescription возвращает краткое описание для UI.
func RewardDescription(r RewardKind) string {
	switch r {
	case RewardMaxHP:
		return "Больше максимального здоровья"
	case RewardAttack:
		return "Больше урона атак"
	case RewardDefense:
		return "Меньше входящего урона"
	case RewardInitiative:
		return "Раньше в порядке ходов"
	case RewardAbilityHeal:
		return "Способность «Лечение» (союзник)"
	case RewardAbilityRanged:
		return "Способность «Выстрел» (дальний бой)"
	case RewardHealUpgrade:
		return "Лечение восстанавливает больше ОЗ"
	case RewardBasicAttackUpgrade:
		return "Базовая атака: +1 урон"
	default:
		return ""
	}
}
