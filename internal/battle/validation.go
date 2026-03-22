package battle

import "fmt"

// TargetKind describes how an action targets.
type TargetKind int

const (
	TargetKindNone TargetKind = iota
	TargetKindUnit
	TargetKindSelf
)

// TargetDescriptor is a normalized target reference for validation/UI/AI.
// Groundwork: later this can be extended to Slot/Group targets without changing ActionRequest shape.
type TargetDescriptor struct {
	Kind   TargetKind
	UnitID UnitID // used when Kind == TargetKindUnit
}

func NoTarget() TargetDescriptor           { return TargetDescriptor{Kind: TargetKindNone} }
func SelfTarget() TargetDescriptor         { return TargetDescriptor{Kind: TargetKindSelf} }
func UnitTarget(id UnitID) TargetDescriptor { return TargetDescriptor{Kind: TargetKindUnit, UnitID: id} }

// ActionRequest is an explicit request to perform an ability with a target descriptor.
type ActionRequest struct {
	Actor   UnitID
	Ability AbilityID
	Target  TargetDescriptor
}

// ValidationCode is a stable machine-readable reason.
type ValidationCode int

const (
	Valid ValidationCode = iota
	ErrNoActor
	ErrActorDead
	ErrActorDisabled
	ErrNoAbility
	ErrAbilityUnavailable
	ErrMissingTarget
	ErrUnexpectedTarget
	ErrNoSuchTargetUnit
	ErrTargetDead
	ErrTargetSideMismatch
	ErrMeleeScreening
	ErrInsufficientMana
	ErrInsufficientEnergy
	ErrAbilityOnCooldown
)

// ValidationResult is a readable outcome of validation.
type ValidationResult struct {
	OK      bool
	Code    ValidationCode
	Message string
}

func okResult() ValidationResult { return ValidationResult{OK: true, Code: Valid} }

func errResult(code ValidationCode, format string, args ...any) ValidationResult {
	return ValidationResult{OK: false, Code: code, Message: fmt.Sprintf(format, args...)}
}

// ListValidTargets enumerates all valid targets for actor+ability in the current battle state.
func ListValidTargets(ctx *BattleContext, actorID UnitID, abilityID AbilityID) ([]TargetDescriptor, ValidationResult) {
	if ctx == nil || ctx.Units == nil {
		return nil, errResult(ErrNoActor, "no battle context")
	}
	actor := ctx.Units[actorID]
	if actor == nil {
		return nil, errResult(ErrNoActor, "actor not found")
	}
	if !actor.IsAlive() {
		return nil, errResult(ErrActorDead, "actor is dead")
	}
	if actor.State.Disabled {
		return nil, errResult(ErrActorDisabled, "actor is disabled")
	}

	// Ability must exist and be available to actor.
	ability, ok := abilityRegistry[abilityID]
	if !ok {
		return nil, errResult(ErrNoAbility, "unknown ability")
	}
	hasAbility := false
	for _, id := range actor.Abilities() {
		if id == abilityID {
			hasAbility = true
			break
		}
	}
	if !hasAbility {
		return nil, errResult(ErrAbilityUnavailable, "ability not available for actor")
	}
	if res := AbilityResourceGate(ctx, actor, abilityID); !res.OK {
		return nil, res
	}

	switch ability.TargetRule {
	case TargetSelf:
		return []TargetDescriptor{SelfTarget()}, okResult()

	case TargetAllyTeam:
		allies := ctx.LivingUnits(actor.Side)
		if len(allies) == 0 {
			return nil, errResult(ErrNoSuchTargetUnit, "no allies")
		}
		return []TargetDescriptor{NoTarget()}, okResult()

	case TargetAllySingle:
		allies := ctx.LivingUnits(actor.Side)
		out := make([]TargetDescriptor, 0, len(allies))
		for _, u := range allies {
			out = append(out, UnitTarget(u.ID))
		}
		return out, okResult()

	case TargetEnemySingle:
		enemySide := ctx.EnemyTeam(actor.Side)
		allEnemies := ctx.LivingUnits(enemySide)
		if len(allEnemies) == 0 {
			return nil, okResult()
		}

		rng := effectiveRange(actor, ability)
		if rng == RangeRanged {
			out := make([]TargetDescriptor, 0, len(allEnemies))
			for _, u := range allEnemies {
				out = append(out, UnitTarget(u.ID))
			}
			return out, okResult()
		}

		// Melee screening: if enemy front row has living units, only front row is targetable.
		var rowTargets []*BattleUnit
		if ctx.FrontRowAlive(enemySide) {
			rowTargets = ctx.LivingUnitsInRow(enemySide, RowFront)
		} else {
			rowTargets = ctx.LivingUnitsInRow(enemySide, RowBack)
		}
		out := make([]TargetDescriptor, 0, len(rowTargets))
		for _, u := range rowTargets {
			out = append(out, UnitTarget(u.ID))
		}
		return out, okResult()

	default:
		// Groundwork for future TargetNone etc. We still provide a safe behavior.
		return nil, okResult()
	}
}

// ValidateAction validates a concrete action request (actor, ability, target).
func ValidateAction(ctx *BattleContext, req ActionRequest) ValidationResult {
	targets, base := ListValidTargets(ctx, req.Actor, req.Ability)
	if !base.OK {
		return base
	}

	ability := abilityRegistry[req.Ability]
	actor := ctx.Units[req.Actor]

	switch ability.TargetRule {
	case TargetSelf:
		if req.Target.Kind != TargetKindSelf {
			return errResult(ErrMissingTarget, "self-target required")
		}
		return okResult()

	case TargetAllyTeam:
		if req.Target.Kind != TargetKindNone {
			return errResult(ErrUnexpectedTarget, "mass heal does not take a target")
		}
		return okResult()

	case TargetAllySingle, TargetEnemySingle:
		if req.Target.Kind != TargetKindUnit || req.Target.UnitID == 0 {
			return errResult(ErrMissingTarget, "unit target required")
		}
		target := ctx.Units[req.Target.UnitID]
		if target == nil {
			return errResult(ErrNoSuchTargetUnit, "target unit not found")
		}
		if !target.IsAlive() {
			return errResult(ErrTargetDead, "target is dead")
		}
		// Side rule check (defensive; enumeration should already enforce it).
		if ability.TargetRule == TargetAllySingle && target.Side != actor.Side {
			return errResult(ErrTargetSideMismatch, "ally target must be on same side")
		}
		if ability.TargetRule == TargetEnemySingle && target.Side == actor.Side {
			return errResult(ErrTargetSideMismatch, "enemy target must be on opposite side")
		}
		// Centralized membership check (covers melee screening too).
		for _, td := range targets {
			if td.Kind == TargetKindUnit && td.UnitID == req.Target.UnitID {
				return okResult()
			}
		}
		if ability.TargetRule == TargetEnemySingle && effectiveRange(actor, ability) == RangeMelee {
			return errResult(ErrMeleeScreening, "melee cannot bypass living enemy front row")
		}
		return errResult(ErrUnexpectedTarget, "target is not valid for this action")

	default:
		// Unknown/unsupported rule: be strict for concrete requests.
		if req.Target.Kind != TargetKindNone {
			return errResult(ErrUnexpectedTarget, "unexpected target for this ability")
		}
		return okResult()
	}
}

// ToBattleAction converts an ActionRequest into current BattleAction representation.
// NOTE: BattleAction currently only supports unit targets; self is normalized to actor; none -> Target=0.
func ToBattleAction(ctx *BattleContext, req ActionRequest) (BattleAction, ValidationResult) {
	v := ValidateAction(ctx, req)
	if !v.OK {
		return BattleAction{}, v
	}
	targetID := UnitID(0)
	switch req.Target.Kind {
	case TargetKindNone:
		targetID = 0
	case TargetKindSelf:
		targetID = req.Actor
	case TargetKindUnit:
		targetID = req.Target.UnitID
	}
	return BattleAction{Actor: req.Actor, Ability: req.Ability, Target: targetID}, okResult()
}

