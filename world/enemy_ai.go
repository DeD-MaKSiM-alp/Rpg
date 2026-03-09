package world

func (w *World) AdvanceTurn(playerX, playerY int) (EntityID, bool) {
	enemies := w.collectAliveEnemies()

	for _, enemy := range enemies {
		// Если игрок слишком далеко, враг пока ничего не делает.
		dist := manhattanDistance(enemy.X, enemy.Y, playerX, playerY)
		if dist > 6 {
			continue
		}

		dx, dy := enemyStepTowardsPlayer(enemy.X, enemy.Y, playerX, playerY)

		if id, engaged := w.advanceEnemyOneStep(enemy, dx, dy, playerX, playerY); engaged {
			return id, true
		}
	}

	return 0, false
}

// collectAliveEnemies собирает всех живых врагов в срез,
// чтобы порядок обработки не зависел от порядка в map.
func (w *World) collectAliveEnemies() []*Entity {
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
	return enemies
}

// enemyStepTowardsPlayer вычисляет направление шага врага к игроку.
func enemyStepTowardsPlayer(enemyX, enemyY, playerX, playerY int) (dx, dy int) {
	if playerX > enemyX {
		dx = 1
	} else if playerX < enemyX {
		dx = -1
	}

	if playerY > enemyY {
		dy = 1
	} else if playerY < enemyY {
		dy = -1
	}

	return dx, dy
}

// advanceEnemyOneStep выполняет логику одного хода врага:
// попытка диагонального шага, затем по X, затем по Y.
// Если враг входит в клетку игрока — возвращаем ID врага и true.
func (w *World) advanceEnemyOneStep(enemy *Entity, dx, dy, playerX, playerY int) (EntityID, bool) {
	// 1) Сначала пробуем диагональ.
	if dx != 0 && dy != 0 {
		nextX := enemy.X + dx
		nextY := enemy.Y + dy

		if isPlayerTile(nextX, nextY, playerX, playerY) {
			return enemy.ID, true
		}

		if w.tryMoveEnemy(enemy, nextX, nextY, playerX, playerY) {
			return 0, false
		}
	}

	// 2) Если диагональ не получилась — пробуем по X.
	if dx != 0 {
		nextX := enemy.X + dx
		nextY := enemy.Y

		if isPlayerTile(nextX, nextY, playerX, playerY) {
			return enemy.ID, true
		}

		if w.tryMoveEnemy(enemy, nextX, nextY, playerX, playerY) {
			return 0, false
		}
	}

	// 3) Если по X не получилось — пробуем по Y.
	if dy != 0 {
		nextX := enemy.X
		nextY := enemy.Y + dy

		if isPlayerTile(nextX, nextY, playerX, playerY) {
			return enemy.ID, true
		}

		if w.tryMoveEnemy(enemy, nextX, nextY, playerX, playerY) {
			return 0, false
		}
	}

	return 0, false
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
