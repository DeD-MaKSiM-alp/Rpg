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

// CombatUnitSeed — вход для построения боевого юнита из внешнего контекста (encounter/world, player snapshot и т.п.).
// Это "conversion layer": на выходе даёт definition (archetype) + origin hook, без смешивания с runtime state боя.
type CombatUnitSeed struct {
	Def    CombatUnitDefinition
	Origin CombatUnitOrigin
}

// BuildEnemyCombatUnitSeed преобразует EncounterEnemy в CombatUnitSeed.
func BuildEnemyCombatUnitSeed(e EncounterEnemy) CombatUnitSeed {
	tpl := GetEnemyTemplate(e.Kind)
	abils := GetRoleAbilities(tpl.Role)
	if len(abils) == 0 {
		abils = []AbilityID{AbilityBasicAttack}
	}
	return CombatUnitSeed{
		Def: CombatUnitDefinition{
			ArchetypeID: "enemy:" + tpl.Name,
			DisplayName: tpl.Name,
			Role:        tpl.Role,
			Base: UnitBaseStats{
				MaxHP:      tpl.HP,
				Attack:     tpl.Attack,
				Defense:    tpl.Defense,
				Initiative: tpl.Initiative,
			},
			IsRanged: tpl.IsRanged,
			Loadout:  AbilityLoadout{Abilities: abils},
		},
		Origin: CombatUnitOrigin{WorldEnemyID: e.EnemyID},
	}
}

// DefaultPlayerCombatUnitSeed — временный seed игрока до появления party/persistent snapshot.
func DefaultPlayerCombatUnitSeed() CombatUnitSeed {
	return CombatUnitSeed{
		Def: CombatUnitDefinition{
			ArchetypeID: "player:default",
			DisplayName: "Игрок",
			Role:        RoleFighter,
			Base: UnitBaseStats{
				MaxHP:      10,
				Attack:     2,
				Defense:    0,
				Initiative: 2,
			},
			IsRanged: false,
			Loadout:  AbilityLoadout{Abilities: []AbilityID{AbilityBasicAttack}},
		},
	}
}
