// Package party is the canonical roster layer between world/explore and battle.
// Party = Active (в бою и построение) + Reserve/Bench (вне боя, не попадают в PlayerCombatSeeds).
// Лидер — всегда Active[0]; найм/save/camp — отдельные этапы. Party-wide XP после побед: progression.ApplyVictoryCombatXPForActiveSurvivors.
package party

import (
	"errors"
	"fmt"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
)

// MaxActiveBattleSlots — максимум участников в Active (ёмкость построения боя 3+3).
const MaxActiveBattleSlots = battlepkg.MaxFrontRowUnits + battlepkg.MaxBackRowUnits

// MaxPartyMembers — верхняя граница размера полного ростера (Active + Reserve).
const MaxPartyMembers = 12

// ErrPartyFull — нельзя добавить участника: достигнут MaxPartyMembers.
var ErrPartyFull = errors.New("party: roster full")

// Party — канонический отряд: Active (участвуют в бою в порядке списка) и Reserve (скамейка).
// Active[0] — лидер (выбор награды после победы). Боевой опыт за победу получают все выжившие в активном строю.
// Порядок Active задаёт построение в бою (см. battle.PlayerSlotForPartyIndex).
type Party struct {
	Active  []hero.Hero
	Reserve []hero.Hero
}

// DefaultParty возвращает отряд из одного героя (обычный single-unit режим).
func DefaultParty() Party {
	return Party{
		Active: []hero.Hero{hero.DefaultHero()},
	}
}

// TwoMemberDemoParty — 2 в строю + 1 в резерве (проверка multi-unit и bench).
func TwoMemberDemoParty() Party {
	return Party{
		Active: []hero.Hero{
			hero.DefaultHero(),
			hero.DefaultHero(),
		},
		Reserve: []hero.Hero{hero.DefaultHero()},
	}
}

// HeroAtGlobalIndex возвращает указатель на героя по индексу списка formation:
// [0, len(Active)) — строй, [len(Active), ...) — резерв.
func (p *Party) HeroAtGlobalIndex(globalIdx int) *hero.Hero {
	if p == nil {
		return nil
	}
	na := len(p.Active)
	if globalIdx >= 0 && globalIdx < na {
		return &p.Active[globalIdx]
	}
	j := globalIdx - na
	if j >= 0 && j < len(p.Reserve) {
		return &p.Reserve[j]
	}
	return nil
}

// Leader возвращает указатель на лидера для мутации (прогрессия, награды).
// Инвариант: у игрока всегда есть хотя бы один активный участник.
func (p *Party) Leader() *hero.Hero {
	if len(p.Active) == 0 {
		return nil
	}
	return &p.Active[0]
}

// PlayerCombatSeeds проецирует активный ростер в сиды для стороны игрока.
// Только участники с CurrentHP > 0; Origin.PartyActiveIndex = индекс в Active (для post-battle sync).
func (p *Party) PlayerCombatSeeds() []battlepkg.CombatUnitSeed {
	if len(p.Active) == 0 {
		return nil
	}
	out := make([]battlepkg.CombatUnitSeed, 0, len(p.Active))
	for i := range p.Active {
		if !p.Active[i].CanEnterBattle() {
			continue
		}
		s := p.Active[i].CombatUnitSeed()
		s.Origin.PartyActiveIndex = i
		out = append(out, s)
	}
	return out
}

// MoveActiveEarlier меняет местами участника с индексом i с предыдущим (i>0). Курсор UI должен следовать за участником (i-1).
func (p *Party) MoveActiveEarlier(i int) bool {
	if p == nil || i <= 0 || i >= len(p.Active) {
		return false
	}
	p.Active[i-1], p.Active[i] = p.Active[i], p.Active[i-1]
	return true
}

// MoveActiveLater меняет местами участника с индексом i со следующим. Курсор UI: i+1.
func (p *Party) MoveActiveLater(i int) bool {
	if p == nil || i < 0 || i >= len(p.Active)-1 {
		return false
	}
	p.Active[i], p.Active[i+1] = p.Active[i+1], p.Active[i]
	return true
}

// MoveActiveToReserve переносит участника с индексом i из Active в конец Reserve.
// Инварианты: минимум один участник остаётся в Active; лидер — новый Active[0] после снятия.
func (p *Party) MoveActiveToReserve(i int) bool {
	if p == nil || i < 0 || i >= len(p.Active) {
		return false
	}
	if len(p.Active) <= 1 {
		return false
	}
	h := p.Active[i]
	p.Active = append(p.Active[:i], p.Active[i+1:]...)
	p.Reserve = append(p.Reserve, h)
	return true
}

// MoveReserveToActive переносит участника с индексом j из Reserve в конец Active.
// Не больше maxActiveBattleSlots в Active (ёмкость боя).
func (p *Party) MoveReserveToActive(j int) bool {
	if p == nil || j < 0 || j >= len(p.Reserve) {
		return false
	}
	if len(p.Active) >= MaxActiveBattleSlots {
		return false
	}
	h := p.Reserve[j]
	p.Reserve = append(p.Reserve[:j], p.Reserve[j+1:]...)
	p.Active = append(p.Active, h)
	return true
}

// ActiveCount / ReserveCount — для UI.
func (p *Party) ActiveCount() int {
	if p == nil {
		return 0
	}
	return len(p.Active)
}

func (p *Party) ReserveCount() int {
	if p == nil {
		return 0
	}
	return len(p.Reserve)
}

// TotalMembers — число героев в Active и Reserve.
func (p *Party) TotalMembers() int {
	if p == nil {
		return 0
	}
	return len(p.Active) + len(p.Reserve)
}

// AddToReserve добавляет героя в конец резерва. Не меняет Active.
func (p *Party) AddToReserve(h hero.Hero) error {
	if p == nil {
		return fmt.Errorf("party: nil")
	}
	if p.TotalMembers() >= MaxPartyMembers {
		return ErrPartyFull
	}
	p.Reserve = append(p.Reserve, h)
	return nil
}

// FormationSlotCaption описывает боевой слот для i-го по порядку участника (тот же индекс, что и battle.PlayerSlotForPartyIndex).
func FormationSlotCaption(partyIndex int) string {
	row, idx := battlepkg.PlayerSlotForPartyIndex(partyIndex)
	if row == battlepkg.BattleRowFront {
		return fmt.Sprintf("Передний ряд · ячейка %d", idx+1)
	}
	return fmt.Sprintf("Задний ряд · ячейка %d", idx+1)
}

// MemberRoleCaption короткая роль в списке (лидер / номер союзника).
func MemberRoleCaption(index int) string {
	if index == 0 {
		return "Лидер (награда за бой)"
	}
	return fmt.Sprintf("Союзник %d", index)
}

// ReserveRowCaption — подпись строки в резерве (без боевого слота).
func ReserveRowCaption(reserveIndex int) string {
	return fmt.Sprintf("Резерв · %d", reserveIndex+1)
}

// --- Отдых в мире (R): ход мира без бесплатного лечения ОЗ ---

// HasFightableMember true, если хотя бы один участник может получить сид боя (CurrentHP > 0).
func (p *Party) HasFightableMember() bool {
	if p == nil {
		return false
	}
	for i := range p.Active {
		if p.Active[i].CanEnterBattle() {
			return true
		}
	}
	return false
}

// ApplyWorldRest — контракт «отдых на R» в explore: не меняет ОЗ (лечение только через явные источники:
// бои/заклинания, POI, будущие еда/зелья и т.д.). Вызывать вместе с advanceWorldTurn.
// Если никто в Active не может сражаться (все 0 ОЗ), лидер получает 1 ОЗ — выход из soft-lock.
func (p *Party) ApplyWorldRest() {
	if p == nil || len(p.Active) == 0 {
		return
	}
	if !p.HasFightableMember() {
		p.Active[0].CurrentHP = 1
	}
}

// RestExploreBanner короткий текст обратной связи после отдыха в explore.
const RestExploreBanner = "Отдых: ход мира (ОЗ не восстанавливаются)"
