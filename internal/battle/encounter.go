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
	// InitialHP: если >0 — стартовое HP в бою (каноническое CurrentHP героя); если 0 — используется Def.Base.MaxHP.
	InitialHP int
}

// BuildEnemyCombatUnitSeed преобразует EncounterEnemy в CombatUnitSeed.
// escalationLevel: 0 = без усиления; 1+ = масштабирование статов (эскалация по выигранным боям).
func BuildEnemyCombatUnitSeed(e EncounterEnemy, escalationLevel int) CombatUnitSeed {
	tpl := GetEnemyTemplate(e.Kind)
	tpl = ScaleEnemyTemplate(tpl, escalationLevel)
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
		Origin: CombatUnitOrigin{WorldEnemyID: e.EnemyID, PartyActiveIndex: -1},
	}
}

// DefaultPlayerCombatUnitSeed — seed игрока по умолчанию (до прогрессии).
func DefaultPlayerCombatUnitSeed() CombatUnitSeed {
	return BuildPlayerCombatSeed(10, 2, 0, 2, []AbilityID{AbilityPowerStrike, AbilityBasicAttack}, 0, 0)
}

// BuildPlayerCombatSeed строит CombatUnitSeed игрока по статам и способностям (для persistent progression).
func BuildPlayerCombatSeed(maxHP, attack, defense, initiative int, abilities []AbilityID, healPower, basicAttackBonus int) CombatUnitSeed {
	if len(abilities) == 0 {
		abilities = []AbilityID{AbilityBasicAttack}
	}
	return CombatUnitSeed{
		Def: CombatUnitDefinition{
			ArchetypeID: "player",
			DisplayName: "Игрок",
			Role:        RoleFighter,
			Base: UnitBaseStats{
				MaxHP:            maxHP,
				Attack:           attack,
				Defense:          defense,
				Initiative:       initiative,
				HealPower:        healPower,
				BasicAttackBonus: basicAttackBonus,
			},
			IsRanged: false,
			Loadout:  AbilityLoadout{Abilities: abilities},
		},
		InitialHP: 0,
	}
}
