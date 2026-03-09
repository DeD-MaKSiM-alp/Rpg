package world

// generateTile определяет итоговый тайл мировой клетки.
//
// Новый подход двухслойный:
//  1. сначала определяем, клетка проходимая или нет;
//  2. затем выбираем конкретный тип тайла внутри этого слоя.
//
// Это удобнее старой схемы, где мы сразу выбирали TileFloor/TileGrass/TileWater/TileWall.
// Теперь гораздо проще контролировать:
//   - сколько в мире проходимого пространства;
//   - сколько препятствий;
//   - какие именно тайлы должны появляться внутри каждой группы.
func generateTile(worldX, worldY, seed int) TileType {
	// Стартовую область вокруг игрока оставляем свободной,
	// чтобы старт всегда был безопасным и удобным для движения.
	if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
		return TileFloor
	}

	terrain := terrainValue(worldX, worldY, seed)

	if isBlockedTile(terrain) {
		return blockedTileType(terrain)
	}

	detail := detailValue(worldX, worldY, seed)
	return walkableTileType(detail)
}

// terrainValue возвращает базовое значение местности в диапазоне 0..99.
//
// Для формы мира используем многослойный noise:
// крупная октава задаёт большие области,
// дополнительные октавы делают края и переходы менее примитивными.
func terrainValue(worldX, worldY, seed int) int {
	scale := 24.0

	n := fractalNoise2D(
		float64(worldX)/scale,
		float64(worldY)/scale,
		seed,
		4,   // octaves
		0.5, // persistence
		2.0, // lacunarity
	)

	return int(n * 100)
}

// detailValue возвращает дополнительное значение в диапазоне 0..99.
//
// Этот noise используется для вариаций внутри уже выбранного слоя:
// например, где будет обычный пол, а где трава.
func detailValue(worldX, worldY, seed int) int {
	scale := 10.0

	n := fractalNoise2D(
		float64(worldX)/scale,
		float64(worldY)/scale,
		seed+1337,
		3,    // octaves
		0.55, // persistence
		2.0,  // lacunarity
	)

	return int(n * 100)
}

// isBlockedTile решает, должна ли клетка быть непроходимой.
//
// На этом этапе мир должен быть в основном проходимым,
// поэтому blocked-зона должна занимать меньшую часть карты.
// Здесь мы сознательно делаем широкую "сушу" посередине
// и только по краям диапазона получаем препятствия.
func isBlockedTile(terrain int) bool {
	return terrain < 26 || terrain > 75
}

// blockedTileType выбирает тип непроходимого тайла.
//
// Низкие значения terrain превращаем в воду,
// высокие — в стены.
func blockedTileType(terrain int) TileType {
	if terrain < 26 {
		return TileWater
	}

	return TileWall
}

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

// walkableTileType выбирает тип проходимого тайла.
//
// Основная масса — обычный пол,
// часть клеток — трава как проходимая вариация поверхности.
func walkableTileType(detail int) TileType {
	if detail > 68 {
		return TileGrass
	}

	return TileFloor
}

func (w *World) IsWalkable(x, y int) bool {
	coord, localX, localY := worldToChunkLocal(x, y)
	chunk := w.getOrCreateChunk(coord)

	tile := chunk.tiles[localY][localX]
	return isTileWalkable(tile)
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
			return true
		}
	}

	return false
}

func generatePickupsForChunk(chunkX, chunkY, seed int, tiles [][]TileType) []Pickup {
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
