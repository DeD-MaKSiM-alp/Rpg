package game

import (
	battlepkg "mygame/internal/battle"
)

// PlayerProgression — персистентные боевые параметры игрока между боями (источник истины для следующего боя).
type PlayerProgression struct {
	MaxHP      int
	Attack     int
	Defense    int
	Initiative int
	Abilities  []battlepkg.AbilityID
}

// DefaultProgression возвращает стартовую прогрессию (то же, что раньше было захардкожено в DefaultPlayerCombatUnitSeed).
func DefaultProgression() PlayerProgression {
	return PlayerProgression{
		MaxHP:      10,
		Attack:     2,
		Defense:    0,
		Initiative: 2,
		Abilities:  []battlepkg.AbilityID{battlepkg.AbilityBasicAttack},
	}
}

// RewardKind — тип награды после победы (один выбор за бой).
type RewardKind int

const (
	RewardMaxHP RewardKind = iota
	RewardAttack
	RewardInitiative
	RewardAbilityHeal
)

// RewardOptions — список наград, предлагаемых после победы (порядок отображения).
var RewardOptions = []RewardKind{
	RewardMaxHP,
	RewardAttack,
	RewardInitiative,
	RewardAbilityHeal,
}

// ApplyReward применяет выбранную награду к прогрессии.
func ApplyReward(prog *PlayerProgression, r RewardKind) {
	switch r {
	case RewardMaxHP:
		prog.MaxHP += 2
	case RewardAttack:
		prog.Attack += 1
	case RewardInitiative:
		prog.Initiative += 1
	case RewardAbilityHeal:
		if !hasAbility(prog.Abilities, battlepkg.AbilityHeal) {
			prog.Abilities = append(prog.Abilities, battlepkg.AbilityHeal)
		}
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
	case RewardInitiative:
		return "+1 Initiative"
	case RewardAbilityHeal:
		return "Unlock: Heal"
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
		return "Increases basic attack damage"
	case RewardInitiative:
		return "Act earlier in turn order"
	case RewardAbilityHeal:
		return "Gain Heal ability (ally)"
	default:
		return ""
	}
}
