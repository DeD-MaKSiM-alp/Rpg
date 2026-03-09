package battle

// ResolveAbility применяет действие по способности и возвращает результат. Единственная точка изменения HP.
func ResolveAbility(ctx *BattleContext, action BattleAction) ActionResult {
	ability := GetAbility(action.Ability)
	switch ability.ID {
	case AbilityBasicAttack, AbilityRangedAttack:
		attacker := ctx.Units[action.Actor]
		target := ctx.Units[action.Target]
		if attacker == nil || target == nil {
			return ActionResult{}
		}
		if !attacker.IsAlive() || !target.IsAlive() {
			return ActionResult{}
		}
		damage := attacker.Attack - target.Defense
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
	default:
		return ActionResult{}
	}
}
