package battle

// ResourceProfile — боевой архетип ресурсов (мана/энергия/ритм). Не добавляет новый глобальный ресурс.
type ResourceProfile int

const (
	// ResourceProfileUnset — выводить профиль из роли и loadout (см. resourceProfileForDef).
	ResourceProfileUnset ResourceProfile = iota
	// ResourceProfileManaFocused — магия/поддержка: большой запас маны, мало энергии; мана ценится, энергия почти не растёт.
	ResourceProfileManaFocused
	// ResourceProfileEnergyFocused — физика/дальний бой: мало маны, много энергии; энергия быстрее восстанавливается.
	ResourceProfileEnergyFocused
	// ResourceProfileStriker — простой боец: умеренные пулы, равный regen; упор на базовый удар и КД спецов.
	ResourceProfileStriker
)

func resourceProfileFromRole(r Role) ResourceProfile {
	switch r {
	case RoleMage, RoleHealer:
		return ResourceProfileManaFocused
	case RoleArcher:
		return ResourceProfileEnergyFocused
	case RoleFighter:
		return ResourceProfileStriker
	default:
		return ResourceProfileStriker
	}
}

// inferBattleRoleFromAbilities — если в Def.Role ещё Fighter (legacy seed), уточняем роль по способностям.
func inferBattleRoleFromAbilities(abils []AbilityID) Role {
	var hasHeal, hasRanged, hasBuff bool
	for _, a := range abils {
		switch a {
		case AbilityHeal, AbilityGroupHeal:
			hasHeal = true
		case AbilityRangedAttack:
			hasRanged = true
		case AbilityBuff:
			hasBuff = true
		}
	}
	if hasHeal {
		return RoleHealer
	}
	if hasRanged {
		return RoleArcher
	}
	if hasBuff {
		return RoleMage
	}
	return RoleFighter
}

func resourceProfileForDef(def *CombatUnitDefinition) ResourceProfile {
	if def != nil && def.ResourceProfile != ResourceProfileUnset {
		return def.ResourceProfile
	}
	role := RoleFighter
	if def != nil {
		role = def.Role
		if role == RoleFighter {
			if inferred := inferBattleRoleFromAbilities(def.Loadout.Abilities); inferred != RoleFighter {
				role = inferred
			}
		}
	}
	return resourceProfileFromRole(role)
}

func applyResourceProfileToState(st *CombatUnitState, p ResourceProfile) {
	var maxM, maxE int
	switch p {
	case ResourceProfileManaFocused:
		maxM, maxE = 18, 4
	case ResourceProfileEnergyFocused:
		maxM, maxE = 4, 18
	case ResourceProfileStriker:
		maxM, maxE = 8, 10
	default:
		maxM, maxE = DefaultBattleManaMax, DefaultBattleEnergyMax
	}
	st.MaxMana = maxM
	st.Mana = maxM
	st.MaxEnergy = maxE
	st.Energy = maxE
	if st.AbilityCooldowns == nil {
		st.AbilityCooldowns = make(map[AbilityID]int)
	}
}

// regenManaEnergyPerRound — восстановление в конце полного раунда (tickRoundResources).
func regenManaEnergyPerRound(p ResourceProfile) (mana, energy int) {
	switch p {
	case ResourceProfileManaFocused:
		return 1, 0
	case ResourceProfileEnergyFocused:
		return 1, 2
	case ResourceProfileStriker:
		return 1, 1
	default:
		return 1, 1
	}
}

// ResourceProfileInspectLineRU — одна строка для inspect (кратко, без дублирования чисел из HUD боя).
func ResourceProfileInspectLineRU(role Role) string {
	switch resourceProfileFromRole(role) {
	case ResourceProfileManaFocused:
		return "Ресурсы боя: магический профиль · касты требуют маны"
	case ResourceProfileEnergyFocused:
		return "Ресурсы боя: энергетический профиль · частые техники и выстрелы"
	case ResourceProfileStriker:
		return "Ресурсы боя: ударник · базовый бой + мощный удар (КД)"
	default:
		return ""
	}
}
