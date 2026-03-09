package generation

import "mygame/world/mapdata"

// GenerateTile возвращает тип тайла для клетки мира.
func GenerateTile(worldX, worldY, seed int) mapdata.TileType {
	if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
		return mapdata.TileFloor
	}
	terrain := terrainValue(worldX, worldY, seed)
	if isBlockedTile(terrain) {
		return blockedTileType(terrain)
	}
	detail := detailValue(worldX, worldY, seed)
	return walkableTileType(detail)
}

func terrainValue(worldX, worldY, seed int) int {
	scale := 24.0
	n := FractalNoise2D(float64(worldX)/scale, float64(worldY)/scale, seed, 4, 0.5, 2.0)
	return int(n * 100)
}

func detailValue(worldX, worldY, seed int) int {
	scale := 10.0
	n := FractalNoise2D(float64(worldX)/scale, float64(worldY)/scale, seed+1337, 3, 0.55, 2.0)
	return int(n * 100)
}

func isBlockedTile(terrain int) bool {
	return terrain < 26 || terrain > 75
}

func blockedTileType(terrain int) mapdata.TileType {
	if terrain < 26 {
		return mapdata.TileWater
	}
	return mapdata.TileWall
}

func walkableTileType(detail int) mapdata.TileType {
	if detail > 68 {
		return mapdata.TileGrass
	}
	return mapdata.TileFloor
}
