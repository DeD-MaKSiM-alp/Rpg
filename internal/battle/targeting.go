package battle

// FirstLivingUnit возвращает первого живого юнита команды.
func (c *BattleContext) FirstLivingUnit(team TeamID) *BattleUnit {
	live := c.LivingUnits(team)
	if len(live) == 0 {
		return nil
	}
	return live[0]
}

// EnemyTeam возвращает противоположную команду.
func (c *BattleContext) EnemyTeam(team TeamID) TeamID {
	if team == TeamPlayer {
		return TeamEnemy
	}
	return TeamPlayer
}

// AllyTargets возвращает живых юнитов команды актёра (включая себя).
func (c *BattleContext) AllyTargets(actor *BattleUnit) []*BattleUnit {
	if actor == nil {
		return nil
	}
	return c.LivingUnits(actor.Side)
}
