package game

type PlayerActionType int

const (
	ActionNone PlayerActionType = iota
	ActionMove
	ActionWait
)

type PlayerAction struct {
	Type PlayerActionType
	DX   int
	DY   int
}
