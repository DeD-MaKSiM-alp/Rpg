package battle

import (
	"fmt"
	"strings"
)

// Боевые ресурсы: mana / energy / cooldown по раундам. Базовая атака не тратит ресурсы и без КД.

const (
	DefaultBattleManaMax   = 12
	DefaultBattleEnergyMax = 12
)

// AbilityCost описывает стоимость и КД способности (0 = нет).
type AbilityCost struct {
	Mana            int
	Energy          int
	CooldownRounds  int
}

// AbilityIsFree — базовая атака не требует ресурсов и не уходит в КД.
func AbilityIsFree(id AbilityID) bool {
	return id == AbilityBasicAttack
}

// abilityCost возвращает стоимость из реестра.
func abilityCost(id AbilityID) AbilityCost {
	a := GetAbility(id)
	return AbilityCost{Mana: a.CostMana, Energy: a.CostEnergy, CooldownRounds: a.CooldownRounds}
}

// AbilityResourceGate — можно ли использовать способность по ресурсам и КД (без проверки целей).
func AbilityResourceGate(ctx *BattleContext, actor *BattleUnit, abilityID AbilityID) ValidationResult {
	if ctx == nil || actor == nil || !actor.IsAlive() {
		return errResult(ErrNoActor, "no actor")
	}
	if AbilityIsFree(abilityID) {
		return okResult()
	}
	c := abilityCost(abilityID)
	if rem, ok := actor.State.AbilityCooldowns[abilityID]; ok && rem > 0 {
		return errResult(ErrAbilityOnCooldown, "перезарядка: ещё %d ход(ов) раунда", rem)
	}
	if actor.State.Mana < c.Mana {
		return errResult(ErrInsufficientMana, "недостаточно маны (%d/%d)", actor.State.Mana, c.Mana)
	}
	if actor.State.Energy < c.Energy {
		return errResult(ErrInsufficientEnergy, "недостаточно энергии (%d/%d)", actor.State.Energy, c.Energy)
	}
	return okResult()
}

// AvailableAbilities — способности из loadout, доступные по ресурсу/КД (для UI и AI).
func (ctx *BattleContext) AvailableAbilities(actor *BattleUnit) []AbilityID {
	if ctx == nil || actor == nil {
		return nil
	}
	all := actor.Abilities()
	out := make([]AbilityID, 0, len(all))
	for _, id := range all {
		if AbilityResourceGate(ctx, actor, id).OK {
			out = append(out, id)
		}
	}
	return out
}

func applyAbilityCost(actor *BattleUnit, abilityID AbilityID) {
	if actor == nil || AbilityIsFree(abilityID) {
		return
	}
	c := abilityCost(abilityID)
	actor.State.Mana -= c.Mana
	if actor.State.Mana < 0 {
		actor.State.Mana = 0
	}
	actor.State.Energy -= c.Energy
	if actor.State.Energy < 0 {
		actor.State.Energy = 0
	}
	if c.CooldownRounds > 0 {
		if actor.State.AbilityCooldowns == nil {
			actor.State.AbilityCooldowns = make(map[AbilityID]int)
		}
		actor.State.AbilityCooldowns[abilityID] = c.CooldownRounds
	}
}

// tickRoundResources — конец раунда: КД −1, лёгкая регенерация (чтобы многоходовые бои оставались играбельны).
func (c *BattleContext) tickRoundResources() {
	if c == nil || c.Units == nil {
		return
	}
	for _, u := range c.Units {
		if u == nil || !u.IsAlive() {
			continue
		}
		if len(u.State.AbilityCooldowns) > 0 {
			for aid, rem := range u.State.AbilityCooldowns {
				if rem <= 1 {
					delete(u.State.AbilityCooldowns, aid)
				} else {
					u.State.AbilityCooldowns[aid] = rem - 1
				}
			}
		}
		rm, re := regenManaEnergyPerRound(resourceProfileForDef(&u.Def))
		for i := 0; i < rm && u.State.Mana < u.State.MaxMana; i++ {
			u.State.Mana++
		}
		for i := 0; i < re && u.State.Energy < u.State.MaxEnergy; i++ {
			u.State.Energy++
		}
	}
}

// SpecialAbilitiesUsable — специальные способности (без базовой атаки), доступные по ресурсу/КД.
func (ctx *BattleContext) SpecialAbilitiesUsable(actor *BattleUnit) []AbilityID {
	if ctx == nil || actor == nil {
		return nil
	}
	var out []AbilityID
	for _, id := range SpecialAbilities(actor) {
		if AbilityResourceGate(ctx, actor, id).OK {
			out = append(out, id)
		}
	}
	return out
}

func initCombatResources(st *CombatUnitState, def *CombatUnitDefinition) {
	if st == nil {
		return
	}
	p := resourceProfileForDef(def)
	applyResourceProfileToState(st, p)
}

// ActorResourceLineRU — строка маны/энергии с текущим и максимумом (HUD активного юнита).
func ActorResourceLineRU(u *BattleUnit) string {
	if u == nil {
		return ""
	}
	return fmt.Sprintf("Мана %d/%d · Энергия %d/%d", u.State.Mana, u.State.MaxMana, u.State.Energy, u.State.MaxEnergy)
}

// AbilityCostLinePlayerRU — стоимость способности для игрока: слова «мана»/«энергия», КД в раундах, оставшееся ожидание КД.
func AbilityCostLinePlayerRU(ctx *BattleContext, actor *BattleUnit, id AbilityID) string {
	if AbilityIsFree(id) {
		return ""
	}
	a := GetAbility(id)
	var parts []string
	if a.CostMana > 0 {
		parts = append(parts, fmt.Sprintf("мана %d", a.CostMana))
	}
	if a.CostEnergy > 0 {
		parts = append(parts, fmt.Sprintf("энергия %d", a.CostEnergy))
	}
	if a.CooldownRounds > 0 {
		parts = append(parts, fmt.Sprintf("КД %dр", a.CooldownRounds))
	}
	s := strings.Join(parts, " · ")
	if ctx != nil && actor != nil && actor.State.AbilityCooldowns != nil {
		if rem, ok := actor.State.AbilityCooldowns[id]; ok && rem > 0 {
			if s != "" {
				s += " · "
			}
			s += fmt.Sprintf("ещё %dр", rem)
		}
	}
	return s
}

// AbilityCostSummaryForUI — алиас для совместимости; используйте AbilityCostLinePlayerRU.
func AbilityCostSummaryForUI(ctx *BattleContext, actor *BattleUnit, id AbilityID) string {
	return AbilityCostLinePlayerRU(ctx, actor, id)
}

// AbilityUnavailableHintRU — короткая причина недоступности для UI (текст по коду gate, без изменения правил).
func AbilityUnavailableHintRU(ctx *BattleContext, actor *BattleUnit, id AbilityID) string {
	if ctx == nil || actor == nil {
		return ""
	}
	g := AbilityResourceGate(ctx, actor, id)
	if g.OK {
		return ""
	}
	c := abilityCost(id)
	switch g.Code {
	case ErrAbilityOnCooldown:
		rem := 0
		if actor.State.AbilityCooldowns != nil {
			rem = actor.State.AbilityCooldowns[id]
		}
		return fmt.Sprintf("КД: ещё %d р.", rem)
	case ErrInsufficientMana:
		return fmt.Sprintf("Мана: нужно %d, сейчас %d", c.Mana, actor.State.Mana)
	case ErrInsufficientEnergy:
		return fmt.Sprintf("Энергия: нужно %d, сейчас %d", c.Energy, actor.State.Energy)
	default:
		return g.Message
	}
}
