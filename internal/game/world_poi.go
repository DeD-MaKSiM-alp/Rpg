package game

import (
	"fmt"
	"math/rand"

	"mygame/internal/hero"
	"mygame/internal/party"
	"mygame/world"
)

func healExploreFraction(p *party.Party, divisor int) {
	if p == nil || len(p.Active) == 0 {
		return
	}
	one := func(h *hero.Hero) {
		if h == nil || h.CurrentHP <= 0 {
			return
		}
		gain := h.MaxHP / divisor
		if gain < 1 {
			gain = 1
		}
		h.CurrentHP += gain
		if h.CurrentHP > h.MaxHP {
			h.CurrentHP = h.MaxHP
		}
	}
	for i := range p.Active {
		one(&p.Active[i])
	}
	for i := range p.Reserve {
		one(&p.Reserve[i])
	}
	if !p.HasFightableMember() {
		p.Active[0].CurrentHP = 1
	}
}

// applySpringHealExplore — источник: заметное, но слабее отдыха R.
func applySpringHealExplore(p *party.Party) {
	healExploreFraction(p, 8)
}

// applyCampfireHealExplore — привал: лёгкое восстановление.
func applyCampfireHealExplore(p *party.Party) {
	healExploreFraction(p, 12)
}

// applyRuinsCombatXP — руины: +amount боевого опыта каждому живому в активном строю.
func applyRuinsCombatXP(p *party.Party, amount int) {
	if p == nil || amount <= 0 {
		return
	}
	for i := range p.Active {
		if p.Active[i].CurrentHP > 0 {
			p.Active[i].CombatExperience += amount
		}
	}
}

// applyRuinsTrapDamage — «засада» в руинах: фиксированный урон по живым в строю (не ниже 1 ОЗ).
func applyRuinsTrapDamage(p *party.Party) {
	if p == nil {
		return
	}
	for i := range p.Active {
		if p.Active[i].CurrentHP > 0 {
			p.Active[i].CurrentHP -= 2
			if p.Active[i].CurrentHP < 1 {
				p.Active[i].CurrentHP = 1
			}
		}
	}
	if !p.HasFightableMember() && len(p.Active) > 0 {
		p.Active[0].CurrentHP = 1
	}
}

// applyPOIWorldEffect — true, если pu — POI и эффект применён (ход мира дальше снаружи).
func (g *Game) applyPOIWorldEffect(pu world.PickupInteractionResult) bool {
	switch pu {
	case world.PickupInteractPOISpring:
		applySpringHealExplore(&g.party)
		g.setExplorePOIMsg("Источник: восстановление ОЗ")
		return true
	case world.PickupInteractPOICache:
		g.pickupCount += 3
		g.setExplorePOIMsg("Тайник: +3 к добыче")
		return true
	case world.PickupInteractPOICampfire:
		g.TrainingMarks++
		applyCampfireHealExplore(&g.party)
		g.setExplorePOIMsg("Привал: +1 знак и лёгкое лечение")
		return true
	default:
		return false
	}
}

func (g *Game) setExplorePOIMsg(s string) {
	g.explorePOIMsg = s
	g.explorePOIMsgTicks = exploreRestFeedbackDurationTicks
}

// --- POI с выбором (руины / алтарь): эффекты после подтверждения в ModePOIChoice ---

func (g *Game) applyRuinsPOIChoiceSafe() {
	applyRuinsCombatXP(&g.party, 1)
	g.setExplorePOIMsg("Руины: +1 боевого опыта каждому в строю")
}

func (g *Game) applyRuinsPOIChoiceRisky() {
	if rand.Intn(2) == 0 {
		applyRuinsCombatXP(&g.party, 3)
		g.setExplorePOIMsg("Руины: удача — +3 боевого опыта каждому в строю")
		return
	}
	applyRuinsTrapDamage(&g.party)
	g.setExplorePOIMsg("Руины: засада — каждый в строю потерял 2 ОЗ")
}

func (g *Game) applyAltarPOIChoiceModest() {
	g.TrainingMarks++
	g.setExplorePOIMsg("Алтарь: +1 знак обучения")
}

func (g *Game) applyAltarPOIChoiceBold() {
	g.TrainingMarks += 2
	loss := 0
	if lh := g.party.Leader(); lh != nil && lh.CurrentHP > 0 {
		loss = lh.MaxHP / 5
		if loss < 1 {
			loss = 1
		}
		lh.CurrentHP -= loss
		if lh.CurrentHP < 1 {
			lh.CurrentHP = 1
		}
	}
	g.setExplorePOIMsg(fmt.Sprintf("Алтарь: +2 знака; лидер −%d ОЗ", loss))
}
