package battle

// Phase — стадия battle loop.
type Phase int

const (
	PhaseStart Phase = iota
	PhaseTurnStart
	PhaseAwaitAction
	PhaseActionPause
	PhaseTurnEnd
	PhaseRoundEnd
	PhaseFinishedWaitInput
)

// Result — итог боя.
type Result int

const (
	ResultNone Result = iota
	ResultVictory
	ResultDefeat
	ResultEscape
)

// BattleOutcome — результат одного обновления боя (внешний API для Game).
type BattleOutcome int

const (
	BattleOutcomeNone BattleOutcome = iota
	BattleOutcomeVictory
	BattleOutcomeDefeat
	BattleOutcomeRetreat
)
