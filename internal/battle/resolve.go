package battle

const healAmount = 2
const buffAttackBonus = 1

// ResolveAbility применяет действие по способности и возвращает результат. Единственная точка изменения HP/статов.
func ResolveAbility(ctx *BattleContext, action BattleAction) ActionResult {
	ability := GetAbility(action.Ability)
	actor := ctx.Units[action.Actor]
	target := ctx.Units[action.Target]
	if actor == nil || !actor.IsAlive() {
		return ActionResult{}
	}

	switch ability.ID {
	case AbilityBasicAttack, AbilityRangedAttack:
		if target == nil || !target.IsAlive() {
			return ActionResult{}
		}
		atk := actor.Attack + actor.AttackModifier
		damage := atk - target.Defense
		if damage < 1 {
			damage = 1
		}
		target.HP -= damage
		killed := false
		if target.HP <= 0 {
			target.HP = 0
			target.Alive = false
			killed = true
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
		amount := healAmount
		target.HP += amount
		if target.HP > target.MaxHP {
			target.HP = target.MaxHP
		}
		return ActionResult{
			Actor:      action.Actor,
			Target:     action.Target,
			HealAmount: amount,
		}
	case AbilityBuff:
		if target == nil || !target.IsAlive() {
			return ActionResult{}
		}
		target.AttackModifier += buffAttackBonus
		return ActionResult{
			Actor:  action.Actor,
			Target: action.Target,
		}
	default:
		return ActionResult{}
	}
}
