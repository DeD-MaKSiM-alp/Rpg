package battle

// ActionPreview is a lightweight, UI-friendly preview of an action outcome.
// It is intentionally approximate and non-authoritative.
type ActionPreview struct {
	DamageMin int
	DamageMax int
	HealMin   int
	HealMax   int
}

func (p ActionPreview) HasDamage() bool { return p.DamageMax > 0 }
func (p ActionPreview) HasHeal() bool   { return p.HealMax > 0 }

// PreviewAction returns an approximate preview for a validated action request.
// req.Actor must be the current acting unit (same as ActiveUnit when issuing a command).
// Groundwork: later this can incorporate statuses, crits, resistances, AoE etc.
func PreviewAction(ctx *BattleContext, req ActionRequest) (ActionPreview, ValidationResult) {
	v := ValidateAction(ctx, req)
	if !v.OK {
		return ActionPreview{}, v
	}

	actor := ctx.Units[req.Actor]
	ability := GetAbility(req.Ability)

	// For now we only preview unit-targeted abilities we already have.
	switch ability.ID {
	case AbilityBasicAttack, AbilityRangedAttack, AbilityPowerStrike:
		if req.Target.Kind != TargetKindUnit {
			return ActionPreview{}, okResult()
		}
		target := ctx.Units[req.Target.UnitID]
		if actor == nil || target == nil {
			return ActionPreview{}, okResult()
		}
		dmg := actor.Attack() - target.Defense()
		if dmg < 1 {
			dmg = 1
		}
		switch ability.ID {
		case AbilityRangedAttack:
			dmg += rangedShotBonus
		case AbilityPowerStrike:
			dmg += powerStrikeBonus
		}
		return ActionPreview{DamageMin: dmg, DamageMax: dmg}, okResult()

	case AbilityHeal:
		if actor == nil {
			return ActionPreview{}, okResult()
		}
		h := actor.HealPower()
		return ActionPreview{HealMin: h, HealMax: h}, okResult()

	case AbilityGroupHeal:
		if actor == nil {
			return ActionPreview{}, okResult()
		}
		g := actor.GroupHealPower()
		return ActionPreview{HealMin: g, HealMax: g}, okResult()

	default:
		return ActionPreview{}, okResult()
	}
}
