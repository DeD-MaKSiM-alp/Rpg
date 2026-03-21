package progression

import (
	battlepkg "mygame/internal/battle"
	"mygame/internal/party"
)

// CombatExperiencePerVictorySurvivor — сколько боевого опыта получает каждый выживший участник активного строя за победу.
const CombatExperiencePerVictorySurvivor = 1

// ApplyVictoryCombatXPForActiveSurvivors начисляет CombatExperience героям в roster.Active по итогам победного боя.
// Условия: юнит стороны игрока, валидный PartyActiveIndex, юнит жив в конце боя.
// Резерв не участвует (нет юнита в бою). Вызывать после syncPartyFromBattle, пока BattleContext ещё доступен.
func ApplyVictoryCombatXPForActiveSurvivors(b *battlepkg.BattleContext, roster *party.Party) {
	if b == nil || roster == nil {
		return
	}
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
		roster.Active[idx].CombatExperience += CombatExperiencePerVictorySurvivor
	}
}
