// Package hero holds the canonical out-of-combat model of the player's combat identity (leader unit).
// World position lives in player.Player; battle runtime lives in battle.CombatUnit; this struct is the
// persistent bridge: progression mutates it, and CombatUnitSeed() projects it into a battle seed.
package hero

import (
	battlepkg "mygame/internal/battle"
)

// Hero — каноническое состояние «главного» бойца отряда между боями (статы, способности).
// В будущем сюда же логично добавить Party []Hero или ссылки на слоты отряда; пока один лидер = вся party.
type Hero struct {
	MaxHP            int
	Attack           int
	Defense          int
	Initiative       int
	HealPower        int // bonus HP healed (added to base 2); see battle.CombatUnit.HealPower
	BasicAttackBonus int // extra damage for basic attack only
	Abilities        []battlepkg.AbilityID
}

// DefaultHero возвращает стартового героя (эквивалент прежней DefaultProgression).
func DefaultHero() Hero {
	return Hero{
		MaxHP:      10,
		Attack:     2,
		Defense:    0,
		Initiative: 2,
		HealPower:  0,
		Abilities:  []battlepkg.AbilityID{battlepkg.AbilityBasicAttack},
	}
}

// CombatUnitSeed строит вход для battle.BuildBattleContextFromEncounter из канонического состояния.
// Единая точка проекции hero → бой; не дублировать разбор полей в game.
func (h *Hero) CombatUnitSeed() battlepkg.CombatUnitSeed {
	return battlepkg.BuildPlayerCombatSeed(
		h.MaxHP,
		h.Attack,
		h.Defense,
		h.Initiative,
		h.Abilities,
		h.HealPower,
		h.BasicAttackBonus,
	)
}
