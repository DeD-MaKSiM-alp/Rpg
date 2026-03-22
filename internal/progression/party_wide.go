package progression

import (
	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
)

// CombatExperiencePerVictorySurvivor — сколько боевого опыта получает каждый выживший участник активного строя за победу.
const CombatExperiencePerVictorySurvivor = 1

// CombatLevelUp — повышение боевого уровня после начисления опыта (для сводки после боя).
type CombatLevelUp struct {
	PartyActiveIndex int
	OldLevel         int
	NewLevel         int
}

// ApplyVictoryCombatXPForActiveSurvivors начисляет CombatExperience героям в roster.Active по итогам победного боя.
// Условия: юнит стороны игрока, валидный PartyActiveIndex, юнит жив в конце боя.
// Резерв не участвует (нет юнита в бою). Вызывать после syncPartyFromBattle, пока BattleContext ещё доступен.
func ApplyVictoryCombatXPForActiveSurvivors(b *battlepkg.BattleContext, roster *party.Party) []CombatLevelUp {
	if b == nil || roster == nil {
		return nil
	}
	var ups []CombatLevelUp
	for _, u := range b.Units {
		if u == nil || u.Side != battlepkg.TeamPlayer {
			continue
		}
		idx := u.Origin.PartyActiveIndex
		if idx < 0 || idx >= len(roster.Active) {
			continue
		}
		if !u.IsAlive() {
			continue
		}
		h := &roster.Active[idx]
		oldXP := h.CombatExperience
		oldLevel := hero.CombatLevelFromTotalXP(oldXP)
		h.CombatExperience += CombatExperiencePerVictorySurvivor
		newLevel := hero.CombatLevelFromTotalXP(h.CombatExperience)
		if newLevel > oldLevel {
			ups = append(ups, CombatLevelUp{
				PartyActiveIndex: idx,
				OldLevel:         oldLevel,
				NewLevel:         newLevel,
			})
		}
	}
	return ups
}
