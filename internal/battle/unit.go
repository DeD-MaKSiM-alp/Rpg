package battle

import (
	"mygame/world/entity"
)

// UnitID уникально идентифицирует боевого юнита в рамках боя.
type UnitID int

// TeamID — сторона боя.
type TeamID int

const (
	TeamPlayer TeamID = iota
	TeamEnemy
)

// UnitSide — доменное имя для стороны боя (алиас на TeamID для совместимости).
type UnitSide = TeamID

// UnitRole — доменное имя роли (алиас на Role).
type UnitRole = Role

// UnitBaseStats — базовые статы архетипа/шаблона (definition layer).
type UnitBaseStats struct {
	MaxHP      int
	Attack     int
	Defense    int
	Initiative int
}

// AbilityLoadout — базовый набор способностей юнита (definition layer).
type AbilityLoadout struct {
	Abilities []AbilityID
}

// CombatUnitDefinition — archetype/template layer.
// Не содержит battle runtime (HP/Alive), не содержит world identity.
type CombatUnitDefinition struct {
	ArchetypeID string // стабильный ID архетипа (для прогрессии/контента)
	DisplayName string
	Role        UnitRole
	Base        UnitBaseStats

	// Profile: базовые свойства архетипа, влияющие на таргетинг/правила.
	IsRanged bool

	Loadout AbilityLoadout
}

// StatusInstance — groundwork для статусов (runtime layer).
// Пока не используется, но задаёт место для будущей нормализации.
type StatusInstance struct {
	ID       string
	Stacks   int
	Duration int // в ходах/раундах (уточнится позже)
}

// CombatModifiers — временные модификаторы боя (runtime layer).
type CombatModifiers struct {
	AttackBonus  int
	DefenseBonus int
	InitiativeBonus int
}

// CombatUnitState — runtime layer (состояние конкретного экземпляра в бою).
type CombatUnitState struct {
	HP       int
	Alive    bool
	Disabled bool // groundwork: stun/sleep/etc.

	Modifiers CombatModifiers
	Statuses  []StatusInstance

	// Placement groundwork (до следующего шага side+slot model).
	Slot int
	Row  RowType
}

// CombatUnitOrigin — hook для связи с внешним миром/персистентностью.
// Не смешивается с definition/runtime: используется только для интеграции (награды, удаление world-enemy).
type CombatUnitOrigin struct {
	WorldEnemyID entity.EntityID // для enemy units; 0 для player units
}

// CombatUnit — каноническая runtime сущность боя.
// Содержит ссылку на definition (archetype) + runtime state + integration hooks.
type CombatUnit struct {
	ID   UnitID
	Side UnitSide

	Def   CombatUnitDefinition
	State CombatUnitState
	Origin CombatUnitOrigin
}

// BattleUnit оставляем как алиас для минимальных изменений существующего battle-кода.
// Новый код должен использовать термин "CombatUnit" как доменный.
type BattleUnit = CombatUnit

// BattleTeam — сторона боя.
type BattleTeam struct {
	ID    TeamID
	Units []UnitID
}

// IsAlive возвращает true, если юнит жив.
func (u *CombatUnit) IsAlive() bool {
	return u != nil && u.State.Alive && u.State.HP > 0
}

func (u *CombatUnit) Name() string { return u.Def.DisplayName }
func (u *CombatUnit) MaxHP() int   { return u.Def.Base.MaxHP }
func (u *CombatUnit) Attack() int {
	return u.Def.Base.Attack + u.State.Modifiers.AttackBonus
}
func (u *CombatUnit) Defense() int {
	return u.Def.Base.Defense + u.State.Modifiers.DefenseBonus
}
func (u *CombatUnit) Initiative() int {
	return u.Def.Base.Initiative + u.State.Modifiers.InitiativeBonus
}
func (u *CombatUnit) IsRanged() bool { return u.Def.IsRanged }
func (u *CombatUnit) Abilities() []AbilityID {
	return u.Def.Loadout.Abilities
}
