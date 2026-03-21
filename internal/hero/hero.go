// Package hero holds one combat-capable character's persistent stats between battles.
// World position lives in player.Player; the roster lives in party.Party; battle runtime lives in battle.CombatUnit.
// Progression mutates hero.Hero; CombatUnitSeed() projects into battle seeds.
// Выбор награды после победы по-прежнему к лидеру; общий боевой опыт — у выживших активных (party-wide).
package hero

import (
	battlepkg "mygame/internal/battle"
	"mygame/internal/unitdata"
)

// Hero — состояние одного бойца между боями (статы, способности). Сборка отряда — в party.Party.
type Hero struct {
	// UnitID — стабильный id шаблона юнита (internal/unitdata); не display name. Пустой = legacy до data layer.
	UnitID string
	MaxHP            int
	CurrentHP        int // каноническое HP между боями; 0 = недоступен для следующего боя (пока нет лечения/лагеря)
	Attack           int
	Defense          int
	Initiative       int
	HealPower        int // bonus HP healed (added to base 2); see battle.CombatUnit.HealPower
	BasicAttackBonus int // extra damage for basic attack only (награды лидера и т.п.)
	// CombatExperience — party-wide progression: накапливается за победы, в которых герой был в активном строю и выжил.
	// В бой идёт как бонус к базовой атаке: +1 за каждые CombatXPStepsPerBasicAttackBonus (см. CombatUnitSeed).
	CombatExperience int
	Abilities        []battlepkg.AbilityID
	// RecruitLabel — если не пусто, подпись в UI (например «Новобранец 2»); иначе используются роли party.
	RecruitLabel string
}

// CombatXPStepsPerBasicAttackBonus — каждые N единиц боевого опыта дают +1 к эффективному бонусу базовой атаки в бою.
const CombatXPStepsPerBasicAttackBonus = 4

// DefaultHero возвращает стартового героя из unit template (лидер милитии).
func DefaultHero() Hero {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireMilitiaSpearmanT1)
	if err != nil {
		// LEGACY fallback: статический профиль до registry (не должно срабатывать в норме).
		h = Hero{
			MaxHP:      10,
			Attack:     2,
			Defense:    0,
			Initiative: 2,
			HealPower:  0,
			Abilities:  []battlepkg.AbilityID{battlepkg.AbilityBasicAttack},
		}
		h.CurrentHP = h.MaxHP
	}
	return h
}

// NewHeroFromUnitTemplate создаёт героя из реестра шаблонов (UnitID + стартовые статы / способности).
func NewHeroFromUnitTemplate(unitID string) (Hero, error) {
	t, err := unitdata.MustGetUnitTemplate(unitID)
	if err != nil {
		return Hero{}, err
	}
	abils := make([]battlepkg.AbilityID, len(t.Abilities))
	copy(abils, t.Abilities)
	h := Hero{
		UnitID:       t.UnitID,
		MaxHP:        t.MaxHP,
		Attack:       t.Attack,
		Defense:      t.Defense,
		Initiative:   t.Initiative,
		HealPower:    t.HealPower,
		BasicAttackBonus: 0,
		CombatExperience: 0,
		Abilities:    abils,
	}
	h.CurrentHP = h.MaxHP
	return h, nil
}

// CanEnterBattle true, если герой может получить сид для боя (есть HP).
func (h *Hero) CanEnterBattle() bool {
	return h != nil && h.CurrentHP > 0
}

// EffectiveBasicAttackBonusForCombat — сумма наград (BasicAttackBonus) и роста от боевого опыта.
func (h *Hero) EffectiveBasicAttackBonusForCombat() int {
	if h == nil {
		return 0
	}
	return h.BasicAttackBonus + h.CombatExperience/CombatXPStepsPerBasicAttackBonus
}

// battleRoleFromAbilities — LEGACY: роль из способностей, если нет канонического шаблона (hero.UnitID).
func battleRoleFromAbilities(abils []battlepkg.AbilityID) battlepkg.Role {
	hasHeal := false
	hasRanged := false
	for _, a := range abils {
		switch a {
		case battlepkg.AbilityHeal, battlepkg.AbilityGroupHeal:
			hasHeal = true
		case battlepkg.AbilityRangedAttack:
			hasRanged = true
		}
	}
	if hasHeal {
		return battlepkg.RoleHealer
	}
	if hasRanged {
		return battlepkg.RoleArcher
	}
	return battlepkg.RoleFighter
}

func mapUnitdataAttackKind(k unitdata.AttackKind) battlepkg.TemplateAttackKind {
	switch k {
	case unitdata.AttackMelee:
		return battlepkg.TemplateAttackMelee
	case unitdata.AttackRanged:
		return battlepkg.TemplateAttackRanged
	case unitdata.AttackHeal:
		return battlepkg.TemplateAttackHeal
	default:
		return battlepkg.TemplateAttackUnknown
	}
}

// combatIsRangedFromRole — дальность «как лучник» по канонической роли (боевые правила).
func combatIsRangedFromRole(role battlepkg.Role) bool {
	return role == battlepkg.RoleArcher
}

// CombatUnitSeed строит один combat seed; party.Party.PlayerCombatSeeds() собирает сиды всего активного ростера.
func (h *Hero) CombatUnitSeed() battlepkg.CombatUnitSeed {
	s := battlepkg.BuildPlayerCombatSeed(
		h.MaxHP,
		h.Attack,
		h.Defense,
		h.Initiative,
		h.Abilities,
		h.HealPower,
		h.EffectiveBasicAttackBonusForCombat(),
	)
	if tpl, ok := unitdata.GetUnitTemplate(h.UnitID); ok {
		s.Def.TemplateUnitID = tpl.UnitID
		s.Def.FactionID = tpl.FactionID
		s.Def.LineID = tpl.LineID
		s.Def.Tier = tpl.Tier
		s.Def.ArchetypeID = tpl.ArchetypeID
		s.Def.Role = tpl.Role
		s.Def.IdentityAttackKind = mapUnitdataAttackKind(tpl.AttackKind)
		s.Def.IsRanged = combatIsRangedFromRole(tpl.Role)
	} else {
		// LEGACY: пустой или неизвестный UnitID — роль из способностей.
		s.Def.Role = battleRoleFromAbilities(h.Abilities)
		s.Def.IsRanged = combatIsRangedFromRole(s.Def.Role)
		s.Def.IdentityAttackKind = battlepkg.TemplateAttackUnknown
	}
	if h.CurrentHP > 0 {
		s.InitialHP = h.CurrentHP
		if s.InitialHP > h.MaxHP {
			s.InitialHP = h.MaxHP
		}
	}
	return s
}
