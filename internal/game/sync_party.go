package game

import (
	battlepkg "mygame/internal/battle"
)

// syncPartyFromBattle переносит исход боя в канонические Hero: CurrentHP по живым/павшим союзникам.
// Вызывается пока BattleContext ещё жив; маппинг по CombatUnitOrigin.PartyActiveIndex.
// Участники с CurrentHP==0 не попадали в бой — их запись не трогаем.
func (g *Game) syncPartyFromBattle() {
	if g.battle == nil {
		return
	}
	for _, u := range g.battle.Units {
		if u == nil || u.Side != battlepkg.TeamPlayer {
			continue
		}
		idx := u.Origin.PartyActiveIndex
		if idx < 0 || idx >= len(g.party.Active) {
			continue
		}
		h := &g.party.Active[idx]
		if u.IsAlive() {
			h.CurrentHP = u.State.HP
			if h.CurrentHP > h.MaxHP {
				h.CurrentHP = h.MaxHP
			}
		} else {
			h.CurrentHP = 0
		}
	}
}
