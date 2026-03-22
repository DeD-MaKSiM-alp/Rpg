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

// BuildVictoryProgressionSummary собирает короткие строки после победы (player-facing: меньше шума).
func BuildVictoryProgressionSummary(b *battlepkg.BattleContext, roster *party.Party, trainingMarksDelta int, levelUps []CombatLevelUp) VictoryProgressionSummary {
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
			lines = append(lines, fmt.Sprintf("Опыт +%d: %s", amount, strings.Join(gotXP, ", ")))
		} else {
			lines = append(lines, fmt.Sprintf("Опыт +%d · в строю %d", amount, len(gotXP)))
		}
	} else {
		lines = append(lines, "Опыт не начислен — нет выживших в строю.")
	}

	if len(levelUps) > 0 {
		var parts []string
		for _, up := range levelUps {
			if roster == nil || up.PartyActiveIndex < 0 || up.PartyActiveIndex >= len(roster.Active) {
				continue
			}
			name := HeroShortLabel(&roster.Active[up.PartyActiveIndex], up.PartyActiveIndex)
			parts = append(parts, fmt.Sprintf("%s %d→%d", name, up.OldLevel, up.NewLevel))
		}
		if len(parts) > 0 {
			lines = append(lines, "Уровень↑: "+strings.Join(parts, " · "))
		}
	}

	lines = append(lines, fmt.Sprintf("Новый уровень каждые %d опыта (+1 к атаке).", hero.CombatXPPerLevel))

	if len(dead) > 0 {
		lines = append(lines, "Без опыта: "+strings.Join(dead, ", "))
	}
	if len(roster.Reserve) > 0 {
		lines = append(lines, fmt.Sprintf("Резерв (%d): в бою не был — без опыта.", len(roster.Reserve)))
	}
	if trainingMarksDelta > 0 {
		lines = append(lines, fmt.Sprintf("Знаки обучения +%d", trainingMarksDelta))
	}
	return VictoryProgressionSummary{Lines: lines}
}
