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

// BattleUnit — runtime-сущность боя (не world entity, не template).
type BattleUnit struct {
	ID          UnitID
	Name        string
	Team        TeamID
	Slot        int
	Row         RowType
	MaxHP       int
	HP          int
	Attack         int
	Defense        int
	Initiative     int
	Alive          bool
	Ranged         bool // юнит-архетип: дальняя атака (переопределяет ability.Range для targeting)
	Abilities      []AbilityID
	AttackModifier int // бонус к атаке от Buff (до конца боя)
	SourceEnemyID entity.EntityID // только для врагов
}

// BattleTeam — сторона боя.
type BattleTeam struct {
	ID    TeamID
	Units []UnitID
}

// IsAlive возвращает true, если юнит жив.
func (u *BattleUnit) IsAlive() bool {
	return u != nil && u.Alive && u.HP > 0
}
