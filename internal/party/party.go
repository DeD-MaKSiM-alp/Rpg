// Package party is the canonical roster layer between world/explore and battle.
// A Party holds the persistent combat-capable members (Hero) in deployment order; the leader is Active[0].
// Bench/reserve, hire/replace, and injured states can extend this struct later without changing the hero model.
package party

import (
	"fmt"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
)

// Party — канонический отряд игрока вне боя: упорядоченный список активных участников.
// Active[0] — лидер (награды после боя, привязка к аватару на карте по смыслу).
// Порядок Active задаёт построение в бою (см. battle.PlayerSlotForPartyIndex).
type Party struct {
	Active []hero.Hero
}

// DefaultParty возвращает отряд из одного героя (обычный single-unit режим).
func DefaultParty() Party {
	return Party{
		Active: []hero.Hero{hero.DefaultHero()},
	}
}

// TwoMemberDemoParty возвращает отряд из двух героев со стартовыми статами (для проверки multi-unit боя).
// Обычный NewGame использует DefaultParty(); подставьте эту функцию вручную при отладке.
func TwoMemberDemoParty() Party {
	return Party{
		Active: []hero.Hero{
			hero.DefaultHero(),
			hero.DefaultHero(),
		},
	}
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
		return "Лидер (награды)"
	}
	return fmt.Sprintf("Союзник %d", index)
}

// --- Отдых в мире (минимальный recovery loop между боями) ---

// RestRecoveryDivisor — восстановление за отдых: max(1, MaxHP/RestRecoveryDivisor) к CurrentHP.
const RestRecoveryDivisor = 4

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

// ApplyWorldRest восстанавливает канонический CurrentHP после «отдыха» в explore (вызывать вместе с advanceWorldTurn).
// Правила:
//   - для каждого участника с CurrentHP > 0: +max(1, MaxHP/RestRecoveryDivisor), clamp к MaxHP;
//   - участников с 0 HP отдых не воскрешает (нужны будущие revive/camp);
//   - если после этого никто не может сражаться (все 0), лидер получает 1 HP — аварийный выход из soft-lock.
func (p *Party) ApplyWorldRest() {
	if p == nil || len(p.Active) == 0 {
		return
	}
	for i := range p.Active {
		h := &p.Active[i]
		if h.CurrentHP <= 0 {
			continue
		}
		gain := h.MaxHP / RestRecoveryDivisor
		if gain < 1 {
			gain = 1
		}
		h.CurrentHP += gain
		if h.CurrentHP > h.MaxHP {
			h.CurrentHP = h.MaxHP
		}
	}
	if !p.HasFightableMember() {
		p.Active[0].CurrentHP = 1
	}
}

// RestExploreBanner короткий текст обратной связи после отдыха в explore.
const RestExploreBanner = "Отдых: +HP (доля MaxHP), ход мира"
