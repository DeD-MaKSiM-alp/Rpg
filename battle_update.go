package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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
