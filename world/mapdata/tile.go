package mapdata

// TileType — тип клетки мира.
type TileType int

const (
	TileFloor TileType = iota
	TileWall
	TileGrass
	TileWater
)
