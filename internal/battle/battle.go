package battle

import (
	"fmt"
	"sort"

	"mygame/world"
)

// BattleContext хранит состояние одного активного боя.
type BattleContext struct {
	Encounter Encounter

	Units     map[UnitID]*BattleUnit
	Teams     map[TeamID]*BattleTeam
	TurnOrder []UnitID
	TurnIndex int
	Round     int

	Phase      Phase
	Result     Result
	PhaseTimer int // кадры до перехода к следующей фазе

	LastLog string

	// EnemyID — удобный доступ к SourceEnemyID (для UI/legacy).
	EnemyID world.EntityID
}

// BuildBattleContextFromEncounter создаёт BattleContext из Encounter.
func BuildBattleContextFromEncounter(enc Encounter) *BattleContext {
	if len(enc.Enemies) == 0 {
		return nil
	}
	ctx := &BattleContext{
		Encounter: enc,
		EnemyID:   enc.SourceEnemyID,
		Units:     make(map[UnitID]*BattleUnit),
		Teams:     make(map[TeamID]*BattleTeam),
		Phase:     PhaseStart,
		Result:    ResultNone,
		LastLog:   "Бой начался.",
	}

	// Player team (player front row)
	playerUnit := &BattleUnit{
		ID:         UnitID(1),
		Name:       "Игрок",
		Team:       TeamPlayer,
		Slot:       0,
		Row:        RowFront,
		MaxHP:      10,
		HP:         10,
		Attack:     2,
		Defense:    0,
		Initiative: 2,
		Alive:      true,
		Ranged:     false,
		Abilities:  []AbilityID{AbilityBasicAttack},
	}
	ctx.Units[playerUnit.ID] = playerUnit
	ctx.Teams[TeamPlayer] = &BattleTeam{ID: TeamPlayer, Units: []UnitID{playerUnit.ID}}

	// Enemy team (first MaxFrontRowUnits in front, rest in back)
	enemyUnits := make([]UnitID, 0, len(enc.Enemies))
	for i, e := range enc.Enemies {
		seed := BuildBattleUnitSeed(e)
		uid := UnitID(2 + i)
		row := RowFront
		if i >= MaxFrontRowUnits {
			row = RowBack
		}
		abil := AbilityBasicAttack
		if seed.IsRanged {
			abil = AbilityRangedAttack
		}
		u := &BattleUnit{
			ID:             uid,
			Name:           seed.Name,
			Team:           TeamEnemy,
			Slot:           i,
			Row:            row,
			MaxHP:          seed.MaxHP,
			HP:             seed.MaxHP,
			Attack:         seed.Attack,
			Defense:        seed.Defense,
			Initiative:     seed.Initiative,
			Alive:          true,
			Ranged:         seed.IsRanged,
			Abilities:      []AbilityID{abil},
			SourceEnemyID:  seed.SourceEnemyID,
		}
		ctx.Units[uid] = u
		enemyUnits = append(enemyUnits, uid)
	}
	ctx.Teams[TeamEnemy] = &BattleTeam{ID: TeamEnemy, Units: enemyUnits}

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
		if a.Initiative != b.Initiative {
			return a.Initiative > b.Initiative
		}
		if a.Team != b.Team {
			return a.Team < b.Team
		}
		if a.Slot != b.Slot {
			return a.Slot < b.Slot
		}
		return a.ID < b.ID
	})
	out := make([]UnitID, 0, len(live))
	for _, u := range live {
		out = append(out, u.ID)
	}
	return out
}

// ActiveUnitID возвращает ID активного юнита (в PhaseTurnStart, PhaseAwaitAction, PhaseTurnResolve, PhaseActionPause).
func (c *BattleContext) ActiveUnitID() UnitID {
	if c.Phase != PhaseTurnStart && c.Phase != PhaseAwaitAction && c.Phase != PhaseTurnResolve && c.Phase != PhaseActionPause {
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

// IsFinished возвращает true, если бой завершён.
func (c *BattleContext) IsFinished() bool {
	return c.Phase == PhaseFinished || c.Result != ResultNone
}

// LivingUnits возвращает живых юнитов команды.
func (c *BattleContext) LivingUnits(team TeamID) []*BattleUnit {
	t := c.Teams[team]
	if t == nil {
		return nil
	}
	var out []*BattleUnit
	for _, id := range t.Units {
		u := c.Units[id]
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
		c.LastLog = "Поражение."
		return
	}
	if !c.TeamAlive(TeamEnemy) {
		c.Result = ResultVictory
		c.LastLog = "Победа!"
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

// DisplayPhase возвращает BattlePhase для UI (ход игрока / ход врага / завершён).
func (c *BattleContext) DisplayPhase() BattlePhase {
	if c.Phase == PhaseFinished || c.Phase == PhaseFinishedWaitInput || c.Result != ResultNone {
		return BattlePhaseFinished
	}
	u := c.ActiveUnit()
	if u == nil {
		return BattlePhaseFinished
	}
	if u.Team == TeamPlayer {
		return BattlePhasePlayerTurn
	}
	return BattlePhaseEnemyTurn
}

// PlayerHP возвращает HP первого живого юнита игрока (для UI).
func (c *BattleContext) PlayerHP() int {
	live := c.LivingUnits(TeamPlayer)
	if len(live) == 0 {
		return 0
	}
	return live[0].HP
}

// EnemyHP возвращает HP первого живого юнита врага (для UI).
func (c *BattleContext) EnemyHP() int {
	live := c.LivingUnits(TeamEnemy)
	if len(live) == 0 {
		return 0
	}
	return live[0].HP
}

// CanPlayerActNow возвращает true, если сейчас ход игрока и можно выполнить действие.
func (c *BattleContext) CanPlayerActNow() bool {
	if c.Phase != PhaseAwaitAction || c.Result != ResultNone {
		return false
	}
	u := c.ActiveUnit()
	return u != nil && u.IsAlive() && u.Team == TeamPlayer
}

// ActiveUnitName возвращает имя активного юнита для UI.
func (c *BattleContext) ActiveUnitName() string {
	u := c.ActiveUnit()
	if u == nil {
		return "-"
	}
	return u.Name
}

// ActiveUnitTeamName возвращает "Player" или "Enemy" для активного юнита.
func (c *BattleContext) ActiveUnitTeamName() string {
	u := c.ActiveUnit()
	if u == nil {
		return "-"
	}
	if u.Team == TeamPlayer {
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
	case PhaseTurnResolve:
		return "TurnResolve"
	case PhaseActionPause:
		return "ActionPause"
	case PhaseTurnEnd:
		return "TurnEnd"
	case PhaseRoundEnd:
		return "RoundEnd"
	case PhaseFinishedWaitInput:
		return "FinishedWaitInput"
	case PhaseFinished:
		return "Finished"
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
	if r.Target != 0 {
		target := c.Units[r.Target]
		if actor != nil && target != nil && r.Damage > 0 {
			c.LastLog = logActionResult(actor, target, r)
		}
	}
	c.UpdateResultIfFinished()
}

func logActionResult(actor, target *BattleUnit, r ActionResult) string {
	if r.Killed {
		return actor.Name + " победил " + target.Name + "."
	}
	return formatDamageLog(actor.Name, target.Name, r.Damage)
}

func formatDamageLog(actorName, targetName string, damage int) string {
	if damage <= 0 {
		return actorName + " атаковал " + targetName + "."
	}
	return fmt.Sprintf("%s атаковал %s на %d урона.", actorName, targetName, damage)
}
