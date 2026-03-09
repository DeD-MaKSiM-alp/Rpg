package battle

// RowType — ряд построения (передний/задний).
type RowType int

const (
	RowFront RowType = iota
	RowBack
)

// MaxFrontRowUnits — максимум юнитов в переднем ряду.
const MaxFrontRowUnits = 3

// LivingUnitsInRow возвращает живых юнитов команды в указанном ряду.
func (c *BattleContext) LivingUnitsInRow(team TeamID, row RowType) []*BattleUnit {
	t := c.Teams[team]
	if t == nil {
		return nil
	}
	var out []*BattleUnit
	for _, id := range t.Units {
		u := c.Units[id]
		if u != nil && u.IsAlive() && u.Row == row {
			out = append(out, u)
		}
	}
	return out
}

// FrontRowAlive возвращает true, если в переднем ряду команды есть живые юниты.
func (c *BattleContext) FrontRowAlive(team TeamID) bool {
	return len(c.LivingUnitsInRow(team, RowFront)) > 0
}

// BackRowAlive возвращает true, если в заднем ряду команды есть живые юниты.
func (c *BattleContext) BackRowAlive(team TeamID) bool {
	return len(c.LivingUnitsInRow(team, RowBack)) > 0
}

// effectiveRange возвращает эффективную дальность способности для актёра (учёт Ranged).
func effectiveRange(actor *BattleUnit, ability Ability) AbilityRange {
	if actor != nil && actor.Ranged {
		return RangeRanged
	}
	return ability.Range
}

// ReachableEnemyTargets возвращает допустимые цели врага для способности по правилам formation.
func (c *BattleContext) ReachableEnemyTargets(actor *BattleUnit, ability Ability) []*BattleUnit {
	if actor == nil {
		return nil
	}
	enemyTeam := c.EnemyTeam(actor.Team)
	allEnemies := c.LivingUnits(enemyTeam)
	if len(allEnemies) == 0 {
		return nil
	}
	if ability.TargetRule != TargetEnemySingle {
		return nil
	}
	rng := effectiveRange(actor, ability)
	if rng == RangeRanged {
		return allEnemies
	}
	if c.FrontRowAlive(enemyTeam) {
		return c.LivingUnitsInRow(enemyTeam, RowFront)
	}
	return c.LivingUnitsInRow(enemyTeam, RowBack)
}

// ReachableAllyTargets возвращает допустимые цели союзника (живые юниты своей команды, включая себя).
func (c *BattleContext) ReachableAllyTargets(actor *BattleUnit, ability Ability) []*BattleUnit {
	if actor == nil {
		return nil
	}
	if ability.TargetRule != TargetAllySingle {
		return nil
	}
	return c.LivingUnits(actor.Team)
}

// ReachableTargets возвращает допустимые цели способности для актёра (враги, союзники или сам актёр).
func (c *BattleContext) ReachableTargets(actor *BattleUnit, ability Ability) []*BattleUnit {
	switch ability.TargetRule {
	case TargetEnemySingle:
		return c.ReachableEnemyTargets(actor, ability)
	case TargetAllySingle:
		return c.ReachableAllyTargets(actor, ability)
	case TargetSelf:
		if actor != nil && actor.IsAlive() {
			return []*BattleUnit{actor}
		}
		return nil
	default:
		return nil
	}
}

// CanTarget проверяет, может ли актёр выбрать цель данной способностью.
func (c *BattleContext) CanTarget(actor *BattleUnit, ability Ability, target *BattleUnit) bool {
	if actor == nil || target == nil {
		return false
	}
	reachable := c.ReachableTargets(actor, ability)
	for _, u := range reachable {
		if u.ID == target.ID {
			return true
		}
	}
	return false
}
