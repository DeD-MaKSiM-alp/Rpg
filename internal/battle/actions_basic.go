package battle

// BuildEnemyAction строит действие врага: первая способность, цель через ReachableEnemyTargets.
func BuildEnemyAction(ctx *BattleContext, actor *BattleUnit) (BattleAction, bool) {
	if actor == nil || len(actor.Abilities) == 0 {
		return BattleAction{}, false
	}
	abilityID := actor.Abilities[0]
	ability := GetAbility(abilityID)
	switch ability.TargetRule {
	case TargetEnemySingle:
		targets := ctx.ReachableEnemyTargets(actor, ability)
		if len(targets) == 0 {
			return BattleAction{}, false
		}
		return BattleAction{
			Actor:   actor.ID,
			Ability: abilityID,
			Target:  targets[0].ID,
		}, true
	case TargetSelf:
		return BattleAction{
			Actor:   actor.ID,
			Ability: abilityID,
			Target:  actor.ID,
		}, true
	default:
		return BattleAction{}, false
	}
}

// BuildFirstAvailablePlayerAction строит первое доступное действие игрока (первая способность, первая допустимая цель).
func BuildFirstAvailablePlayerAction(ctx *BattleContext, actor *BattleUnit) (BattleAction, bool) {
	return buildPlayerAction(ctx, actor)
}

func buildPlayerAction(ctx *BattleContext, actor *BattleUnit) (BattleAction, bool) {
	if actor == nil || actor.Team != TeamPlayer {
		return BattleAction{}, false
	}
	if len(actor.Abilities) == 0 {
		return BattleAction{}, false
	}
	abilityID := actor.Abilities[0]
	ability := GetAbility(abilityID)
	if ability.TargetRule != TargetEnemySingle {
		return BattleAction{}, false
	}
	targets := ctx.ReachableEnemyTargets(actor, ability)
	if len(targets) == 0 {
		return BattleAction{}, false
	}
	return BattleAction{
		Actor:   actor.ID,
		Ability: abilityID,
		Target:  targets[0].ID,
	}, true
}
