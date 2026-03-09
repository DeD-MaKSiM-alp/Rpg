package main

import (
	"mygame/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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

// Update обрабатывает один кадр боевого режима
// и возвращает итог этого кадра.
//
// Важно:
// здесь живёт именно боевая логика.
// Game снаружи должен только вызвать Update()
// и обработать результат боя.
func (b *BattleContext) Update() BattleAction {
	// Защита от некорректного вызова.
	if b == nil {
		return BattleActionNone
	}

	// Escape позволяет выйти из тестового боя вручную.
	// Это полезно оставить даже после появления базовой боёвки.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		b.Phase = BattlePhaseFinished
		b.LastLog = "Игрок покинул бой."
		return BattleActionRetreat
	}

	switch b.Phase {
	case BattlePhasePlayerTurn:
		return b.updatePlayerTurn()

	case BattlePhaseEnemyTurn:
		return b.updateEnemyTurn()

	case BattlePhaseFinished:
		return BattleActionNone
	}

	return BattleActionNone
}

// updatePlayerTurn обрабатывает ход игрока.
//
// Пока у игрока только одно действие:
// Space = обычная тестовая атака.
// Позже здесь появятся выбор действия, способности, защита и цели.
func (b *BattleContext) updatePlayerTurn() BattleAction {
	// Space = базовая тестовая атака игрока.
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		damage := 2
		b.EnemyHP -= damage

		// Если враг побеждён — завершаем бой победой.
		if b.EnemyHP <= 0 {
			b.EnemyHP = 0
			b.Phase = BattlePhaseFinished
			b.LastLog = "Игрок атаковал и победил врага."
			return BattleActionVictory
		}

		// Иначе передаём ход врагу.
		b.Phase = BattlePhaseEnemyTurn
		b.LastLog = "Игрок атаковал. Ход врага."
	}

	return BattleActionNone
}

// updateEnemyTurn обрабатывает ход врага.
//
// Пока враг действует автоматически без ввода:
// наносит фиксированный тестовый урон и либо побеждает,
// либо передаёт ход обратно игроку.
//
// Важно:
// здесь нет проверки на клавиши.
// Враг должен ходить как часть внутренней логики боя.
func (b *BattleContext) updateEnemyTurn() BattleAction {
	damage := 1
	b.PlayerHP -= damage

	// Если игрок побеждён — завершаем бой поражением.
	if b.PlayerHP <= 0 {
		b.PlayerHP = 0
		b.Phase = BattlePhaseFinished
		b.LastLog = "Враг атаковал и победил игрока."
		return BattleActionDefeat
	}

	// Иначе возвращаем ход игроку.
	b.Phase = BattlePhasePlayerTurn
	b.LastLog = "Враг атаковал. Ход игрока."
	return BattleActionNone
}
