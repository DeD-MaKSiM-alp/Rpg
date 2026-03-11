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
	case AbilityBasicAttack, AbilityRangedAttack:
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
		return ActionPreview{DamageMin: dmg, DamageMax: dmg}, okResult()

	case AbilityHeal:
		// healAmount is the current battle constant; statuses/bonuses will come later.
		return ActionPreview{HealMin: healAmount, HealMax: healAmount}, okResult()

	default:
		return ActionPreview{}, okResult()
	}
}

