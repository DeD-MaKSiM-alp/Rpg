package entity

// IntentType — тип намерения врага на ход.
type IntentType int

const (
	IntentWait IntentType = iota
	IntentMove
	IntentAttack
)

// Intent — намерение одной сущности на ход (не меняет состояние мира).
type Intent struct {
	EntityID EntityID
	Type     IntentType
	TargetX  int
	TargetY  int
}

// IsAdjacent8 возвращает true, если клетки (ax,ay) и (bx,by) соседние по 8 направлениям.
// Возвращает false для одной и той же клетки и если дистанция больше 1 по любой оси.
func IsAdjacent8(ax, ay, bx, by int) bool {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx > 1 || dy > 1 {
		return false
	}
	return dx != 0 || dy != 0
}

// StepToward возвращает желаемый шаг (dx, dy) от (fromX, fromY) к (toX, toY).
// Возвращает -1/0/1 по каждой оси, допускает диагональ. Не применяет движение.
func StepToward(fromX, fromY, toX, toY int) (dx, dy int) {
	if toX > fromX {
		dx = 1
	} else if toX < fromX {
		dx = -1
	}
	if toY > fromY {
		dy = 1
	} else if toY < fromY {
		dy = -1
	}
	return dx, dy
}
