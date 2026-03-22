// Package unitdata holds minimal in-code unit templates (Empire early game) for hero identity and creation.
// Runtime combat state stays in hero.Hero; battle seeds are still built from hero stats (see hero.CombatUnitSeed).
package unitdata

import (
	"fmt"

	battlepkg "mygame/internal/battle"
)

// Stable template IDs (design-doc aligned where possible).
const (
	EmpireMilitiaSpearmanT1 = "empire_militia_spearman_t1"
	EmpireWarriorRecruit    = "empire_warrior_recruit"
	EmpireWarriorSquire     = "empire_warrior_squire"
	EmpireArcherRecruit     = "empire_archer_recruit"
	EmpireArcherMarksmanBase = "empire_archer_marksman_base"
	EmpireHealerNovice      = "empire_healer_novice"
	EmpireHealerAcolyte     = "empire_healer_acolyte"
	// Tier 3 — первые формы по линиям (design-doc: empire_warrior_dd_1, empire_archer_pure_1, empire_healer_single_1).
	EmpireWarriorDD1    = "empire_warrior_dd_1"
	EmpireWarriorTank1  = "empire_warrior_tank_1"
	EmpireArcherPure1   = "empire_archer_pure_1"
	EmpireHealerSingle1 = "empire_healer_single_1"
	EmpireHealerGroup1  = "empire_healer_group_1"
)

// AttackKind mirrors design-doc attack_type for inspect / data layer (not the full combat rules engine).
type AttackKind int

const (
	AttackMelee AttackKind = iota
	AttackRanged
	AttackHeal
)

// UnitTemplate — identity + starting stat profile for a recruitable/playable archetype.
type UnitTemplate struct {
	UnitID      string
	DisplayName string
	FactionID   string
	LineID      string
	Tier        int
	// ArchetypeID — короткий машинный ключ (как в design-doc), для inspect.
	ArchetypeID string
	Role        battlepkg.Role
	AttackKind  AttackKind

	MaxHP     int
	Attack    int
	Defense   int
	Initiative int
	HealPower int

	Abilities []battlepkg.AbilityID
	// InspectNote — одна строка для карточки бойца (опционально).
	InspectNote string
	// UpgradeToUnitID — один следующий шаг (линейный путь), если UpgradeOptions пуст.
	UpgradeToUnitID string
	// UpgradeOptions — 2+ целей ветвления (tier 2→3); если не пусто, UpgradeToUnitID игнорируется.
	UpgradeOptions []string
}

var unitRegistry = map[string]UnitTemplate{
	EmpireMilitiaSpearmanT1: {
		UnitID:      EmpireMilitiaSpearmanT1,
		DisplayName: "Милиция · копейщик",
		FactionID:   "empire",
		LineID:      "warrior",
		Tier:        1,
		ArchetypeID: "melee_generalist",
		Role:        battlepkg.RoleFighter,
		AttackKind:  AttackMelee,
		MaxHP:       10,
		Attack:      2,
		Defense:     0,
		Initiative:  2,
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityPowerStrike, battlepkg.AbilityBasicAttack},
		InspectNote:     "Стартовый офицер отряда, ближний бой.",
		UpgradeToUnitID: "empire_warrior_squire",
	},
	EmpireWarriorRecruit: {
		UnitID:      EmpireWarriorRecruit,
		DisplayName: "Новобранец",
		FactionID:   "empire",
		LineID:      "warrior",
		Tier:        1,
		ArchetypeID: "melee_generalist",
		Role:        battlepkg.RoleFighter,
		AttackKind:  AttackMelee,
		MaxHP:       9,
		Attack:      2,
		Defense:     0,
		Initiative:  2,
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityPowerStrike, battlepkg.AbilityBasicAttack},
		InspectNote:     "Базовый пехотный рекрут Империи.",
		UpgradeToUnitID: "empire_warrior_squire",
	},
	EmpireArcherRecruit: {
		UnitID:      EmpireArcherRecruit,
		DisplayName: "Рекрут-лучник",
		FactionID:   "empire",
		LineID:      "archer",
		Tier:        1,
		ArchetypeID: "ranged_generalist",
		Role:        battlepkg.RoleArcher,
		AttackKind:  AttackRanged,
		MaxHP:       8,
		Attack:      2,
		Defense:     0,
		Initiative:  3,
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityRangedAttack, battlepkg.AbilityBasicAttack},
		InspectNote:     "Дальний бой; основной выстрел по любой цели.",
		UpgradeToUnitID: "empire_archer_marksman_base",
	},
	EmpireHealerNovice: {
		UnitID:      EmpireHealerNovice,
		DisplayName: "Послушник",
		FactionID:   "empire",
		LineID:      "healer",
		Tier:        1,
		ArchetypeID: "healer_generalist",
		Role:        battlepkg.RoleHealer,
		AttackKind:  AttackHeal,
		MaxHP:       8,
		Attack:      1,
		Defense:     0,
		Initiative:  2,
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityHeal, battlepkg.AbilityBasicAttack},
		InspectNote:     "Поддержка союзников лечением.",
		UpgradeToUnitID: "empire_healer_acolyte",
	},
	// Tier 2 — минимальные статы для promotion flow; дальнейшая ветка — отдельные этапы.
	EmpireWarriorSquire: {
		UnitID:      EmpireWarriorSquire,
		DisplayName: "Оруженосец",
		FactionID:   "empire",
		LineID:      "warrior",
		Tier:        2,
		ArchetypeID: "melee_generalist",
		Role:        battlepkg.RoleFighter,
		AttackKind:  AttackMelee,
		MaxHP:       12,
		Attack:      3,
		Defense:     1,
		Initiative:  2,
		HealPower:    0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityPowerStrike, battlepkg.AbilityBasicAttack},
		InspectNote: "Воинская линия, tier 2.",
		UpgradeOptions: []string{EmpireWarriorTank1, EmpireWarriorDD1},
	},
	EmpireArcherMarksmanBase: {
		UnitID:      EmpireArcherMarksmanBase,
		DisplayName: "Стрелок",
		FactionID:   "empire",
		LineID:      "archer",
		Tier:        2,
		ArchetypeID: "ranged_generalist",
		Role:        battlepkg.RoleArcher,
		AttackKind:  AttackRanged,
		MaxHP:       9,
		Attack:      3,
		Defense:     0,
		Initiative:  4,
		HealPower:    0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityRangedAttack, battlepkg.AbilityBasicAttack},
		InspectNote:     "Стрелковая линия, tier 2.",
		UpgradeToUnitID: EmpireArcherPure1,
	},
	EmpireHealerAcolyte: {
		UnitID:      EmpireHealerAcolyte,
		DisplayName: "Аколит",
		FactionID:   "empire",
		LineID:      "healer",
		Tier:        2,
		ArchetypeID: "healer_generalist",
		Role:        battlepkg.RoleHealer,
		AttackKind:  AttackHeal,
		MaxHP:       9,
		Attack:      1,
		Defense:     0,
		Initiative:  2,
		HealPower:    0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityHeal, battlepkg.AbilityBasicAttack},
		InspectNote: "Линия хилов, tier 2.",
		UpgradeOptions: []string{EmpireHealerSingle1, EmpireHealerGroup1},
	},
	// Tier 3 — те же семейства способностей; у воинов добавлен мощный удар.
	EmpireWarriorDD1: {
		UnitID:      EmpireWarriorDD1,
		DisplayName: "Мечник",
		FactionID:   "empire",
		LineID:      "warrior",
		Tier:        3,
		ArchetypeID: "melee_dd",
		Role:        battlepkg.RoleFighter,
		AttackKind:  AttackMelee,
		MaxHP:       15,
		Attack:      4,
		Defense:     2,
		Initiative:  3,
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityPowerStrike, battlepkg.AbilityBasicAttack},
		InspectNote: "Воинская линия, tier 3 (ДД).",
	},
	EmpireWarriorTank1: {
		UnitID:      EmpireWarriorTank1,
		DisplayName: "Щитоносец",
		FactionID:   "empire",
		LineID:      "warrior",
		Tier:        3,
		ArchetypeID: "tank",
		Role:        battlepkg.RoleFighter,
		AttackKind:  AttackMelee,
		MaxHP:       17,
		Attack:      2,
		Defense:     4,
		Initiative:  1,
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityPowerStrike, battlepkg.AbilityBasicAttack},
		InspectNote: "Воинская линия, tier 3 (танк).",
	},
	EmpireArcherPure1: {
		UnitID:      EmpireArcherPure1,
		DisplayName: "Лучник",
		FactionID:   "empire",
		LineID:      "archer",
		Tier:        3,
		ArchetypeID: "ranged_dd",
		Role:        battlepkg.RoleArcher,
		AttackKind:  AttackRanged,
		MaxHP:       10,
		Attack:      4,
		Defense:     0,
		Initiative:  5,
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityRangedAttack, battlepkg.AbilityBasicAttack},
		InspectNote: "Стрелковая линия, tier 3 (чистый урон).",
	},
	EmpireHealerSingle1: {
		UnitID:      EmpireHealerSingle1,
		DisplayName: "Целитель",
		FactionID:   "empire",
		LineID:      "healer",
		Tier:        3,
		ArchetypeID: "single_healer",
		Role:        battlepkg.RoleHealer,
		AttackKind:  AttackHeal,
		MaxHP:       10,
		Attack:      1,
		Defense:     0,
		Initiative:  2,
		HealPower:   1,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityHeal, battlepkg.AbilityBasicAttack},
		InspectNote: "Линия хилов, tier 3 (сильное одиночное лечение).",
	},
	EmpireHealerGroup1: {
		UnitID:      EmpireHealerGroup1,
		DisplayName: "Пастырь",
		FactionID:   "empire",
		LineID:      "healer",
		Tier:        3,
		ArchetypeID: "group_healer",
		Role:        battlepkg.RoleHealer,
		AttackKind:  AttackHeal,
		MaxHP:       11,
		Attack:      1,
		Defense:     0,
		Initiative:  2,
		// Бонус к GroupHealPower (1+bonus на союзника); слабее по цели, чем одиночное Heal (2+bonus).
		HealPower:   0,
		Abilities:   []battlepkg.AbilityID{battlepkg.AbilityGroupHeal, battlepkg.AbilityBasicAttack},
		InspectNote: "Линия хилов, tier 3: массовое лечение союзников (по чуть-чуть каждому).",
	},
}

// EarlyRecruitUnitIDs — фиксированный пул раннего найма (циклический выбор).
func EarlyRecruitUnitIDs() []string {
	return []string{
		EmpireWarriorRecruit,
		EmpireArcherRecruit,
		EmpireHealerNovice,
	}
}

// GetUnitTemplate returns a copy of the template if id is registered.
func GetUnitTemplate(id string) (UnitTemplate, bool) {
	if id == "" {
		return UnitTemplate{}, false
	}
	t, ok := unitRegistry[id]
	if !ok {
		return UnitTemplate{}, false
	}
	return t, true
}

// ErrUnknownUnit — неизвестный unit_id (для фабрик героя).
type ErrUnknownUnit struct {
	UnitID string
}

func (e ErrUnknownUnit) Error() string {
	return fmt.Sprintf("unitdata: unknown unit_id %q", e.UnitID)
}

// MustGetUnitTemplate returns template or error-like panic is avoided — use Get in gameplay code.
func MustGetUnitTemplate(id string) (UnitTemplate, error) {
	t, ok := GetUnitTemplate(id)
	if !ok {
		return UnitTemplate{}, ErrUnknownUnit{UnitID: id}
	}
	return t, nil
}

// FactionDisplayRU — короткая подпись для UI.
func FactionDisplayRU(factionID string) string {
	switch factionID {
	case "empire":
		return "Империя"
	default:
		if factionID == "" {
			return "—"
		}
		return factionID
	}
}

// LineDisplayRU — линия развития.
func LineDisplayRU(lineID string) string {
	switch lineID {
	case "warrior":
		return "воинская"
	case "archer":
		return "стрелковая"
	case "healer":
		return "линия хилов"
	default:
		if lineID == "" {
			return "—"
		}
		return lineID
	}
}

// AttackKindDisplayRU — тип атаки для карточки.
func AttackKindDisplayRU(k AttackKind) string {
	switch k {
	case AttackMelee:
		return "ближний бой"
	case AttackRanged:
		return "дальний бой"
	case AttackHeal:
		return "лечение"
	default:
		return "—"
	}
}
