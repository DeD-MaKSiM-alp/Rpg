package player

import (
	"mygame/world"
)

// WorldForMove — минимальный интерфейс мира, нужный для TryMovePlayer.
type WorldForMove interface {
	GetEnemyAt(x, y int) *world.Entity
	IsWalkable(x, y int) bool
	InteractPickupAfterMove(x, y int) world.PickupInteractionResult
}

// TryMovePlayer tries to move the player or initiate combat.
// "moved" means the action was accepted (step or contact with enemy), not necessarily that the position changed.
// pickup — результат взаимодействия с пикапом на клетке после шага (обычный / лагерь рекрута / нет).
func TryMovePlayer(pl *Player, w WorldForMove, dx, dy int) (moved bool, enemyID world.EntityID, pickup world.PickupInteractionResult) {
	nextX := pl.GridX + dx
	nextY := pl.GridY + dy

	enemy := w.GetEnemyAt(nextX, nextY)
	if enemy != nil {
		return true, enemy.ID, world.PickupInteractNone
	}

	if !w.IsWalkable(nextX, nextY) {
		return false, 0, world.PickupInteractNone
	}

	pl.Move(dx, dy)
	pickup = w.InteractPickupAfterMove(pl.GridX, pl.GridY)
	return true, 0, pickup
}
