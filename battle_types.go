package main

// BattlePhase описывает текущую фазу боя.
//
// Пока нам нужны только три состояния:
// - ход игрока;
// - ход врага;
// - бой завершён.
//
// Позже сюда удобно добавлять более детальные фазы:
// - выбор действия;
// - выбор цели;
// - анимацию применения удара;
// - показ результата раунда.
type BattlePhase int

const (
	BattlePhasePlayerTurn BattlePhase = iota
	BattlePhaseEnemyTurn
	BattlePhaseFinished
)

// BattleAction описывает результат одного обновления боя.
//
// Game не должен знать внутренние детали боевой логики,
// поэтому BattleContext возвращает наружу только итог:
// - ничего не произошло;
// - игрок победил;
// - игрок проиграл;
// - игрок вышел из боя вручную.
type BattleAction int

const (
	BattleActionNone BattleAction = iota
	BattleActionVictory
	BattleActionDefeat
	BattleActionRetreat
)
