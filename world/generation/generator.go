package generation

import "mygame/world/mapdata"

// IsTileWalkable возвращает, можно ли пройти по данному типу тайла.
func IsTileWalkable(tile mapdata.TileType) bool {
	switch tile {
	case mapdata.TileFloor, mapdata.TileGrass:
		return true
	case mapdata.TileWall, mapdata.TileWater:
		return false
	default:
		return false
	}
}
