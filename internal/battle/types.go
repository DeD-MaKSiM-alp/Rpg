package battle

// Phase — стадия battle loop.
type Phase int

const (
	PhaseStart Phase = iota
	PhaseTurnStart
	PhaseAwaitAction
	PhaseTurnResolve
	PhaseActionPause
	PhaseTurnEnd
	PhaseRoundEnd
	PhaseFinishedWaitInput
	PhaseFinished // legacy; переход через PhaseFinishedWaitInput
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

// BattleOutcome — результат одного обновления боя (внешний API для Game).
type BattleOutcome int

const (
	BattleOutcomeNone BattleOutcome = iota
	BattleOutcomeVictory
	BattleOutcomeDefeat
	BattleOutcomeRetreat
)
