package player

import (
	"mygame/world"
)

// WorldForMove — минимальный интерфейс мира, нужный для TryMovePlayer.
type WorldForMove interface {
	GetEnemyAt(x, y int) *world.Entity
	IsWalkable(x, y int) bool
	CollectPickupAt(x, y int) bool
}

// TryMovePlayer пытается переместить игрока на одну клетку в заданном направлении.
// Возвращает: удалось ли выполнить действие (в т.ч. вступление в бой), ID врага при контакте, был ли подобран предмет.
func TryMovePlayer(pl *Player, w WorldForMove, dx, dy int) (moved bool, enemyID world.EntityID, pickedUp bool) {
	nextX := pl.GridX + dx
	nextY := pl.GridY + dy

	enemy := w.GetEnemyAt(nextX, nextY)
	if enemy != nil {
		return true, enemy.ID, false
	}

	if !w.IsWalkable(nextX, nextY) {
		return false, 0, false
	}

	pl.Move(dx, dy)
	if w.CollectPickupAt(pl.GridX, pl.GridY) {
		pickedUp = true
	}
	return true, 0, pickedUp
}
