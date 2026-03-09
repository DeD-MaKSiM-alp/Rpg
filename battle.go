package main

import (
	"mygame/world"
)

// BattleContext хранит состояние одного активного боя.
//
// Пока это ещё не "настоящий Disciples-бой", а минимальный пошаговый каркас:
// - есть один игрок;
// - есть один враг;
// - обе стороны имеют тестовый HP;
// - стороны ходят по очереди.
type BattleContext struct {
	// EnemyID — ID врага из мира, с которым начался бой.
	EnemyID world.EntityID

	// PlayerHP и EnemyHP — временные тестовые боевые очки здоровья.
	// Это отдельные боевые значения внутри текущего столкновения.
	PlayerHP int
	EnemyHP  int

	// Phase — текущая фаза боя.
	Phase BattlePhase

	// LastLog — короткое текстовое описание последнего события в бою.
	// Это удобно и для отладки, и для первого боевого UI.
	LastLog string
}

// NewBattleContext создаёт новый контекст боя.
//
// Пока задаём фиксированные тестовые значения,
// чтобы быстро получить рабочий пошаговый бой.
// Позже здесь можно будет подтягивать реальные статы игрока и врага.
func NewBattleContext(enemyID world.EntityID) *BattleContext {
	return &BattleContext{
		EnemyID:  enemyID,
		PlayerHP: 10,
		EnemyHP:  6,
		Phase:    BattlePhasePlayerTurn,
		LastLog:  "Бой начался. Ход игрока.",
	}
}

