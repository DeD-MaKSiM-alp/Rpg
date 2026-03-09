package battle

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Update обрабатывает один кадр боевого режима и возвращает итог.
func (b *BattleContext) Update() BattleAction {
	if b == nil {
		return BattleActionNone
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		b.Result = ResultEscape
		b.Phase = PhaseFinished
		b.LastLog = "Игрок покинул бой."
		return BattleActionRetreat
	}

	switch b.Phase {
	case PhaseStart:
		b.Phase = PhaseTurnStart
		return BattleActionNone

	case PhaseTurnStart:
		b.UpdateResultIfFinished()
		if b.IsFinished() {
			return b.ToBattleAction()
		}
		// Пропуск мёртвых в начале очереди
		for b.TurnIndex < len(b.TurnOrder) {
			u := b.ActiveUnit()
			if u != nil && u.IsAlive() {
				b.Phase = PhaseAwaitAction
				return BattleActionNone
			}
			b.TurnIndex++
		}
		b.Phase = PhaseRoundEnd
		return BattleActionNone

	case PhaseAwaitAction:
		active := b.ActiveUnit()
		if active == nil || !active.IsAlive() {
			b.Phase = PhaseTurnEnd
			return BattleActionNone
		}
		if active.Team == TeamPlayer {
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				b.SubmitPlayerAttack()
				b.Phase = PhaseTurnResolve
			}
		} else {
			b.ExecuteEnemyAutoAction()
			b.Phase = PhaseTurnResolve
		}
		return BattleActionNone

	case PhaseTurnResolve:
		b.UpdateResultIfFinished()
		if b.IsFinished() {
			return b.ToBattleAction()
		}
		b.Phase = PhaseTurnEnd
		return BattleActionNone

	case PhaseTurnEnd:
		if b.IsFinished() {
			return b.ToBattleAction()
		}
		b.AdvanceTurn()
		if b.IsFinished() {
			return b.ToBattleAction()
		}
		b.Phase = PhaseTurnStart
		return BattleActionNone

	case PhaseRoundEnd:
		b.Phase = PhaseTurnStart
		return BattleActionNone

	case PhaseFinished:
		return b.ToBattleAction()
	}

	return BattleActionNone
}

// SubmitPlayerAttack — базовая атака игрока (Space).
func (b *BattleContext) SubmitPlayerAttack() {
	attacker := b.ActiveUnit()
	if attacker == nil || attacker.Team != TeamPlayer {
		return
	}
	targets := b.LivingUnits(TeamEnemy)
	if len(targets) == 0 {
		return
	}
	target := targets[0]
	damage := attacker.Attack - target.Defense
	if damage < 1 {
		damage = 1
	}
	target.ApplyDamage(attacker.Attack)
	b.LastLog = fmt.Sprintf("Игрок атаковал %s на %d урона.", target.Name, damage)
}

// ExecuteEnemyAutoAction — враг атакует первую живую цель игрока.
func (b *BattleContext) ExecuteEnemyAutoAction() {
	attacker := b.ActiveUnit()
	if attacker == nil || attacker.Team != TeamEnemy {
		return
	}
	targets := b.LivingUnits(TeamPlayer)
	if len(targets) == 0 {
		return
	}
	target := targets[0]
	damage := attacker.Attack - target.Defense
	if damage < 1 {
		damage = 1
	}
	target.ApplyDamage(attacker.Attack)
	b.LastLog = fmt.Sprintf("%s атаковал игрока на %d урона.", attacker.Name, damage)
}
