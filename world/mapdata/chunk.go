package mapdata

import "mygame/world/entity"

// ChunkSize — размер одного чанка в клетках.
const ChunkSize = 16

// ChunkCoord — координаты чанка в мире.
type ChunkCoord struct {
	X int
	Y int
}

// PickupKey — ключ для собранных предметов.
type PickupKey struct {
	X int
	Y int
}

// Chunk — кусок мира фиксированного размера (тайлы и пикапы).
type Chunk struct {
	ChunkX  int
	ChunkY  int
	Tiles   [][]TileType
	Pickups []entity.Pickup
}
