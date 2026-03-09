package entity

// EntityID уникально идентифицирует сущность в мире.
type EntityID int

// EntityType — тип сущности (враг и т.д.).
type EntityType int

const (
	EntityEnemy EntityType = iota
)

// Entity — сущность в мире (враг, NPC и т.д.).
type Entity struct {
	ID    EntityID
	Type  EntityType
	X     int
	Y     int
	Alive bool
}

// EntitySpawnKey — ключ для отслеживания побеждённых спавнов врагов.
type EntitySpawnKey struct {
	Type EntityType
	X    int
	Y    int
}
