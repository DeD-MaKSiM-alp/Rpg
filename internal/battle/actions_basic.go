package battle

// BuildEnemyAction выбирает способность, получает допустимые цели и создаёт BattleAction.
func BuildEnemyAction(ctx *BattleContext, actor *BattleUnit) (BattleAction, bool) {
	if actor == nil || len(actor.Abilities()) == 0 {
		return BattleAction{}, false
	}
	for _, abilityID := range actor.Abilities() {
		ability := GetAbility(abilityID)
		targets := ctx.ReachableTargets(actor, ability)
		if len(targets) == 0 {
			continue
		}
		return BattleAction{
			Actor:   actor.ID,
			Ability: abilityID,
			Target:  targets[0].ID,
		}, true
	}
	return BattleAction{}, false
}

// BuildFirstAvailablePlayerAction строит первое доступное действие игрока (первая способность с допустимой целью).
func BuildFirstAvailablePlayerAction(ctx *BattleContext, actor *BattleUnit) (BattleAction, bool) {
	return buildPlayerAction(ctx, actor)
}

func buildPlayerAction(ctx *BattleContext, actor *BattleUnit) (BattleAction, bool) {
	if actor == nil || actor.Side != TeamPlayer {
		return BattleAction{}, false
	}
	if len(actor.Abilities()) == 0 {
		return BattleAction{}, false
	}
	for _, abilityID := range actor.Abilities() {
		ability := GetAbility(abilityID)
		targets := ctx.ReachableTargets(actor, ability)
		if len(targets) == 0 {
			continue
		}
		return BattleAction{
			Actor:   actor.ID,
			Ability: abilityID,
			Target:  targets[0].ID,
		}, true
	}
	return BattleAction{}, false
}
