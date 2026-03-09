package battle

// BattleAction — действие, выполняемое юнитом (Actor использует Ability по Target).
type BattleAction struct {
	Actor   UnitID
	Ability AbilityID
	Target  UnitID
}

// ActionResult — описание фактического эффекта действия (runtime применяет его).
type ActionResult struct {
	Actor  UnitID
	Target UnitID
	Damage int
	Killed bool
}
