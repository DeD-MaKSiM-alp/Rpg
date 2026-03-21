package game

import (
	"strings"

	"mygame/internal/hero"
	"mygame/internal/party"
)

// PromotionExploreHUDLine — одна строка для верхнего HUD в explore: готовность повышения по отряду.
// Логика только через EvaluatePromotionGate (домен не дублируется).
func PromotionExploreHUDLine(p *party.Party, atCamp bool, trainingMarks int) string {
	if p == nil || p.TotalMembers() == 0 {
		return ""
	}
	list := promotionHeroPointers(p)
	var hasPath bool
	for _, hp := range list {
		targets, err := hero.PromotionTargetUnitIDs(hp)
		if err != nil || len(targets) == 0 {
			continue
		}
		hasPath = true
		for _, tid := range targets {
			if EvaluatePromotionGate(hp, atCamp, trainingMarks, tid).Allowed {
				return "Повышение: готово — F5 → I → P"
			}
		}
	}
	if !hasPath {
		return ""
	}
	if !atCamp {
		return "Повышение: нужен лагерь (лазурный маркер) · F5 → I"
	}
	for _, hp := range list {
		targets, err := hero.PromotionTargetUnitIDs(hp)
		if err != nil || len(targets) == 0 {
			continue
		}
		if len(targets) == 1 {
			res := EvaluatePromotionGate(hp, atCamp, trainingMarks, targets[0])
			return trimPromoHUDLine(res.Message)
		}
		allMarks, anyMarks := true, false
		for _, tid := range targets {
			res := EvaluatePromotionGate(hp, atCamp, trainingMarks, tid)
			if strings.Contains(res.Message, "знаков") {
				anyMarks = true
			} else {
				allMarks = false
			}
		}
		if allMarks && anyMarks {
			res := EvaluatePromotionGate(hp, atCamp, trainingMarks, targets[0])
			return trimPromoHUDLine(res.Message)
		}
		return "Повышение: выберите ветку в карточке (F5 → I)"
	}
	return ""
}

func trimPromoHUDLine(s string) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > 72 {
		rs := []rune(s)
		return string(rs[:69]) + "…"
	}
	return s
}

// PromotionExploreStripLine — короткая строка для полоски отряда (explore), без дублирования длинного HUD.
func PromotionExploreStripLine(p *party.Party, atCamp bool, trainingMarks int) string {
	if p == nil || p.TotalMembers() == 0 {
		return ""
	}
	list := promotionHeroPointers(p)
	for _, hp := range list {
		targets, err := hero.PromotionTargetUnitIDs(hp)
		if err != nil || len(targets) == 0 {
			continue
		}
		for _, tid := range targets {
			if EvaluatePromotionGate(hp, atCamp, trainingMarks, tid).Allowed {
				return "Повышение: готово — F5 → I"
			}
		}
	}
	for _, hp := range list {
		targets, err := hero.PromotionTargetUnitIDs(hp)
		if err != nil || len(targets) == 0 {
			continue
		}
		if !atCamp {
			return "Повышение: нужен лагерь"
		}
		if len(targets) >= 2 {
			allMarks, anyMarks := true, false
			for _, tid := range targets {
				res := EvaluatePromotionGate(hp, atCamp, trainingMarks, tid)
				if strings.Contains(res.Message, "знаков") {
					anyMarks = true
				} else {
					allMarks = false
				}
			}
			if allMarks && anyMarks {
				return "Повышение: мало знаков"
			}
			return "Повышение: выберите ветку (F5→I)"
		}
		res := EvaluatePromotionGate(hp, atCamp, trainingMarks, targets[0])
		if strings.Contains(res.Message, "знаков") {
			return "Повышение: мало знаков"
		}
		break
	}
	return ""
}

// PromotionInspectHeadline — первая строка блока «Развитие» в карточке; branchIdx для ветвления (-1 = ветка не выбрана).
func PromotionInspectHeadline(h *hero.Hero, atCamp bool, trainingMarks int, promoteTargets []string, branchIdx int) string {
	if h == nil || len(promoteTargets) == 0 {
		return ""
	}
	var sel string
	if len(promoteTargets) == 1 {
		sel = promoteTargets[0]
	} else {
		if branchIdx < 0 || branchIdx >= len(promoteTargets) {
			return "Повышение: выберите ветку (←/→), затем P"
		}
		sel = promoteTargets[branchIdx]
	}
	res := EvaluatePromotionGate(h, atCamp, trainingMarks, sel)
	if res.Allowed {
		return "Повышение: готово — нажмите P"
	}
	if !atCamp {
		return "Повышение: нужен лагерь (лазурный маркер)"
	}
	return trimPromoHUDLine(res.Message)
}

// PromotionFormationRowHints — подпись справа у строки состава; индекс = globalIdx (строй, затем резерв).
func PromotionFormationRowHints(p *party.Party, atCamp bool, trainingMarks int) []string {
	if p == nil {
		return nil
	}
	n := len(p.Active) + len(p.Reserve)
	out := make([]string, n)
	for i := range p.Active {
		out[i] = promotionFormationRowHint(&p.Active[i], atCamp, trainingMarks)
	}
	for j := range p.Reserve {
		out[len(p.Active)+j] = promotionFormationRowHint(&p.Reserve[j], atCamp, trainingMarks)
	}
	return out
}

func promotionHeroPointers(p *party.Party) []*hero.Hero {
	var list []*hero.Hero
	for i := range p.Active {
		list = append(list, &p.Active[i])
	}
	for i := range p.Reserve {
		list = append(list, &p.Reserve[i])
	}
	return list
}

func promotionFormationRowHint(h *hero.Hero, atCamp bool, trainingMarks int) string {
	if h == nil {
		return ""
	}
	targets, err := hero.PromotionTargetUnitIDs(h)
	if err != nil || len(targets) == 0 {
		return ""
	}
	for _, tid := range targets {
		if EvaluatePromotionGate(h, atCamp, trainingMarks, tid).Allowed {
			return "Повышение!"
		}
	}
	if !atCamp {
		return "Лагерь"
	}
	if len(targets) >= 2 {
		allMarks, anyMarks := true, false
		for _, tid := range targets {
			res := EvaluatePromotionGate(h, atCamp, trainingMarks, tid)
			if strings.Contains(res.Message, "знаков") {
				anyMarks = true
			} else {
				allMarks = false
			}
		}
		if allMarks && anyMarks {
			return "Знаки"
		}
		return "Ветка"
	}
	res := EvaluatePromotionGate(h, atCamp, trainingMarks, targets[0])
	if strings.Contains(res.Message, "знаков") {
		return "Знаки"
	}
	return ""
}
