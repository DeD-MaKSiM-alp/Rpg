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
