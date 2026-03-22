package battle

// TargetRule — правило выбора цели способности.
type TargetRule int

const (
	TargetEnemySingle TargetRule = iota
	TargetAllySingle
	TargetSelf
	// TargetAllyTeam — без выбора юнита; эффект на всю живую союзную сторону (напр. массовое лечение).
	TargetAllyTeam
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
	AbilityGroupHeal
	// AbilityPowerStrike — ударник: сильный ближний удар за энергию и КД (не глобальный ресурс).
	AbilityPowerStrike
)

// Ability описывает способность (ID, имя, дальность, правило целей, стоимость/КД).
type Ability struct {
	ID         AbilityID
	Name       string
	Range      AbilityRange
	TargetRule TargetRule
	// CostMana / CostEnergy — расход при использовании (0 = нет).
	// CooldownRounds — сколько полных раундов перезарядки после применения (0 = без КД).
	CostMana         int
	CostEnergy       int
	CooldownRounds   int
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
		ID:               AbilityRangedAttack,
		Name:             "Shoot",
		Range:            RangeRanged,
		TargetRule:       TargetEnemySingle,
		CostEnergy:       2,
		CooldownRounds:   0,
	},
	AbilityHeal: {
		ID:               AbilityHeal,
		Name:             "Heal",
		Range:            RangeMelee,
		TargetRule:       TargetAllySingle,
		CostMana:         6,
		CooldownRounds:   4,
	},
	AbilityGroupHeal: {
		ID:               AbilityGroupHeal,
		Name:             "Массовое лечение",
		Range:            RangeMelee,
		TargetRule:       TargetAllyTeam,
		CostMana:         9,
		CooldownRounds:   5,
	},
	AbilityBuff: {
		ID:               AbilityBuff,
		Name:             "Buff",
		Range:            RangeMelee,
		TargetRule:       TargetAllySingle,
		CostMana:         4,
		CooldownRounds:   3,
	},
	AbilityPowerStrike: {
		ID:               AbilityPowerStrike,
		Name:             "PowerStrike",
		Range:            RangeMelee,
		TargetRule:       TargetEnemySingle,
		CostEnergy:       4,
		CooldownRounds:   3,
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
		// Ударник: мощный удар на КД + базовый удар как основной ритм.
		return []AbilityID{AbilityPowerStrike, AbilityBasicAttack}
	case RoleArcher:
		// Базовый удар остаётся бесплатным запасным вариантом при нулевой энергии.
		return []AbilityID{AbilityRangedAttack, AbilityBasicAttack}
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
