package player

import (
	"mygame/world"
)

// WorldForMove — минимальный интерфейс мира, нужный для TryMovePlayer.
type WorldForMove interface {
	GetEnemyAt(x, y int) *world.Entity
	IsWalkable(x, y int) bool
	CollectPickupAt(x, y int) bool
	AdvanceTurn(px, py int) (world.EntityID, bool)
}

// TryMovePlayer пытается переместить игрока на одну клетку в заданном направлении.
func TryMovePlayer(pl *Player, w WorldForMove, pickupCount *int, startBattle func(world.EntityID), dx, dy int) {
	nextX := pl.GridX + dx
	nextY := pl.GridY + dy

	enemy := w.GetEnemyAt(nextX, nextY)
	if enemy != nil {
		startBattle(enemy.ID)
		return
	}

	if !w.IsWalkable(nextX, nextY) {
		return
	}

	pl.Move(dx, dy)

	if w.CollectPickupAt(pl.GridX, pl.GridY) {
		*pickupCount++
	}

	enemyID, startedBattle := w.AdvanceTurn(pl.GridX, pl.GridY)
	if startedBattle {
		startBattle(enemyID)
	}
}
