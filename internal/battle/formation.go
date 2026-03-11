package battle

// NOTE: legacy reachable helpers are kept as thin wrappers around the unified validation layer.

// LivingUnitsInRow возвращает живых юнитов команды в указанном ряду.
func (c *BattleContext) LivingUnitsInRow(team TeamID, row RowType) []*BattleUnit {
	st := c.SideState(team)
	if st == nil {
		return nil
	}
	var out []*BattleUnit
	for i := range st.Slots {
		sl := &st.Slots[i]
		if sl.ID.Row != row || sl.IsEmpty() {
			continue
		}
		u := c.Units[sl.Occupied]
		if u != nil && u.IsAlive() {
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
	if actor != nil && actor.IsRanged() {
		return RangeRanged
	}
	return ability.Range
}

// ReachableEnemyTargets возвращает допустимые цели врага для способности по правилам formation.
func (c *BattleContext) ReachableEnemyTargets(actor *BattleUnit, ability Ability) []*BattleUnit {
	if c == nil || actor == nil {
		return nil
	}
	if ability.TargetRule != TargetEnemySingle {
		return nil
	}
	tds, _ := ListValidTargets(c, actor.ID, ability.ID)
	out := make([]*BattleUnit, 0, len(tds))
	for _, td := range tds {
		if td.Kind != TargetKindUnit {
			continue
		}
		if u := c.Units[td.UnitID]; u != nil {
			out = append(out, u)
		}
	}
	return out
}

// ReachableAllyTargets возвращает допустимые цели союзника (живые юниты своей команды, включая себя).
func (c *BattleContext) ReachableAllyTargets(actor *BattleUnit, ability Ability) []*BattleUnit {
	if c == nil || actor == nil {
		return nil
	}
	if ability.TargetRule != TargetAllySingle {
		return nil
	}
	tds, _ := ListValidTargets(c, actor.ID, ability.ID)
	out := make([]*BattleUnit, 0, len(tds))
	for _, td := range tds {
		if td.Kind != TargetKindUnit {
			continue
		}
		if u := c.Units[td.UnitID]; u != nil {
			out = append(out, u)
		}
	}
	return out
}

// ReachableTargets возвращает допустимые цели способности для актёра (враги, союзники или сам актёр).
func (c *BattleContext) ReachableTargets(actor *BattleUnit, ability Ability) []*BattleUnit {
	if c == nil || actor == nil {
		return nil
	}
	tds, _ := ListValidTargets(c, actor.ID, ability.ID)
	out := make([]*BattleUnit, 0, len(tds))
	for _, td := range tds {
		switch td.Kind {
		case TargetKindSelf:
			out = append(out, actor)
		case TargetKindUnit:
			if u := c.Units[td.UnitID]; u != nil {
				out = append(out, u)
			}
		}
	}
	return out
}

// CanTarget проверяет, может ли актёр выбрать цель данной способностью.
func (c *BattleContext) CanTarget(actor *BattleUnit, ability Ability, target *BattleUnit) bool {
	if c == nil || actor == nil || target == nil {
		return false
	}
	v := ValidateAction(c, ActionRequest{
		Actor:   actor.ID,
		Ability: ability.ID,
		Target:  UnitTarget(target.ID),
	})
	return v.OK
}
