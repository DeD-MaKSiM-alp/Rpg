package battle

import (
	"testing"

	"mygame/world/entity"
)

func TestResourceProfileForDef_manaFromHealerRole(t *testing.T) {
	def := CombatUnitDefinition{
		Role:   RoleHealer,
		Loadout: AbilityLoadout{Abilities: []AbilityID{AbilityHeal, AbilityBasicAttack}},
	}
	if p := resourceProfileForDef(&def); p != ResourceProfileManaFocused {
		t.Fatalf("want ManaFocused, got %v", p)
	}
}

func TestResourceProfileForDef_inferFromAbilitiesWhenFighter(t *testing.T) {
	def := CombatUnitDefinition{
		Role:   RoleFighter,
		Loadout: AbilityLoadout{Abilities: []AbilityID{AbilityGroupHeal, AbilityBasicAttack}},
	}
	if p := resourceProfileForDef(&def); p != ResourceProfileManaFocused {
		t.Fatalf("infer healer loadout: want ManaFocused, got %v", p)
	}
}

func TestResourceProfileForDef_energyFromArcher(t *testing.T) {
	def := CombatUnitDefinition{
		Role:   RoleArcher,
		Loadout: AbilityLoadout{Abilities: []AbilityID{AbilityRangedAttack, AbilityBasicAttack}},
	}
	if p := resourceProfileForDef(&def); p != ResourceProfileEnergyFocused {
		t.Fatalf("want EnergyFocused, got %v", p)
	}
}

func TestInitCombatResources_appliesPools(t *testing.T) {
	st := CombatUnitState{Alive: true}
	def := CombatUnitDefinition{Role: RoleMage, Loadout: AbilityLoadout{Abilities: []AbilityID{AbilityBuff, AbilityBasicAttack}}}
	initCombatResources(&st, &def)
	if st.MaxMana != 18 || st.MaxEnergy != 4 {
		t.Fatalf("mage pools: MaxMana=%d MaxEnergy=%d", st.MaxMana, st.MaxEnergy)
	}
}

func TestTickRoundResources_manaVsEnergyRegen(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	mageSeed := BuildPlayerCombatSeed(10, 1, 0, 2, []AbilityID{AbilityBuff, AbilityBasicAttack}, 0, 0)
	mageSeed.Def.Role = RoleMage
	archSeed := BuildPlayerCombatSeed(10, 1, 0, 2, []AbilityID{AbilityRangedAttack, AbilityBasicAttack}, 0, 0)
	archSeed.Def.Role = RoleArcher
	ctx := BuildBattleContextFromEncounter(enc, []CombatUnitSeed{mageSeed, archSeed}, 0)
	var mageU, archU *CombatUnit
	for _, u := range ctx.Units {
		if u == nil || u.Side != TeamPlayer {
			continue
		}
		if u.Def.Role == RoleMage {
			mageU = u
		}
		if u.Def.Role == RoleArcher {
			archU = u
		}
	}
	if mageU == nil || archU == nil {
		t.Fatal("missing units")
	}
	mageU.State.Mana = 0
	mageU.State.Energy = 0
	archU.State.Mana = 0
	archU.State.Energy = 0
	ctx.tickRoundResources()
	if mageU.State.Mana != 1 || mageU.State.Energy != 0 {
		t.Fatalf("mana profile regen: mana=%d energy=%d", mageU.State.Mana, mageU.State.Energy)
	}
	if archU.State.Energy != 2 {
		t.Fatalf("energy profile: want energy 2, got %d", archU.State.Energy)
	}
	if archU.State.Mana != 1 {
		t.Fatalf("energy profile mana trickle: got %d", archU.State.Mana)
	}
}
