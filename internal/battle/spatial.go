package battle

import "fmt"

// BattleRow — ряд построения (front/back).
type BattleRow int

const (
	BattleRowFront BattleRow = iota
	BattleRowBack
)

// RowType — legacy имя ряда (compatibility).
type RowType = BattleRow

const (
	RowFront RowType = BattleRowFront
	RowBack  RowType = BattleRowBack
)

// Formation size (baseline contract).
const (
	MaxFrontRowUnits = 3
	MaxBackRowUnits  = 3
)

// BattleSlotID — стабильный идентификатор позиции (side + row + index).
type BattleSlotID struct {
	Side  BattleSide
	Row   BattleRow
	Index int
}

func (id BattleSlotID) String() string {
	return fmt.Sprintf("%v/%v/%d", id.Side, id.Row, id.Index)
}

// BattleSlot — фиксированная позиция в построении.
// Source of truth: occupied unit is tracked here (not inside unit.State).
type BattleSlot struct {
	ID       BattleSlotID
	Occupied UnitID // 0 means empty
}

func (s BattleSlot) IsEmpty() bool { return s.Occupied == 0 }

// BattleSideState — formation slots for a side + helpers.
type BattleSideState struct {
	Side  BattleSide
	Slots []BattleSlot // fixed set
}

func NewBattleSideState(side BattleSide) *BattleSideState {
	st := &BattleSideState{Side: side}
	// Fixed slot set: front + back.
	for i := 0; i < MaxFrontRowUnits; i++ {
		st.Slots = append(st.Slots, BattleSlot{ID: BattleSlotID{Side: side, Row: BattleRowFront, Index: i}})
	}
	for i := 0; i < MaxBackRowUnits; i++ {
		st.Slots = append(st.Slots, BattleSlot{ID: BattleSlotID{Side: side, Row: BattleRowBack, Index: i}})
	}
	return st
}

func (s *BattleSideState) Slot(row BattleRow, index int) *BattleSlot {
	if s == nil || index < 0 {
		return nil
	}
	for i := range s.Slots {
		sl := &s.Slots[i]
		if sl.ID.Row == row && sl.ID.Index == index {
			return sl
		}
	}
	return nil
}

func (s *BattleSideState) SlotByUnit(id UnitID) *BattleSlot {
	if s == nil || id == 0 {
		return nil
	}
	for i := range s.Slots {
		sl := &s.Slots[i]
		if sl.Occupied == id {
			return sl
		}
	}
	return nil
}

func (s *BattleSideState) OccupiedSlots() []*BattleSlot {
	if s == nil {
		return nil
	}
	out := make([]*BattleSlot, 0, len(s.Slots))
	for i := range s.Slots {
		if !s.Slots[i].IsEmpty() {
			out = append(out, &s.Slots[i])
		}
	}
	return out
}

func (s *BattleSideState) OccupiedSlotsInRow(row BattleRow) []*BattleSlot {
	if s == nil {
		return nil
	}
	out := make([]*BattleSlot, 0, len(s.Slots))
	for i := range s.Slots {
		sl := &s.Slots[i]
		if sl.ID.Row == row && !sl.IsEmpty() {
			out = append(out, sl)
		}
	}
	return out
}

// --- BattleContext helpers (spatial model integration) ---

func (c *BattleContext) SideState(side BattleSide) *BattleSideState {
	if c == nil || c.Sides == nil {
		return nil
	}
	return c.Sides[side]
}

// Slot returns a slot by side/row/index.
func (c *BattleContext) Slot(side BattleSide, row BattleRow, index int) *BattleSlot {
	return c.SideState(side).Slot(row, index)
}

// SlotByUnit returns the slot that currently contains the unit.
func (c *BattleContext) SlotByUnit(id UnitID) *BattleSlot {
	if c == nil || c.Sides == nil || id == 0 {
		return nil
	}
	if u := c.Units[id]; u != nil {
		if st := c.SideState(u.Side); st != nil {
			return st.SlotByUnit(id)
		}
	}
	// Fallback: scan both sides (defensive).
	for _, st := range c.Sides {
		if sl := st.SlotByUnit(id); sl != nil {
			return sl
		}
	}
	return nil
}

func (c *BattleContext) UnitInSlot(slot *BattleSlot) *BattleUnit {
	if c == nil || slot == nil || slot.Occupied == 0 {
		return nil
	}
	return c.Units[slot.Occupied]
}

func (c *BattleContext) IsSlotEmpty(slot *BattleSlot) bool {
	return slot == nil || slot.IsEmpty()
}

func (c *BattleContext) OccupiedSlots(side BattleSide) []*BattleSlot {
	return c.SideState(side).OccupiedSlots()
}

// PlaceUnit places unit into slot (source of truth is slot occupancy).
// Also mirrors placement into unit.State.Row/Slot for compatibility.
func (c *BattleContext) PlaceUnit(unitID UnitID, slotID BattleSlotID) bool {
	if c == nil || unitID == 0 || c.Units == nil {
		return false
	}
	u := c.Units[unitID]
	if u == nil {
		return false
	}
	st := c.SideState(slotID.Side)
	if st == nil {
		return false
	}
	if u.Side != slotID.Side {
		return false
	}
	sl := st.Slot(slotID.Row, slotID.Index)
	if sl == nil || !sl.IsEmpty() {
		return false
	}
	sl.Occupied = unitID
	u.State.Row = slotID.Row
	u.State.Slot = slotID.Index
	return true
}

