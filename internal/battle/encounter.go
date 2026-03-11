package battle

import (
	"mygame/world"
	"mygame/world/entity"
)

// Encounter описывает предстоящий бой (слой между world enemy и battle unit).
type Encounter struct {
	Enemies []EncounterEnemy
}

// EncounterEnemy — участник encounter (ссылка на врага в мире и его тип).
type EncounterEnemy struct {
	EnemyID entity.EntityID
	Kind    entity.EnemyKind
}

// BuildEncounterFromWorld строит Encounter по enemyID из мира. Не меняет мир.
func BuildEncounterFromWorld(w *world.World, enemyID entity.EntityID) (Encounter, bool) {
	e := w.GetEntityByID(enemyID)
	if e == nil || !e.Alive || e.Type != entity.EntityEnemy {
		return Encounter{}, false
	}
	return Encounter{
		Enemies: []EncounterEnemy{{
			EnemyID: enemyID,
			Kind:    entity.EnemyKind(e.Kind),
		}},
	}, true
}

// BattleUnitSeed — подготовительные боевые данные для создания BattleUnit (отвязаны от world entity).
type BattleUnitSeed struct {
	ArchetypeID    string
	Name          string
	MaxHP         int
	Attack        int
	Defense       int
	Initiative    int
	IsRanged      bool
	Role          Role
	Abilities     []AbilityID
	SourceEnemyID entity.EntityID
}

// BuildBattleUnitSeed преобразует EncounterEnemy в BattleUnitSeed.
func BuildBattleUnitSeed(e EncounterEnemy) BattleUnitSeed {
	tpl := GetEnemyTemplate(e.Kind)
	return BattleUnitSeed{
		ArchetypeID:    "enemy:" + tpl.Name,
		Name:          tpl.Name,
		MaxHP:         tpl.HP,
		Attack:        tpl.Attack,
		Defense:       tpl.Defense,
		Initiative:    tpl.Initiative,
		IsRanged:      tpl.IsRanged,
		Role:          tpl.Role,
		Abilities:     GetRoleAbilities(tpl.Role),
		SourceEnemyID: e.EnemyID,
	}
}
