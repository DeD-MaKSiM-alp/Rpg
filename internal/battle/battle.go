package battle

import (
	"fmt"
	"sort"
)

// BattleContext хранит состояние одного активного боя.
type BattleContext struct {
	Encounter Encounter

	Units     map[UnitID]*BattleUnit
	Sides     map[BattleSide]*BattleSideState
	TurnOrder []UnitID
	TurnIndex int
	Round     int

	Phase       Phase
	Result      Result
	PauseFrames int // кадры до перехода к следующей фазе

	// PlayerTurn holds player action selection state machine data.
	PlayerTurn PlayerTurnState

	LastMessage string
}

// BuildBattleContextFromEncounter создаёт BattleContext из Encounter.
func BuildBattleContextFromEncounter(enc Encounter) *BattleContext {
	if len(enc.Enemies) == 0 {
		return nil
	}
	ctx := &BattleContext{
		Encounter:   enc,
		Units:       make(map[UnitID]*BattleUnit),
		Sides:       make(map[BattleSide]*BattleSideState),
		Phase:       PhaseStart,
		Result:      ResultNone,
		LastMessage: "Бой начался.",
	}

	// Initialize spatial model (source of truth for placement).
	ctx.Sides[BattleSidePlayer] = NewBattleSideState(BattleSidePlayer)
	ctx.Sides[BattleSideEnemy] = NewBattleSideState(BattleSideEnemy)

	spawnUnit := func(id UnitID, side BattleSide, seed CombatUnitSeed) *BattleUnit {
		u := &BattleUnit{
			ID:     id,
			Side:   side,
			Def:    seed.Def,
			Origin: seed.Origin,
			State: CombatUnitState{
				HP:    seed.Def.Base.MaxHP,
				Alive: true,
			},
		}
		return u
	}

	// Player team (temporary: single unit, front row).
	// NOTE: canonical party composition will be introduced in the next steps.
	playerSeed := DefaultPlayerCombatUnitSeed()
	playerUnit := spawnUnit(UnitID(1), BattleSidePlayer, playerSeed)
	ctx.Units[playerUnit.ID] = playerUnit
	_ = ctx.PlaceUnit(playerUnit.ID, BattleSlotID{Side: BattleSidePlayer, Row: BattleRowFront, Index: 0})

	// Enemy team (first MaxFrontRowUnits in front, rest in back)
	for i, e := range enc.Enemies {
		if i >= MaxFrontRowUnits+MaxBackRowUnits {
			break
		}
		seed := BuildEnemyCombatUnitSeed(e)
		uid := UnitID(2 + i)
		u := spawnUnit(uid, BattleSideEnemy, seed)
		ctx.Units[uid] = u

		row := BattleRowFront
		index := i
		if i >= MaxFrontRowUnits {
			row = BattleRowBack
			index = i - MaxFrontRowUnits
		}
		_ = ctx.PlaceUnit(uid, BattleSlotID{Side: BattleSideEnemy, Row: row, Index: index})
	}

	ctx.TurnOrder = BuildTurnOrder(ctx)
	ctx.TurnIndex = 0
	ctx.Round = 1
	ctx.Phase = PhaseTurnStart

	return ctx
}

// BuildTurnOrder строит очередь ходов по Initiative (убыв.), при равной — по TeamID, Slot, UnitID.
func BuildTurnOrder(ctx *BattleContext) []UnitID {
	var live []*BattleUnit
	for _, u := range ctx.Units {
		if u.IsAlive() {
			live = append(live, u)
		}
	}
	sort.Slice(live, func(i, j int) bool {
		a, b := live[i], live[j]
		if a.Initiative() != b.Initiative() {
			return a.Initiative() > b.Initiative()
		}
		if a.Side != b.Side {
			return a.Side < b.Side
		}
		slotOrder := func(u *BattleUnit) int {
			sl := ctx.SlotByUnit(u.ID)
			if sl == nil {
				return 1_000_000
			}
			rowWeight := 0
			if sl.ID.Row == BattleRowBack {
				rowWeight = 1000
			}
			return rowWeight + sl.ID.Index
		}
		if slotOrder(a) != slotOrder(b) {
			return slotOrder(a) < slotOrder(b)
		}
		return a.ID < b.ID
	})
	out := make([]UnitID, 0, len(live))
	for _, u := range live {
		out = append(out, u.ID)
	}
	return out
}

// ActiveUnitID возвращает ID активного юнита (в PhaseTurnStart, PhaseAwaitAction, PhaseActionPause).
func (c *BattleContext) ActiveUnitID() UnitID {
	if c.Phase != PhaseTurnStart && c.Phase != PhaseAwaitAction && c.Phase != PhaseActionPause {
		return 0
	}
	if c.TurnIndex < 0 || c.TurnIndex >= len(c.TurnOrder) {
		return 0
	}
	return c.TurnOrder[c.TurnIndex]
}

// ActiveUnit возвращает активного юнита.
func (c *BattleContext) ActiveUnit() *BattleUnit {
	id := c.ActiveUnitID()
	if id == 0 {
		return nil
	}
	return c.Units[id]
}

// IsFinished возвращает true, если бой завершён (Result определён).
func (c *BattleContext) IsFinished() bool {
	return c.Result != ResultNone
}

// LivingUnits возвращает живых юнитов команды.
func (c *BattleContext) LivingUnits(team TeamID) []*BattleUnit {
	st := c.SideState(team)
	if st == nil {
		return nil
	}
	var out []*BattleUnit
	for i := range st.Slots {
		sl := &st.Slots[i]
		if sl.IsEmpty() {
			continue
		}
		u := c.Units[sl.Occupied]
		if u != nil && u.IsAlive() {
			out = append(out, u)
		}
	}
	return out
}

// TeamAlive возвращает true, если в команде есть живые юниты.
func (c *BattleContext) TeamAlive(team TeamID) bool {
	return len(c.LivingUnits(team)) > 0
}

// UpdateResultIfFinished проверяет конец боя и выставляет Result (Phase не меняет).
func (c *BattleContext) UpdateResultIfFinished() {
	if !c.TeamAlive(TeamPlayer) {
		c.Result = ResultDefeat
		c.LastMessage = "Поражение."
		return
	}
	if !c.TeamAlive(TeamEnemy) {
		c.Result = ResultVictory
		c.LastMessage = "Победа!"
		return
	}
}

// AdvanceTurn переходит к следующему живому юниту в очереди.
func (c *BattleContext) AdvanceTurn() {
	c.TurnIndex++
	for c.TurnIndex < len(c.TurnOrder) {
		id := c.TurnOrder[c.TurnIndex]
		u := c.Units[id]
		if u != nil && u.IsAlive() {
			return
		}
		c.TurnIndex++
	}

	// Конец раунда
	c.Phase = PhaseRoundEnd
	c.Round++
	c.TurnOrder = BuildTurnOrder(c)
	c.TurnIndex = 0
	c.UpdateResultIfFinished()
	if c.IsFinished() {
		return
	}
	c.Phase = PhaseTurnStart
}

// DisplayPhaseLabel возвращает текст фазы для UI (ход игрока / ход врага / завершён).
func (c *BattleContext) DisplayPhaseLabel() string {
	if c.Phase == PhaseFinishedWaitInput || c.Result != ResultNone {
		return "Бой завершён"
	}
	u := c.ActiveUnit()
	if u == nil {
		return "Бой завершён"
	}
	if u.Side == TeamPlayer {
		return ">>> ХОД ИГРОКА <<<"
	}
	return ">>> ХОД ВРАГА <<<"
}

// TeamFirstHP возвращает HP первого живого юнита команды (для UI).
func (c *BattleContext) TeamFirstHP(team TeamID) int {
	live := c.LivingUnits(team)
	if len(live) == 0 {
		return 0
	}
	return live[0].State.HP
}

// CanPlayerActNow возвращает true, если сейчас ход игрока и можно выполнить действие.
func (c *BattleContext) CanPlayerActNow() bool {
	if c.Phase != PhaseAwaitAction || c.Result != ResultNone {
		return false
	}
	u := c.ActiveUnit()
	return u != nil && u.IsAlive() && u.Side == TeamPlayer
}

// ActiveUnitName возвращает имя активного юнита для UI.
func (c *BattleContext) ActiveUnitName() string {
	u := c.ActiveUnit()
	if u == nil {
		return "-"
	}
	return u.Name()
}

// ActiveUnitTeamName возвращает "Player" или "Enemy" для активного юнита.
func (c *BattleContext) ActiveUnitTeamName() string {
	u := c.ActiveUnit()
	if u == nil {
		return "-"
	}
	if u.Side == TeamPlayer {
		return "Player"
	}
	return "Enemy"
}

// PhaseString возвращает строковое представление фазы для debug.
func (c *BattleContext) PhaseString() string {
	switch c.Phase {
	case PhaseStart:
		return "Start"
	case PhaseTurnStart:
		return "TurnStart"
	case PhaseAwaitAction:
		return "AwaitAction"
	case PhaseActionPause:
		return "ActionPause"
	case PhaseTurnEnd:
		return "TurnEnd"
	case PhaseRoundEnd:
		return "RoundEnd"
	case PhaseFinishedWaitInput:
		return "FinishedWaitInput"
	}
	return "?"
}

// ResultString возвращает строковое представление результата боя.
func (c *BattleContext) ResultString() string {
	switch c.Result {
	case ResultVictory:
		return "Victory"
	case ResultDefeat:
		return "Defeat"
	case ResultEscape:
		return "Escape"
	}
	return "-"
}

// FormationSummary возвращает краткое описание построения для UI.
func (c *BattleContext) FormationSummary() string {
	pf := len(c.LivingUnitsInRow(TeamPlayer, RowFront))
	pb := len(c.LivingUnitsInRow(TeamPlayer, RowBack))
	ef := len(c.LivingUnitsInRow(TeamEnemy, RowFront))
	eb := len(c.LivingUnitsInRow(TeamEnemy, RowBack))
	return fmt.Sprintf("Player: %d front, %d back | Enemy: %d front, %d back", pf, pb, ef, eb)
}

// ToBattleOutcome возвращает BattleOutcome для Game (Result -> внешний API).
func (c *BattleContext) ToBattleOutcome() BattleOutcome {
	switch c.Result {
	case ResultVictory:
		return BattleOutcomeVictory
	case ResultDefeat:
		return BattleOutcomeDefeat
	case ResultEscape:
		return BattleOutcomeRetreat
	}
	return BattleOutcomeNone
}

// ApplyActionResult вызывается после ResolveAbility: логи/анимации и UpdateResultIfFinished.
func (c *BattleContext) ApplyActionResult(r ActionResult) {
	if r.Actor == 0 {
		return
	}
	actor := c.Units[r.Actor]
	target := c.Units[r.Target]
	if actor != nil && target != nil {
		if r.HealAmount > 0 {
			c.LastMessage = fmt.Sprintf("%s вылечил %s на %d.", actor.Name(), target.Name(), r.HealAmount)
		} else if r.Damage > 0 {
			c.LastMessage = logActionResult(actor, target, r)
		} else if r.Target != 0 {
			c.LastMessage = fmt.Sprintf("%s усилил %s.", actor.Name(), target.Name())
		}
	}
	c.UpdateResultIfFinished()
}

func logActionResult(actor, target *BattleUnit, r ActionResult) string {
	if r.Killed {
		return actor.Name() + " победил " + target.Name() + "."
	}
	return formatDamageLog(actor.Name(), target.Name(), r.Damage)
}

func formatDamageLog(actorName, targetName string, damage int) string {
	if damage <= 0 {
		return actorName + " атаковал " + targetName + "."
	}
	return fmt.Sprintf("%s атаковал %s на %d урона.", actorName, targetName, damage)
}
