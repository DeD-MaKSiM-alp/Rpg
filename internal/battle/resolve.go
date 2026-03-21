package battle

import "fmt"

const buffAttackBonus = 1

// ResolveAbility применяет действие по способности и возвращает результат. Единственная точка изменения HP/статов.
func ResolveAbility(ctx *BattleContext, action BattleAction) ActionResult {
	if ctx == nil || ctx.Units == nil {
		return ActionResult{}
	}
	ability := GetAbility(action.Ability)
	actor := ctx.Units[action.Actor]
	if actor == nil || !actor.IsAlive() {
		return ActionResult{}
	}

	// Unified rule gate: never assume caller validated correctly.
	req := ActionRequest{Actor: action.Actor, Ability: action.Ability, Target: UnitTarget(action.Target)}
	if ability.TargetRule == TargetSelf {
		req.Target = SelfTarget()
	}
	if action.Target == 0 {
		req.Target = NoTarget()
	}
	if v := ValidateAction(ctx, req); !v.OK {
		return ActionResult{}
	}
	target := ctx.Units[action.Target]

	switch ability.ID {
	case AbilityBasicAttack, AbilityRangedAttack:
		if target == nil || !target.IsAlive() {
			return ActionResult{}
		}
		atk := actor.Attack()
		damage := atk - target.Defense()
		if damage < 1 {
			damage = 1
		}
		target.State.HP -= damage
		killed := false
		if target.State.HP <= 0 {
			target.State.HP = 0
			target.State.Alive = false
			killed = true
		}
		ctx.AddBattleLog(fmt.Sprintf("%s hits %s for %d.", actor.Name(), target.Name(), damage))
		if killed {
			ctx.AddBattleLog(fmt.Sprintf("%s dies.", target.Name()))
		}
		return ActionResult{
			Actor:  action.Actor,
			Target: action.Target,
			Damage: damage,
			Killed: killed,
		}
	case AbilityHeal:
		if target == nil || !target.IsAlive() {
			return ActionResult{}
		}
		amount := actor.HealPower()
		target.State.HP += amount
		if target.State.HP > target.MaxHP() {
			target.State.HP = target.MaxHP()
		}
		ctx.AddBattleLog(fmt.Sprintf("%s heals %s for %d.", actor.Name(), target.Name(), amount))
		return ActionResult{
			Actor:      action.Actor,
			Target:     action.Target,
			HealAmount: amount,
		}
	case AbilityBuff:
		if target == nil || !target.IsAlive() {
			return ActionResult{}
		}
		target.State.Modifiers.AttackBonus += buffAttackBonus
		ctx.AddBattleLog(fmt.Sprintf("%s buffs %s.", actor.Name(), target.Name()))
		return ActionResult{
			Actor:  action.Actor,
			Target: action.Target,
		}
	default:
		return ActionResult{}
	}
}
