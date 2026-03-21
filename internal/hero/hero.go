// Package hero holds one combat-capable character's persistent stats between battles.
// World position lives in player.Player; the roster lives in party.Party; battle runtime lives in battle.CombatUnit.
// Progression mutates hero.Hero (usually the party leader); CombatUnitSeed() projects into battle seeds.
package hero

import (
	battlepkg "mygame/internal/battle"
)

// Hero — состояние одного бойца между боями (статы, способности). Сборка отряда — в party.Party.
type Hero struct {
	MaxHP            int
	CurrentHP        int // каноническое HP между боями; 0 = недоступен для следующего боя (пока нет лечения/лагеря)
	Attack           int
	Defense          int
	Initiative       int
	HealPower        int // bonus HP healed (added to base 2); see battle.CombatUnit.HealPower
	BasicAttackBonus int // extra damage for basic attack only
	Abilities        []battlepkg.AbilityID
}

// DefaultHero возвращает стартового героя (эквивалент прежней DefaultProgression).
func DefaultHero() Hero {
	h := Hero{
		MaxHP:      10,
		Attack:     2,
		Defense:    0,
		Initiative: 2,
		HealPower:  0,
		Abilities:  []battlepkg.AbilityID{battlepkg.AbilityBasicAttack},
	}
	h.CurrentHP = h.MaxHP
	return h
}

// CanEnterBattle true, если герой может получить сид для боя (есть HP).
func (h *Hero) CanEnterBattle() bool {
	return h != nil && h.CurrentHP > 0
}

// CombatUnitSeed строит один combat seed; party.Party.PlayerCombatSeeds() собирает сиды всего активного ростера.
func (h *Hero) CombatUnitSeed() battlepkg.CombatUnitSeed {
	s := battlepkg.BuildPlayerCombatSeed(
		h.MaxHP,
		h.Attack,
		h.Defense,
		h.Initiative,
		h.Abilities,
		h.HealPower,
		h.BasicAttackBonus,
	)
	if h.CurrentHP > 0 {
		s.InitialHP = h.CurrentHP
		if s.InitialHP > h.MaxHP {
			s.InitialHP = h.MaxHP
		}
	}
	return s
}
