package main

import (
	"mygame/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// BattleAction описывает итог одного шага боевого режима.
//
// Пока нам нужны только три варианта:
// - ничего не произошло;
// - игрок тестово победил;
// - игрок вышел из боя без победы.
//
// Позже сюда удобно добавлять:
// - конец раунда;
// - выбор цели;
// - применение способности;
// - завершение настоящего боя с результатом.
type BattleAction int

const (
	BattleActionNone BattleAction = iota
	BattleActionVictory
	BattleActionRetreat
)

// BattleContext хранит состояние одного активного боя.
//
// Пока это минимальная версия:
// мы знаем только, с каким врагом сейчас сражаемся.
//
// Важно:
// дальше именно сюда удобно добавлять:
// - участников боя;
// - очередь хода;
// - выбранное действие;
// - цели;
// - результат боя.
type BattleContext struct {
	// EnemyID — ID врага из world, с которым начался бой.
	EnemyID world.EntityID
}

// Update обрабатывает один кадр боевого режима
// и возвращает результат этого кадра.
//
// Сейчас это тестовая логика:
// - B означает тестовую победу;
// - Escape означает выход из боя без победы.
//
// Позже этот метод станет главным входом
// для всей пошаговой боевой логики.
func (b *BattleContext) Update() BattleAction {
	// Защита на случай некорректного вызова.
	if b == nil {
		return BattleActionNone
	}

	// B = тестовая победа в бою.
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		return BattleActionVictory
	}

	// Escape = выйти из боя без победы.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return BattleActionRetreat
	}

	return BattleActionNone
}
