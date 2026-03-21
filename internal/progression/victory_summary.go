package progression

import (
	"fmt"
	"strings"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
	"mygame/internal/unitdata"
)

// VictoryProgressionSummary — человекочитаемые строки для экрана после победы (механика не меняется).
type VictoryProgressionSummary struct {
	Lines []string
}

// HeroShortLabel — краткое имя бойца для сводок (шаблон · рекрут · роль).
func HeroShortLabel(h *hero.Hero, slotIndex int) string {
	if h == nil {
		return "—"
	}
	if tpl, ok := unitdata.GetUnitTemplate(h.UnitID); ok {
		return tpl.DisplayName
	}
	if h.RecruitLabel != "" {
		return h.RecruitLabel
	}
	return party.MemberRoleCaption(slotIndex)
}

// BuildVictoryProgressionSummary собирает строки после победы.
// Вызывать после syncPartyFromBattle и ApplyVictoryCombatXPForActiveSurvivors, пока BattleContext ещё доступен.
// trainingMarksDelta — сколько знаков начислено за эту победу (совпадает с game.TrainingMarksPerVictory).
func BuildVictoryProgressionSummary(b *battlepkg.BattleContext, roster *party.Party, trainingMarksDelta int) VictoryProgressionSummary {
	if b == nil || roster == nil {
		return VictoryProgressionSummary{}
	}
	amount := CombatExperiencePerVictorySurvivor
	var gotXP []string
	var dead []string
	for _, u := range b.Units {
		if u == nil || u.Side != battlepkg.TeamPlayer {
			continue
		}
		idx := u.Origin.PartyActiveIndex
		if idx < 0 || idx >= len(roster.Active) {
			continue
		}
		name := HeroShortLabel(&roster.Active[idx], idx)
		if !u.IsAlive() {
			dead = append(dead, name)
			continue
		}
		gotXP = append(gotXP, name)
	}

	var lines []string
	if len(gotXP) > 0 {
		if len(gotXP) <= 4 {
			lines = append(lines, fmt.Sprintf("Боевой опыт (+%d каждому): %s", amount, strings.Join(gotXP, ", ")))
		} else {
			lines = append(lines, fmt.Sprintf("Боевой опыт (+%d): %d участников в строю", amount, len(gotXP)))
		}
	} else {
		lines = append(lines, "Боевой опыт: некому начислить (нет выживших в строю).")
	}
	lines = append(lines, fmt.Sprintf("Опыт усиливает обычную атаку: каждые %d шага — +1 к урону базового удара.", hero.CombatXPStepsPerBasicAttackBonus))
	if len(dead) > 0 {
		lines = append(lines, fmt.Sprintf("Поверженные: %s — боевой опыт не получили.", strings.Join(dead, ", ")))
	}
	if len(roster.Reserve) > 0 {
		lines = append(lines, fmt.Sprintf("Резерв (%d): в бою не участвовали — боевой опыт не начисляется.", len(roster.Reserve)))
	}
	if trainingMarksDelta > 0 {
		lines = append(lines, fmt.Sprintf("Знаки обучения: +%d (тратятся на повышение в лагере).", trainingMarksDelta))
	}
	return VictoryProgressionSummary{Lines: lines}
}
