package battle

// TargetRule — правило выбора цели способности.
type TargetRule int

const (
	TargetEnemySingle TargetRule = iota
	TargetAllySingle
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
	AbilityHeal
	AbilityBuff
)

// Ability описывает способность (ID, имя, дальность, правило целей).
type Ability struct {
	ID         AbilityID
	Name       string
	Range      AbilityRange
	TargetRule TargetRule
}

// abilityRegistry — реестр способностей.
var abilityRegistry = map[AbilityID]Ability{
	AbilityBasicAttack: {
		ID:         AbilityBasicAttack,
		Name:       "Attack",
		Range:      RangeMelee,
		TargetRule: TargetEnemySingle,
	},
	AbilityRangedAttack: {
		ID:         AbilityRangedAttack,
		Name:       "Shoot",
		Range:      RangeRanged,
		TargetRule: TargetEnemySingle,
	},
	AbilityHeal: {
		ID:         AbilityHeal,
		Name:       "Heal",
		Range:      RangeMelee,
		TargetRule: TargetAllySingle,
	},
	AbilityBuff: {
		ID:         AbilityBuff,
		Name:       "Buff",
		Range:      RangeMelee,
		TargetRule: TargetAllySingle,
	},
}

// GetAbility возвращает способность по ID.
func GetAbility(id AbilityID) Ability {
	return abilityRegistry[id]
}

// Role — боевая роль юнита (набор способностей).
type Role int

const (
	RoleFighter Role = iota
	RoleArcher
	RoleHealer
	RoleMage
)

// GetRoleAbilities возвращает способности роли.
func GetRoleAbilities(role Role) []AbilityID {
	switch role {
	case RoleFighter:
		return []AbilityID{AbilityBasicAttack}
	case RoleArcher:
		return []AbilityID{AbilityRangedAttack}
	case RoleHealer:
		return []AbilityID{AbilityHeal, AbilityBasicAttack}
	case RoleMage:
		return []AbilityID{AbilityBuff, AbilityBasicAttack}
	default:
		return []AbilityID{AbilityBasicAttack}
	}
}

// FilterBasicAttack возвращает срез способностей без базовой атаки (для списка «только специальные»).
func FilterBasicAttack(abils []AbilityID) []AbilityID {
	out := make([]AbilityID, 0, len(abils))
	for _, id := range abils {
		if id != AbilityBasicAttack {
			out = append(out, id)
		}
	}
	return out
}

// SpecialAbilities возвращает способности юнита без базовой атаки (только специальные: heal, buff и т.д.).
// Используется для панели способностей: базовая атака = действие по умолчанию (клик по врагу).
func SpecialAbilities(u *BattleUnit) []AbilityID {
	if u == nil {
		return nil
	}
	return FilterBasicAttack(u.Abilities())
}

// HasBasicAttack возвращает true, если у юнита есть базовая атака (для default attack mode).
func HasBasicAttack(u *BattleUnit) bool {
	if u == nil {
		return false
	}
	for _, id := range u.Abilities() {
		if id == AbilityBasicAttack {
			return true
		}
	}
	return false
}
