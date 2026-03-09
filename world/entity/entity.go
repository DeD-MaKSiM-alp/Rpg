package entity

// EntityID уникально идентифицирует сущность в мире.
type EntityID int

// EntityType — тип сущности (враг и т.д.).
type EntityType int

const (
	EntityEnemy EntityType = iota
)

// EnemyKind — вид врага в мире (не боевой юнит, только тип сущности).
type EnemyKind int

const (
	EnemyKindSlime EnemyKind = iota
	EnemyKindWolf
	EnemyKindBandit
)

// Entity — сущность в мире (враг, NPC и т.д.).
type Entity struct {
	ID       EntityID
	Type     EntityType
	Kind     EnemyKind // для врагов: вид (Slime, Wolf, Bandit и т.д.)
	X        int
	Y        int
	Alive    bool
}

// EntitySpawnKey — ключ для отслеживания побеждённых спавнов врагов.
type EntitySpawnKey struct {
	Type EntityType
	X    int
	Y    int
}
