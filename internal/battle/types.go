package battle

// Phase — стадия battle loop.
type Phase int

const (
	PhaseStart Phase = iota
	PhaseTurnStart
	PhaseAwaitAction
	PhaseTurnResolve
	PhaseTurnEnd
	PhaseRoundEnd
	PhaseFinished
)

// Result — итог боя.
type Result int

const (
	ResultNone Result = iota
	ResultVictory
	ResultDefeat
	ResultEscape
)

// BattlePhase — упрощённая фаза для UI (ход игрока / ход врага / завершён).
type BattlePhase int

const (
	BattlePhasePlayerTurn BattlePhase = iota
	BattlePhaseEnemyTurn
	BattlePhaseFinished
)

// BattleAction описывает результат одного обновления боя (внешний API для Game).
type BattleAction int

const (
	BattleActionNone BattleAction = iota
	BattleActionVictory
	BattleActionDefeat
	BattleActionRetreat
)
