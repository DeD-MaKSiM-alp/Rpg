// progression_format.go — короткие строки для HUD / карточек; доменная логика в hero/progression.
package ui

import (
	"fmt"

	"mygame/internal/hero"
)

// FormatCombatXPInspectLines — 1–2 строки для карточек (состав F5·I, ПКМ в бою); боевой уровень и опыт до следующего.
func FormatCombatXPInspectLines(h *hero.Hero) []string {
	if h == nil {
		return nil
	}
	lvl := h.CombatLevel()
	need := h.CombatXPToNextLevel()
	fromLevel := h.CombatAttackBonusFromLevel()
	reward := h.BasicAttackBonus
	total := h.EffectiveBasicAttackBonusForCombat()
	return []string{
		fmt.Sprintf("Боевой уровень %d · опыт %d · до следующего уровня ещё %d · бонус базовой атаки +%d (от уровня +%d · награды после боя +%d)", lvl, h.CombatExperience, need, total, fromLevel, reward),
		fmt.Sprintf("Каждые %d опыта — новый уровень и +1 к базовой атаке от уровня.", hero.CombatXPPerLevel),
	}
}

// FormatLeaderExploreStripLine — одна строка под заголовком полоски отряда в explore (лидер).
func FormatLeaderExploreStripLine(h *hero.Hero) string {
	if h == nil {
		return ""
	}
	return fmt.Sprintf("Лидер · ур.%d · опыт %d · ещё %d до ур. · атака +%d",
		h.CombatLevel(), h.CombatExperience, h.CombatXPToNextLevel(), h.EffectiveBasicAttackBonusForCombat())
}

// FormatLeaderHUDProgressionLine — компактная строка прогресса лидера (сейчас в основном для тестов; в explore — полоска отряда).
func FormatLeaderHUDProgressionLine(h *hero.Hero) string {
	if h == nil {
		return ""
	}
	return fmt.Sprintf("Лидер · ур.%d · опыт %d · ещё %d до ур. · атака +%d",
		h.CombatLevel(), h.CombatExperience, h.CombatXPToNextLevel(), h.EffectiveBasicAttackBonusForCombat())
}
