package entity

// GetEnemyAt возвращает врага в клетке (worldX, worldY) или nil.
func GetEnemyAt(entities map[EntityID]*Entity, worldX, worldY int) *Entity {
	for _, e := range entities {
		if !e.Alive || e.Type != EntityEnemy {
			continue
		}
		if e.X == worldX && e.Y == worldY {
			return e
		}
	}
	return nil
}

// IsEnemyBlockingTile возвращает true, если в клетке есть живой враг (кроме ignoreID).
func IsEnemyBlockingTile(entities map[EntityID]*Entity, worldX, worldY int, ignoreID EntityID) bool {
	for _, e := range entities {
		if !e.Alive || e.Type != EntityEnemy || e.ID == ignoreID {
			continue
		}
		if e.X == worldX && e.Y == worldY {
			return true
		}
	}
	return false
}

// RemoveEnemy помечает врага мёртвым, записывает спавн в defeated и удаляет из entities.
func RemoveEnemy(entities map[EntityID]*Entity, defeated map[EntitySpawnKey]bool, id EntityID) {
	e, ok := entities[id]
	if !ok {
		return
	}
	e.Alive = false
	defeated[SpawnKey(e.X, e.Y)] = true
	delete(entities, id)
}

// SpawnKey возвращает ключ спавна для координат (для врагов).
func SpawnKey(worldX, worldY int) EntitySpawnKey {
	return EntitySpawnKey{Type: EntityEnemy, X: worldX, Y: worldY}
}
