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

// ApplyReward применяет выбранную награду к каноническому герою.
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
		return "+2 Max HP"
	case RewardAttack:
		return "+1 Attack"
	case RewardDefense:
		return "+1 Defense"
	case RewardInitiative:
		return "+1 Initiative"
	case RewardAbilityHeal:
		return "Unlock: Heal"
	case RewardAbilityRanged:
		return "Unlock: Shoot"
	case RewardHealUpgrade:
		return "Heal +2"
	case RewardBasicAttackUpgrade:
		return "Basic Attack +1"
	default:
		return "?"
	}
}

// RewardDescription возвращает краткое описание для UI.
func RewardDescription(r RewardKind) string {
	switch r {
	case RewardMaxHP:
		return "Increases maximum health"
	case RewardAttack:
		return "Increases attack damage"
	case RewardDefense:
		return "Reduces incoming damage"
	case RewardInitiative:
		return "Act earlier in turn order"
	case RewardAbilityHeal:
		return "Gain Heal ability (ally)"
	case RewardAbilityRanged:
		return "Gain Shoot ability (ranged)"
	case RewardHealUpgrade:
		return "Heal restores more HP"
	case RewardBasicAttackUpgrade:
		return "Basic attack deals +1 damage"
	default:
		return ""
	}
}
