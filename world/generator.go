package world

// isTileWalkable определяет, можно ли пройти по конкретному типу тайла.
//
// Это важный шаг для расширения мира:
// теперь игровая логика не зависит от проверки "tile == TileFloor".
// Мы можем добавлять новые типы клеток и отдельно решать,
// какие из них проходимы, а какие нет.
func isTileWalkable(tile TileType) bool {
	switch tile {
	case TileFloor, TileGrass:
		return true
	case TileWall, TileWater:
		return false
	default:
		return false
	}
}

func (w *World) IsWalkable(x, y int) bool {
	coord, localX, localY := worldToChunkLocal(x, y)
	chunk := w.getOrCreateChunk(coord)

	tile := chunk.tiles[localY][localX]
	return isTileWalkable(tile)
}

func pickupKey(worldX, worldY int) PickupKey {
	return PickupKey{
		X: worldX,
		Y: worldY,
	}
}

func (w *World) CollectPickupAt(worldX, worldY int) bool {
	coord, _, _ := worldToChunkLocal(worldX, worldY)
	chunk := w.getOrCreateChunk(coord)

	for i := range chunk.pickups {
		pickup := &chunk.pickups[i]

		if pickup.Collected {
			continue
		}

		if pickup.X == worldX && pickup.Y == worldY {
			pickup.Collected = true
			w.collectedPickups[pickupKey(worldX, worldY)] = true
			return true
		}
	}

	return false
}

func generatePickupsForChunk(w *World, chunkX, chunkY, seed int, tiles [][]TileType) []Pickup {
	// Не спавним pickup в стартовом чанке, чтобы стартовая зона
	// пока оставалась максимально чистой и предсказуемой.
	if chunkX == 0 && chunkY == 0 {
		return nil
	}

	// Решаем, будет ли вообще pickup в этом чанке.
	// Делаем это детерминированно через seed и координаты чанка.
	spawnRoll := hash2D(chunkX, chunkY, seed+5000) % 100
	if spawnRoll >= 28 {
		return nil
	}

	// Пытаемся несколько раз найти подходящую проходимую клетку.
	for attempt := 0; attempt < 8; attempt++ {
		localX := hash2D(chunkX, chunkY, seed+6000+attempt*17) % chunkSize
		localY := hash2D(chunkY, chunkX, seed+7000+attempt*23) % chunkSize

		tile := tiles[localY][localX]
		if !isTileWalkable(tile) {
			continue
		}

		worldX := chunkX*chunkSize + localX
		worldY := chunkY*chunkSize + localY

		// Не кладём pickup слишком близко к старту.
		if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
			continue
		}

		if w.collectedPickups[pickupKey(worldX, worldY)] {
			return nil
		}

		return []Pickup{
			{
				X:         worldX,
				Y:         worldY,
				Collected: false,
			},
		}
	}

	return nil
}
