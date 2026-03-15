package game

import (
	"math/rand"
	battlepkg "mygame/internal/battle"
)

// PlayerProgression — персистентные боевые параметры игрока между боями (источник истины для следующего боя).
type PlayerProgression struct {
	MaxHP            int
	Attack           int
	Defense          int
	Initiative       int
	HealPower        int   // 0 = default 2 in resolve
	BasicAttackBonus int   // extra damage for basic attack only
	Abilities        []battlepkg.AbilityID
}

// DefaultProgression возвращает стартовую прогрессию.
func DefaultProgression() PlayerProgression {
	return PlayerProgression{
		MaxHP:      10,
		Attack:     2,
		Defense:    0,
		Initiative: 2,
		HealPower:  0,
		Abilities:  []battlepkg.AbilityID{battlepkg.AbilityBasicAttack},
	}
}

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

const rewardOfferCount = 3

// CanOffer возвращает true, если награду r имеет смысл предлагать при текущей прогрессии (нет дублей/бесполезных).
func CanOffer(prog *PlayerProgression, r RewardKind) bool {
	switch r {
	case RewardAbilityHeal:
		return !hasAbility(prog.Abilities, battlepkg.AbilityHeal)
	case RewardAbilityRanged:
		return !hasAbility(prog.Abilities, battlepkg.AbilityRangedAttack)
	case RewardHealUpgrade:
		return hasAbility(prog.Abilities, battlepkg.AbilityHeal)
	default:
		return true
	}
}

// GenerateRewardOffer возвращает 2 или 3 награды из пула, применимых к текущей прогрессии (без дублей unlock и т.п.).
func GenerateRewardOffer(prog *PlayerProgression, count int) []RewardKind {
	var valid []RewardKind
	for _, r := range rewardPool {
		if CanOffer(prog, r) {
			valid = append(valid, r)
		}
	}
	if len(valid) == 0 {
		return nil
	}
	if count > len(valid) {
		count = len(valid)
	}
	// Shuffle and take first count
	shuffled := make([]RewardKind, len(valid))
	copy(shuffled, valid)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return shuffled[:count]
}

// ApplyReward применяет выбранную награду к прогрессии.
func ApplyReward(prog *PlayerProgression, r RewardKind) {
	switch r {
	case RewardMaxHP:
		prog.MaxHP += 2
	case RewardAttack:
		prog.Attack += 1
	case RewardDefense:
		prog.Defense += 1
	case RewardInitiative:
		prog.Initiative += 1
	case RewardAbilityHeal:
		if !hasAbility(prog.Abilities, battlepkg.AbilityHeal) {
			prog.Abilities = append(prog.Abilities, battlepkg.AbilityHeal)
		}
	case RewardAbilityRanged:
		if !hasAbility(prog.Abilities, battlepkg.AbilityRangedAttack) {
			prog.Abilities = append(prog.Abilities, battlepkg.AbilityRangedAttack)
		}
	case RewardHealUpgrade:
		prog.HealPower += 2
	case RewardBasicAttackUpgrade:
		prog.BasicAttackBonus += 1
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
