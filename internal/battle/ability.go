package battle

// TargetRule — правило выбора цели способности.
type TargetRule int

const (
	TargetEnemySingle TargetRule = iota
	TargetSelf
)

// AbilityRange — дальность способности (контакт/дальняя).
type AbilityRange int

const (
	RangeMelee AbilityRange = iota
	RangeRanged
)

// AbilityID идентифицирует способность.
type AbilityID int

const (
	AbilityBasicAttack AbilityID = iota
	AbilityRangedAttack
)

// Ability описывает способность (тип, имя, правило целей, дальность).
type Ability struct {
	ID         AbilityID
	Name       string
	TargetRule TargetRule
	Range     AbilityRange
}

var abilityRegistry = map[AbilityID]Ability{
	AbilityBasicAttack: {
		ID:         AbilityBasicAttack,
		Name:       "Attack",
		TargetRule: TargetEnemySingle,
		Range:      RangeMelee,
	},
	AbilityRangedAttack: {
		ID:         AbilityRangedAttack,
		Name:       "Shoot",
		TargetRule: TargetEnemySingle,
		Range:      RangeRanged,
	},
}

// GetAbility возвращает способность по ID.
func GetAbility(id AbilityID) Ability {
	return abilityRegistry[id]
}
