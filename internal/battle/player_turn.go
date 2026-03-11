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
	PlayerConfirmAction
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

	Pending ActionRequest
}

func (p *PlayerTurnState) Reset() {
	*p = PlayerTurnState{}
}

func (p *PlayerTurnState) IsActiveFor(actor UnitID) bool {
	return p != nil && p.Phase != PlayerTurnNone && p.Actor == actor && actor != 0
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
	case PlayerConfirmAction:
		return "Confirm"
	case PlayerResolveAction:
		return "Resolve"
	case PlayerTurnEnd:
		return "End"
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

