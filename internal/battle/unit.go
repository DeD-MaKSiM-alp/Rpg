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
	ID     UnitID
	Name   string
	Team   TeamID
	Slot   int
	MaxHP  int
	HP     int
	Attack int
	Defense int
	Initiative int
	Alive  bool
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

// ApplyDamage наносит урон, возвращает фактический урон.
func (u *BattleUnit) ApplyDamage(amount int) int {
	if u == nil || amount <= 0 {
		return 0
	}
	actual := amount - u.Defense
	if actual < 1 {
		actual = 1
	}
	u.HP -= actual
	if u.HP < 0 {
		u.HP = 0
	}
	if u.HP <= 0 {
		u.Alive = false
	}
	return actual
}

// Kill помечает юнита мёртвым.
func (u *BattleUnit) Kill() {
	if u != nil {
		u.HP = 0
		u.Alive = false
	}
}
