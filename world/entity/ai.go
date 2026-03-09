package entity

// AggroRadiusManhattan — радиус агра врага по Манхэттену (враг не агрится дальше).
const AggroRadiusManhattan = 6

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

// BuildEnemyIntent возвращает намерение врага на этот ход. Не меняет состояние мира.
// Если враг неактивен/невалиден — wait. Если игрок далеко — wait.
// Если враг соседний с игроком (8 направлений) — attack по координатам игрока.
// Иначе при игроке в радиусе агра — move в сторону игрока. Иначе — wait.
func BuildEnemyIntent(e *Entity, playerX, playerY int) Intent {
	if e == nil || !e.Alive || e.Type != EntityEnemy {
		id := EntityID(0)
		if e != nil {
			id = e.ID
		}
		return Intent{EntityID: id, Type: IntentWait}
	}
	if ManhattanDistance(e.X, e.Y, playerX, playerY) > AggroRadiusManhattan {
		return Intent{EntityID: e.ID, Type: IntentWait}
	}
	if IsAdjacent8(e.X, e.Y, playerX, playerY) {
		return Intent{
			EntityID: e.ID,
			Type:     IntentAttack,
			TargetX:  playerX,
			TargetY:  playerY,
		}
	}
	dx, dy := StepToward(e.X, e.Y, playerX, playerY)
	return Intent{
		EntityID: e.ID,
		Type:     IntentMove,
		TargetX:  e.X + dx,
		TargetY:  e.Y + dy,
	}
}
