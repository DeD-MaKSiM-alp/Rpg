package main

/*
Update обрабатывает один кадр игры:
- Если режим — боевой, обновляет бой;
- Иначе обновляет исследовательский режим.
*/
func (g *Game) Update() error {
	if g.mode == ModeBattle {
		g.updateBattleMode()
		return nil
	}

	return g.updateExploreMode()
}

// updateExploreMode обрабатывает один кадр в режиме исследования мира:
// читает ввод, обновляет буфер направления, двигает игрока и поддерживает мир вокруг.
func (g *Game) updateExploreMode() error {
	newDirection := g.readDirectionInput()

	if g.hasBufferedInput {
		g.updateBufferedInput(newDirection)
	} else {
		g.startInputBufferIfNeeded(newDirection)
	}

	g.updateCamera()
	g.updateStreamingWorld()

	return nil
}

/*
updateBattleMode обрабатывает один кадр боевого режима:
- читает ввод игрока;
- обновляет контекст боя;
- обрабатывает результаты боя;
- вызывает методы Game для обновления состояния игры.
*/
func (g *Game) updateBattleMode() {
	// Страховка:
	// если по какой-то причине игра находится в ModeBattle,
	// но контекст боя отсутствует, выходим обратно в исследование.
	if g.battle == nil {
		g.endBattle()
		return
	}

	// Делегируем обновление самому боевому контексту.
	// Game не знает, какие именно фазы и правила есть внутри боя.
	action := g.battle.Update()

	switch action {
	case BattleActionVictory:
		// При победе удаляем врага из мира
		// и выходим обратно в режим исследования.
		g.world.RemoveEnemy(g.battle.EnemyID)
		g.endBattle()
		return

	case BattleActionDefeat:
		// Пока что поражение просто завершает бой.
		//
		// Позже здесь можно будет добавить:
		// - экран поражения;
		// - откат игрока;
		// - штраф;
		// - загрузку сейва;
		// - последствия в мире.
		g.endBattle()
		return

	case BattleActionRetreat:
		// Выходим из боя без удаления врага.
		g.endBattle()
		return

	case BattleActionNone:
		// Бой продолжается, ничего дополнительно делать не нужно.
		return
	}
}
