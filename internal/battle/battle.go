package battle

import (
	"mygame/world"
)

// BattleContext хранит состояние одного активного боя.
type BattleContext struct {
	// Encounter — источник боя (связь с world entities).
	Encounter Encounter

	// EnemyID — ID врага из мира (удобный доступ к SourceEnemyID).
	EnemyID world.EntityID

	// PlayerHP и EnemyHP — боевые очки здоровья.
	PlayerHP int
	EnemyHP  int

	Phase  BattlePhase
	LastLog string
}

// BuildBattleContextFromEncounter создаёт BattleContext из Encounter.
func BuildBattleContextFromEncounter(enc Encounter) *BattleContext {
	if len(enc.Enemies) == 0 {
		return nil
	}
	seed := BuildBattleUnitSeed(enc.Enemies[0])
	return &BattleContext{
		Encounter: enc,
		EnemyID:   enc.SourceEnemyID,
		PlayerHP:  10,
		EnemyHP:   seed.MaxHP,
		Phase:     BattlePhasePlayerTurn,
		LastLog:   "Бой начался. Ход игрока.",
	}
}
