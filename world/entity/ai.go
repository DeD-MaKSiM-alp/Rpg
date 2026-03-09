package entity

// ManhattanDistance возвращает расстояние Манхэттена между двумя точками.
func ManhattanDistance(ax, ay, bx, by int) int {
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

// IsPlayerTile возвращает true, если (nextX, nextY) совпадает с (playerX, playerY).
func IsPlayerTile(nextX, nextY, playerX, playerY int) bool {
	return nextX == playerX && nextY == playerY
}

// EnemyStepTowardsPlayer возвращает направление шага (dx, dy) к игроку.
func EnemyStepTowardsPlayer(enemyX, enemyY, playerX, playerY int) (dx, dy int) {
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
