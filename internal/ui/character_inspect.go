package ui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
	"mygame/internal/unitdata"
)

// DrawCharacterInspectOverlay — карточка бойца (F5 → I / ПКМ); тот же визуальный шаблон, что и battle inspect.
// promotionHeadline — первая строка блока «Развитие» (готовность повышения из game.PromotionInspectHeadline); может быть пустой.
func DrawCharacterInspectOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, globalIdx int, screenW, screenH int, feedbackBanner string, atCamp bool, trainingMarks int, promoteTargets []string, promoteCosts []int, branchIdx int, promotionHeadline string) {
	if hudFace == nil || p == nil {
		return
	}
	h := p.HeroAtGlobalIndex(globalIdx)
	if h == nil {
		return
	}

	sw := float32(screenW)
	sh := float32(screenH)
	na := len(p.Active)
	inReserve := globalIdx >= na

	m := buildFormationInspectCardModel(h, globalIdx, na, inReserve, atCamp, trainingMarks, promoteTargets, promoteCosts, branchIdx, feedbackBanner, promotionHeadline)

	panelW := DefaultInspectCardPanelWidth(screenW)
	panelH := EstimateInspectCardHeight(m)
	px := (sw - panelW) / 2
	py := (sh - panelH) * 0.45
	if py < 16 {
		py = 16
	}

	DrawInspectCardChrome(screen, px, py, panelW, panelH, false)
	DrawInspectCardContent(screen, hudFace, px, py, panelW, m)
}

func buildFormationInspectCardModel(h *hero.Hero, globalIdx, na int, inReserve bool, atCamp bool, trainingMarks int, promoteTargets []string, promoteCosts []int, branchIdx int, feedbackBanner string, promotionHeadline string) InspectCardModel {
	m := InspectCardModel{
		RoleIcon:    InspectRoleIconFromHero(h),
		Title:       inspectPrimaryTitle(h, globalIdx, na),
		ContextLine: formationContextLine(globalIdx, na, inReserve),
		HPCur:       h.CurrentHP,
		HPMax:       h.MaxHP,
		Alive:       h.CurrentHP > 0,
		IsEnemy:     false,
		Footer:      inspectPromotionFooterHint(atCamp, len(promoteTargets)),
		FeedbackBanner: strings.TrimSpace(feedbackBanner),
	}
	m.Badges = compactTierRangeBadgesFromHero(h)
	m.ProfileLines = templateProfileShortLines(h)
	healTotal := 2 + h.HealPower
	m.StatsLine = fmt.Sprintf("Атака %d · Защита %d · Инициатива %d · Лечение +%d", h.Attack, h.Defense, h.Initiative, healTotal)
	m.ExtraStatLine = ""
	m.AbilityLines = abilityLinesBullet(h.Abilities)
	lines := formationInspectProgressLines(h, inReserve, atCamp, trainingMarks, promoteTargets, promoteCosts, branchIdx)
	if strings.TrimSpace(promotionHeadline) != "" {
		lines = append([]string{strings.TrimSpace(promotionHeadline)}, lines...)
	}
	m.ProgressLines = lines
	return m
}

func formationContextLine(globalIdx, na int, inReserve bool) string {
	if inReserve {
		return "Резерв — вне боя"
	}
	return party.FormationSlotCaption(globalIdx)
}

const maxFormationInspectProgressLines = 8

func formationInspectProgressLines(h *hero.Hero, inReserve bool, atCamp bool, trainingMarks int, promoteTargets []string, promoteCosts []int, branchIdx int) []string {
	if h == nil {
		return nil
	}
	var out []string
	out = append(out, FormatCombatXPInspectLines(h)...)
	promo := inspectPromotionLines(h, atCamp, trainingMarks, promoteTargets, promoteCosts, branchIdx)
	for _, ln := range promo {
		out = append(out, ln)
		if len(out) >= maxFormationInspectProgressLines-1 {
			break
		}
	}
	if inReserve {
		out = append(out, "Резерв: опыт в бою не растёт.")
	} else if h.CurrentHP <= 0 {
		out = append(out, "0 ОЗ — в бой нельзя.")
	}
	if len(out) > maxFormationInspectProgressLines {
		return out[:maxFormationInspectProgressLines]
	}
	return out
}

func inspectPromotionFooterHint(atCamp bool, nTargets int) string {
	if nTargets >= 2 {
		if atCamp {
			return "I / Esc — закрыть · ↑↓ — другой · ←/→ — ветка · P — повышение"
		}
		return "I / Esc — закрыть · ↑↓ — другой · ←/→ — ветка · P — в лагере"
	}
	if atCamp {
		return "I / Esc — закрыть · ↑↓ — другой · P — повышение (знаки обучения)"
	}
	return "I / Esc — закрыть · ↑↓ — другой · P — в лагере лазурного маркера"
}

func inspectPrimaryTitle(h *hero.Hero, globalIdx, na int) string {
	if tpl, ok := unitdata.GetUnitTemplate(h.UnitID); ok {
		return tpl.DisplayName
	}
	if h.RecruitLabel != "" {
		return h.RecruitLabel
	}
	if globalIdx < na {
		return party.MemberRoleCaption(globalIdx)
	}
	return party.ReserveRowCaption(globalIdx - na)
}

func inspectPromotionLines(h *hero.Hero, atCamp bool, trainingMarks int, promoteTargets []string, promoteCosts []int, branchIdx int) []string {
	if err := hero.ValidatePromotionPathsExist(h); err != nil {
		return []string{hero.PromotionErrUserMessage(err)}
	}
	if len(promoteTargets) == 0 {
		return []string{"Нет шага повышения."}
	}
	if len(promoteTargets) == 1 {
		nextID := promoteTargets[0]
		nextLabel := promotionTargetDisplayName(nextID)
		cost := 0
		if len(promoteCosts) > 0 {
			cost = promoteCosts[0]
		}
		if !atCamp {
			return []string{fmt.Sprintf("Следующий: «%s» · %d знаков · только в лагере", nextLabel, cost)}
		}
		if trainingMarks < cost {
			return []string{fmt.Sprintf("Следующий: «%s» · %d/%d знаков", nextLabel, trainingMarks, cost)}
		}
		return []string{fmt.Sprintf("Следующий: «%s» · P (%d знаков)", nextLabel, cost)}
	}
	// Несколько веток: одна строка-сводка + при необходимости строка статуса.
	var parts []string
	for i := range promoteTargets {
		tpl, ok := unitdata.GetUnitTemplate(promoteTargets[i])
		if !ok {
			continue
		}
		c := 0
		if i < len(promoteCosts) {
			c = promoteCosts[i]
		}
		mark := ""
		if branchIdx == i {
			mark = "▸ "
		}
		parts = append(parts, fmt.Sprintf("%s«%s» %d", mark, tpl.DisplayName, c))
	}
	if len(parts) == 0 {
		return []string{"Нет шага повышения."}
	}
	summary := strings.Join(parts, " · ")
	if branchIdx < 0 {
		if atCamp {
			summary = "←/→ ветка · " + summary
		} else {
			summary = "Лагерь · " + summary
		}
		return []string{summary}
	}
	cost := 0
	if branchIdx >= 0 && branchIdx < len(promoteCosts) {
		cost = promoteCosts[branchIdx]
	}
	if !atCamp {
		return []string{summary, fmt.Sprintf("В лагере · %d/%d знаков", trainingMarks, cost)}
	}
	if trainingMarks < cost {
		return []string{summary, fmt.Sprintf("Нужно %d знаков · есть %d", cost, trainingMarks)}
	}
	return []string{summary, fmt.Sprintf("P — повышение (%d знаков)", cost)}
}

func promotionTargetDisplayName(unitID string) string {
	if tpl, ok := unitdata.GetUnitTemplate(unitID); ok {
		return tpl.DisplayName
	}
	return unitID
}

func roleLabelRu(r battlepkg.Role) string {
	switch r {
	case battlepkg.RoleFighter:
		return "боец"
	case battlepkg.RoleArcher:
		return "лучник"
	case battlepkg.RoleHealer:
		return "целитель"
	case battlepkg.RoleMage:
		return "маг"
	default:
		return "—"
	}
}

