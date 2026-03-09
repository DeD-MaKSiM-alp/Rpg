package main

/*
TryMovePlayer пытается переместить игрока на одну клетку в заданном направлении.
- Проверяет, не находится ли игрок на клетке с врагом;
- Проверяет, можно ли двигаться на эту клетку;
- Перемещает игрока;
- Собирает pickup, если есть;
- Продвигает мир на один ход, чтобы враги могли двигаться.
- Если на клетке с врагом, начинает бой.
*/
func (g *Game) TryMovePlayer(dx, dy int) {
	nextX := g.player.gridX + dx
	nextY := g.player.gridY + dy

	// Если в целевой клетке враг — не двигаемся,
	// а входим в режим боя.
	enemy := g.world.GetEnemyAt(nextX, nextY)
	if enemy != nil {
		g.startBattle(enemy.ID)
		return
	}

	if !g.world.IsWalkable(nextX, nextY) {
		return
	}

	g.player.Move(dx, dy)

	if g.world.CollectPickupAt(g.player.gridX, g.player.gridY) {
		g.pickupCount++
	}

	enemyID, startedBattle := g.world.AdvanceTurn(g.player.gridX, g.player.gridY)
	if startedBattle {
		g.startBattle(enemyID)
	}
}
