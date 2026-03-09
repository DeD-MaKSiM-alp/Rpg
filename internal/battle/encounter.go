package battle

import (
	"mygame/world"
	"mygame/world/entity"
)

// Encounter описывает предстоящий бой (подготовительные данные, не runtime battle state).
type Encounter struct {
	SourceEnemyID entity.EntityID // кто инициировал бой
	Enemies       []EncounterEnemy
}

// EncounterEnemy — участник encounter из мира.
type EncounterEnemy struct {
	WorldEnemyID entity.EntityID
	Kind         entity.EnemyKind
}

// BuildEncounterFromWorld строит Encounter по enemyID из мира. Не меняет мир.
func BuildEncounterFromWorld(w *world.World, enemyID entity.EntityID) (Encounter, bool) {
	e := w.GetEntityByID(enemyID)
	if e == nil || !e.Alive || e.Type != entity.EntityEnemy {
		return Encounter{}, false
	}
	return Encounter{
		SourceEnemyID: enemyID,
		Enemies: []EncounterEnemy{{
			WorldEnemyID: enemyID,
			Kind:         entity.EnemyKind(e.Kind),
		}},
	}, true
}

// BattleUnitSeed — подготовительные боевые данные для создания юнита (отвязаны от world entity).
type BattleUnitSeed struct {
	Name          string
	MaxHP         int
	Attack        int
	Defense       int
	Initiative    int
	IsRanged      bool
	SourceEnemyID entity.EntityID
}

// BuildBattleUnitSeed преобразует EncounterEnemy в BattleUnitSeed.
func BuildBattleUnitSeed(e EncounterEnemy) BattleUnitSeed {
	tpl := GetEnemyTemplate(e.Kind)
	return BattleUnitSeed{
		Name:          tpl.Name,
		MaxHP:         tpl.MaxHP,
		Attack:        tpl.Attack,
		Defense:       tpl.Defense,
		Initiative:    tpl.Initiative,
		IsRanged:      tpl.IsRanged,
		SourceEnemyID: e.WorldEnemyID,
	}
}
