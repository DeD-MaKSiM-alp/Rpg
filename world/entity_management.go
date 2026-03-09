package world

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

func (w *World) RemoveEnemy(id EntityID) {
	entity, exists := w.entities[id]
	if !exists {
		return
	}

	entity.Alive = false
	w.defeatedEnemySpawns[enemySpawnKey(entity.X, entity.Y)] = true
	delete(w.entities, id)
}
