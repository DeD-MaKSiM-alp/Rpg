package battle

import (
	"fmt"
)

// PlayerTurnPhase is a sub-phase state machine for player's active unit turn.
type PlayerTurnPhase int

const (
	PlayerTurnNone PlayerTurnPhase = iota
	PlayerTurnStart
	PlayerChooseAbility
	PlayerChooseTarget
	PlayerResolveAction
	PlayerTurnEnd
)

type PlayerTurnState struct {
	Phase PlayerTurnPhase

	Actor UnitID

	SelectedAbilityIndex int
	SelectedAbilityID    AbilityID

	ValidTargets      []TargetDescriptor
	SelectedTargetIdx int
	SelectedTarget    TargetDescriptor

	// UI/UX-only fields (hover highlights); do not affect rules.
	HoverAbilityIndex int  // -1 if none
	HoverTargetUnitID UnitID

	HoverBackButton bool

	Pending ActionRequest
}

func (p *PlayerTurnState) Reset() {
	*p = PlayerTurnState{}
}

func (p *PlayerTurnState) IsActiveFor(actor UnitID) bool {
	return p != nil && p.Phase != PlayerTurnNone && p.Actor == actor && actor != 0
}

// playerTurnSelectAbility sets selected ability id/index to match the given ability in actor loadout.
func playerTurnSelectAbility(actor *BattleUnit, p *PlayerTurnState, id AbilityID) {
	if actor == nil || p == nil {
		return
	}
	abs := actor.Abilities()
	for i := range abs {
		if abs[i] == id {
			p.SelectedAbilityIndex = i
			p.SelectedAbilityID = id
			return
		}
	}
	if len(abs) > 0 {
		p.SelectedAbilityIndex = 0
		p.SelectedAbilityID = abs[0]
	}
}

// playerTurnResetToBasicAttack clears target selection and selects basic attack in loadout order (if present).
func playerTurnResetToBasicAttack(actor *BattleUnit, p *PlayerTurnState) {
	if p == nil || !HasBasicAttack(actor) {
		return
	}
	playerTurnSelectAbility(actor, p, AbilityBasicAttack)
	p.ValidTargets = nil
	p.SelectedTargetIdx = 0
	p.SelectedTarget = NoTarget()
	p.Pending = ActionRequest{}
}

// playerTurnTrySpecialAbilityClick mirrors a left-click on one special-ability row: validates first,
// then commits SelectedAbilityID / phase. Used by HUD mouse and unit tests (no Ebiten).
func playerTurnTrySpecialAbilityClick(b *BattleContext, actor *BattleUnit, abilID AbilityID) (BattleAction, bool) {
	if b == nil || actor == nil {
		return BattleAction{}, false
	}
	pt := &b.PlayerTurn
	ability := GetAbility(abilID)
	switch ability.TargetRule {
	case TargetEnemySingle, TargetAllySingle:
		targets, v := ListValidTargets(b, actor.ID, abilID)
		if !v.OK {
			b.AddBattleLog(v.Message)
			return BattleAction{}, false
		}
		if len(targets) == 0 {
			b.AddBattleLog("Нет валидных целей.")
			return BattleAction{}, false
		}
		playerTurnSelectAbility(actor, pt, abilID)
		pt.ValidTargets = targets
		pt.SelectedTargetIdx = 0
		pt.SelectedTarget = targets[0]
		pt.Pending = ActionRequest{}
		pt.Phase = PlayerChooseTarget
		return BattleAction{}, false

	case TargetSelf:
		selfT := SelfTarget()
		req := ActionRequest{Actor: actor.ID, Ability: abilID, Target: selfT}
		if v := ValidateAction(b, req); !v.OK {
			b.AddBattleLog(v.Message)
			return BattleAction{}, false
		}
		act, v2 := ToBattleAction(b, req)
		if !v2.OK {
			b.AddBattleLog(v2.Message)
			return BattleAction{}, false
		}
		playerTurnSelectAbility(actor, pt, abilID)
		pt.SelectedTarget = selfT
		pt.Phase = PlayerResolveAction
		return act, true

	case TargetAllyTeam:
		fallthrough
	default:
		noneT := NoTarget()
		req := ActionRequest{Actor: actor.ID, Ability: abilID, Target: noneT}
		if v := ValidateAction(b, req); !v.OK {
			b.AddBattleLog(v.Message)
			return BattleAction{}, false
		}
		act, v2 := ToBattleAction(b, req)
		if !v2.OK {
			b.AddBattleLog(v2.Message)
			return BattleAction{}, false
		}
		playerTurnSelectAbility(actor, pt, abilID)
		pt.SelectedTarget = noneT
		pt.Phase = PlayerResolveAction
		return act, true
	}
}

func (p *PlayerTurnState) PhaseString() string {
	switch p.Phase {
	case PlayerTurnNone:
		return "-"
	case PlayerTurnStart:
		return "Start"
	case PlayerChooseAbility:
		return "ChooseAbility"
	case PlayerChooseTarget:
		return "ChooseTarget"
	case PlayerResolveAction:
		return "Resolve"
	case PlayerTurnEnd:
		return "End"
	default:
		return "?"
	}
}

// PhaseLabelRU — подпись подфазы хода игрока для player-facing HUD.
func (p *PlayerTurnState) PhaseLabelRU() string {
	if p == nil {
		return "—"
	}
	switch p.Phase {
	case PlayerTurnNone:
		return "—"
	case PlayerTurnStart:
		return "старт"
	case PlayerChooseAbility:
		return "способность"
	case PlayerChooseTarget:
		return "цель"
	case PlayerResolveAction:
		return "выполнение"
	case PlayerTurnEnd:
		return "конец"
	default:
		return "?"
	}
}

func (c *BattleContext) PlayerTurnStatusString() string {
	if c == nil {
		return "-"
	}
	if c.Phase != PhaseAwaitAction {
		return "-"
	}
	u := c.ActiveUnit()
	if u == nil || u.Side != TeamPlayer || !u.IsAlive() {
		return "-"
	}
	p := &c.PlayerTurn
	if !p.IsActiveFor(u.ID) {
		return "player_turn: (not initialized)"
	}
	target := "-"
	if p.SelectedTarget.Kind == TargetKindSelf {
		target = "self"
	} else if p.SelectedTarget.Kind == TargetKindUnit {
		if tu := c.Units[p.SelectedTarget.UnitID]; tu != nil {
			target = fmt.Sprintf("%s(#%d)", tu.Name(), tu.ID)
		} else {
			target = fmt.Sprintf("unit#%d", p.SelectedTarget.UnitID)
		}
	} else if p.SelectedTarget.Kind == TargetKindNone {
		target = "none"
	}
	return fmt.Sprintf("player_turn:%s abil:%d idx:%d targets:%d tIdx:%d t:%s",
		p.PhaseString(), p.SelectedAbilityID, p.SelectedAbilityIndex, len(p.ValidTargets), p.SelectedTargetIdx, target)
}

