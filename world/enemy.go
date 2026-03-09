package world

func enemySpawnKey(worldX, worldY int) EntitySpawnKey {
	return EntitySpawnKey{
		Type: EntityEnemy,
		X:    worldX,
		Y:    worldY,
	}
}

func (w *World) addEntity(entityType EntityType, worldX, worldY int) *Entity {
	w.nextEntityID++

	entity := &Entity{
		ID:    w.nextEntityID,
		Type:  entityType,
		X:     worldX,
		Y:     worldY,
		Alive: true,
	}

	w.entities[entity.ID] = entity
	return entity
}

func (w *World) GetEnemyAt(worldX, worldY int) *Entity {
	for _, entity := range w.entities {
		if !entity.Alive {
			continue
		}

		if entity.Type != EntityEnemy {
			continue
		}

		if entity.X == worldX && entity.Y == worldY {
			return entity
		}
	}

	return nil
}

func (w *World) isEnemyBlockingTile(worldX, worldY int, ignoreID EntityID) bool {
	for _, entity := range w.entities {
		if !entity.Alive {
			continue
		}

		if entity.Type != EntityEnemy {
			continue
		}

		if entity.ID == ignoreID {
			continue
		}

		if entity.X == worldX && entity.Y == worldY {
			return true
		}
	}

	return false
}

func (w *World) tryMoveEnemy(entity *Entity, nextX, nextY, playerX, playerY int) bool {
	// Враг не может встать в клетку игрока.
	if nextX == playerX && nextY == playerY {
		return false
	}

	// Враг не может пройти в непроходимый тайл.
	if !w.IsWalkable(nextX, nextY) {
		return false
	}

	// Враг не может встать в клетку другого врага.
	if w.isEnemyBlockingTile(nextX, nextY, entity.ID) {
		return false
	}

	entity.X = nextX
	entity.Y = nextY
	return true
}

func (w *World) RemoveEnemy(id EntityID) {
	entity, exists := w.entities[id]
	if !exists {
		return
	}

	entity.Alive = false
	w.defeatedEnemySpawns[enemySpawnKey(entity.X, entity.Y)] = true
	delete(w.entities, id)
}

func (w *World) generateEnemiesForChunk(chunkX, chunkY, seed int, tiles [][]TileType) {
	// В стартовом чанке врагов не создаём.
	if chunkX == 0 && chunkY == 0 {
		return
	}

	// Не в каждом чанке есть враг.
	spawnRoll := hash2D(chunkX, chunkY, seed+9000) % 100
	if spawnRoll >= 22 {
		return
	}

	// Пытаемся найти одну подходящую клетку.
	for attempt := 0; attempt < 8; attempt++ {
		localX := hash2D(chunkX, chunkY, seed+10000+attempt*19) % chunkSize
		localY := hash2D(chunkY, chunkX, seed+11000+attempt*29) % chunkSize

		tile := tiles[localY][localX]
		if !isTileWalkable(tile) {
			continue
		}

		worldX := chunkX*chunkSize + localX
		worldY := chunkY*chunkSize + localY

		// Не ставим врага слишком близко к старту.
		if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
			continue
		}

		spawnKey := enemySpawnKey(worldX, worldY)

		// Если враг уже был побеждён — не возрождаем.
		if w.defeatedEnemySpawns[spawnKey] {
			return
		}

		// Если такой враг уже существует в мире — не дублируем.
		if w.GetEnemyAt(worldX, worldY) != nil {
			return
		}

		w.addEntity(EntityEnemy, worldX, worldY)
		return
	}
}

func manhattanDistance(ax, ay, bx, by int) int {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}

	dy := ay - by
	if dy < 0 {
		dy = -dy
	}

	return dx + dy
}

func (w *World) AdvanceTurn(playerX, playerY int) {
	// Чтобы поведение не зависело от случайного порядка обхода map,
	// сначала собираем врагов в срез.
	enemies := make([]*Entity, 0, len(w.entities))
	for _, entity := range w.entities {
		if !entity.Alive {
			continue
		}

		if entity.Type != EntityEnemy {
			continue
		}

		enemies = append(enemies, entity)
	}

	for _, enemy := range enemies {
		// Если игрок слишком далеко, враг пока ничего не делает.
		dist := manhattanDistance(enemy.X, enemy.Y, playerX, playerY)
		if dist > 6 {
			continue
		}

		dx := 0
		dy := 0

		if playerX > enemy.X {
			dx = 1
		} else if playerX < enemy.X {
			dx = -1
		}

		if playerY > enemy.Y {
			dy = 1
		} else if playerY < enemy.Y {
			dy = -1
		}

		// Сначала пытаемся идти по диагонали,
		// если игрок смещён и по X, и по Y.
		if dx != 0 && dy != 0 {
			if w.tryMoveEnemy(enemy, enemy.X+dx, enemy.Y+dy, playerX, playerY) {
				continue
			}
		}

		// Если диагональ не получилась — пробуем по X.
		if dx != 0 {
			if w.tryMoveEnemy(enemy, enemy.X+dx, enemy.Y, playerX, playerY) {
				continue
			}
		}

		// Если по X тоже не получилось — пробуем по Y.
		if dy != 0 {
			if w.tryMoveEnemy(enemy, enemy.X, enemy.Y+dy, playerX, playerY) {
				continue
			}
		}
	}
}
