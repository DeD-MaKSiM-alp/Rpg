package battle

// BuildEnemyAction выбирает способность, получает допустимые цели и создаёт BattleAction.
func BuildEnemyAction(ctx *BattleContext, actor *BattleUnit) (BattleAction, bool) {
	if actor == nil || len(actor.Abilities()) == 0 {
		return BattleAction{}, false
	}
	for _, abilityID := range ctx.AvailableAbilities(actor) {
		validTargets, _ := ListValidTargets(ctx, actor.ID, abilityID)
		if len(validTargets) == 0 {
			continue
		}
		req := ActionRequest{Actor: actor.ID, Ability: abilityID, Target: validTargets[0]}
		act, v := ToBattleAction(ctx, req)
		if v.OK {
			return act, true
		}
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
	if len(ctx.AvailableAbilities(actor)) == 0 {
		return BattleAction{}, false
	}
	for _, abilityID := range ctx.AvailableAbilities(actor) {
		validTargets, _ := ListValidTargets(ctx, actor.ID, abilityID)
		if len(validTargets) == 0 {
			continue
		}
		req := ActionRequest{Actor: actor.ID, Ability: abilityID, Target: validTargets[0]}
		act, v := ToBattleAction(ctx, req)
		if v.OK {
			return act, true
		}
	}
	return BattleAction{}, false
}
