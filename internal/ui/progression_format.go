// progression_format.go — короткие строки для HUD / карточек; доменная логика в hero/progression не меняется.
package ui

import (
	"fmt"

	"mygame/internal/hero"
)

// CombatXPToNextBonusStep — сколько единиц боевого опыта до следующего +1 к бонусу базовой атаки только от CombatExperience.
func CombatXPToNextBonusStep(xp int) int {
	steps := hero.CombatXPStepsPerBasicAttackBonus
	next := ((xp / steps) + 1) * steps
	return next - xp
}

// FormatCombatXPInspectLines — 1–2 строки для карточек (состав F5·I, ПКМ в бою); формулировки совпадают с HUD/explore.
func FormatCombatXPInspectLines(h *hero.Hero) []string {
	if h == nil {
		return nil
	}
	xp := h.CombatExperience
	total := h.EffectiveBasicAttackBonusForCombat()
	base := h.BasicAttackBonus
	fromXP := xp / hero.CombatXPStepsPerBasicAttackBonus
	need := CombatXPToNextBonusStep(xp)
	return []string{
		fmt.Sprintf("Боевой опыт: %d · к базовой атаке +%d (награды лидера +%d · от опыта +%d)", xp, total, base, fromXP),
		fmt.Sprintf("До следующего +1 от опыта: ещё %d (шаг каждые %d).", need, hero.CombatXPStepsPerBasicAttackBonus),
	}
}

// FormatLeaderExploreStripLine — одна строка под заголовком полоски отряда в explore (лидер).
func FormatLeaderExploreStripLine(h *hero.Hero) string {
	if h == nil {
		return ""
	}
	xp := h.CombatExperience
	bonus := h.EffectiveBasicAttackBonusForCombat()
	need := CombatXPToNextBonusStep(xp)
	return fmt.Sprintf("Лидер: опыт %d · бонус к базовой атаке +%d · ещё %d до +1 от опыта", xp, bonus, need)
}

// FormatLeaderHUDProgressionLine — компактная строка верхнего HUD (обрезается по ширине экрана).
func FormatLeaderHUDProgressionLine(h *hero.Hero) string {
	if h == nil {
		return ""
	}
	xp := h.CombatExperience
	bonus := h.EffectiveBasicAttackBonusForCombat()
	need := CombatXPToNextBonusStep(xp)
	return fmt.Sprintf("Лидер: опыт %d · бонус к базовой атаке +%d · +1 от опыта через %d", xp, bonus, need)
}
