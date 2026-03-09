package world

type EntityID int

type EntityType int

const (
	EntityEnemy EntityType = iota
)

type Entity struct {
	ID   EntityID
	Type EntityType

	X int
	Y int

	Alive bool
}

type EntitySpawnKey struct {
	Type EntityType
	X    int
	Y    int
}
