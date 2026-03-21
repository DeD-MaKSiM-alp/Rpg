package battle

// BattleAction — действие, выполняемое юнитом (Actor использует Ability по Target).
type BattleAction struct {
	Actor   UnitID
	Ability AbilityID
	Target  UnitID
}

// HealApplication — одно применение лечения (для группового лечения несколько записей).
type HealApplication struct {
	Target UnitID
	Amount int
}

// ActionResult — описание фактического эффекта действия (runtime применяет его).
type ActionResult struct {
	Actor      UnitID
	Target     UnitID
	Damage     int
	Killed     bool
	HealAmount int // если > 0 — лечение одной цели (когда HealApplications пусто)
	// HealApplications — если не пусто, лечение по нескольким целям (напр. массовый хил).
	HealApplications []HealApplication
}
