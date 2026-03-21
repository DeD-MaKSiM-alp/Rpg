package ui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
	"mygame/internal/unitdata"
)

// DrawCharacterInspectOverlay — карточка бойца (канонические поля Hero); открывается из состава (F5) по I.
// feedbackBanner — краткий результат promotion (успех/ошибка); может быть пустым.
// atCamp — игрок на клетке лагеря наёмников (для строки про promotion).
func DrawCharacterInspectOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, globalIdx int, screenW, screenH int, feedbackBanner string, atCamp bool) {
	if hudFace == nil || p == nil {
		return
	}
	h := p.HeroAtGlobalIndex(globalIdx)
	if h == nil {
		return
	}

	sw := float32(screenW)
	sh := float32(screenH)
	// Затемнение уже задано экраном состава; здесь только панель карточки.

	na := len(p.Active)
	inReserve := globalIdx >= na
	title := inspectPrimaryTitle(h, globalIdx, na)
	sub := inspectSubtitle(h, globalIdx, na, inReserve)

	metrics := battlepkg.HUDMetrics{LineH: uiLineH}
	pad := float32(18)
	panelW := float32(520)
	if sw-40 < panelW {
		panelW = sw - 40
	}
	lineH := uiLineH
	lines := buildInspectLines(h, inReserve, atCamp)
	extraFeedback := float32(0)
	if feedbackBanner != "" {
		extraFeedback = lineH * 1.15
	}
	panelH := pad*2 + lineH*1.4 + lineH*1.2 + float32(len(lines)+2)*lineH + 24 + extraFeedback
	px := (sw - panelW) / 2
	py := (sh - panelH) * 0.45
	if py < 20 {
		py = 20
	}

	vector.FillRect(screen, px, py, panelW, panelH, Theme.PostBattlePanelBG, false)
	vector.StrokeRect(screen, px, py, panelW, panelH, 2, Theme.PostBattleBorder, false)

	ix := px + 16
	y := py + 14
	drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: panelW - 32, H: lineH * 1.2}, title, metrics, Theme.TextPrimary)
	y += lineH * 1.25
	drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: panelW - 32, H: lineH * 1.1}, sub, metrics, Theme.TextSecondary)
	y += lineH * 1.35
	DrawThinAccentLine(screen, ix, y, panelW-32)
	y += 10

	for _, ln := range lines {
		col := Theme.TextSecondary
		if strings.HasPrefix(ln, "—") || strings.Contains(ln, "Способности") || strings.HasPrefix(ln, "Повышение:") {
			col = Theme.TextMuted
		}
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: panelW - 32, H: lineH * 1.05}, ln, metrics, col)
		y += lineH * 1.08
	}

	if feedbackBanner != "" {
		y += 4
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: panelW - 32, H: lineH * 1.1}, feedbackBanner, metrics, Theme.TextSecondary)
		y += lineH * 1.12
	}

	y += 6
	drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: panelW - 32, H: lineH * 1.1}, inspectPromotionFooterHint(atCamp), metrics, Theme.TextMuted)
}

func inspectPromotionFooterHint(atCamp bool) string {
	if atCamp {
		return "I / Esc — закрыть · ↑↓ — другой · P — повышение (в лагере)"
	}
	return "I / Esc — закрыть · ↑↓ — другой · P — повышение только на лазурном лагере"
}

func inspectPrimaryTitle(h *hero.Hero, globalIdx, na int) string {
	if tpl, ok := unitdata.GetUnitTemplate(h.UnitID); ok {
		return tpl.DisplayName
	}
	// LEGACY: до unit_id — подпись рекрута или роль в отряде.
	if h.RecruitLabel != "" {
		return h.RecruitLabel
	}
	if globalIdx < na {
		return party.MemberRoleCaption(globalIdx)
	}
	return party.ReserveRowCaption(globalIdx - na)
}

func inspectSubtitle(h *hero.Hero, globalIdx, na int, inReserve bool) string {
	var parts []string
	if h.RecruitLabel != "" {
		parts = append(parts, h.RecruitLabel)
	}
	if inReserve {
		parts = append(parts, "Резерв · вне боя до вывода в строй")
	} else {
		parts = append(parts, party.FormationSlotCaption(globalIdx))
	}
	return strings.Join(parts, " · ")
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

func buildInspectLines(h *hero.Hero, inReserve bool, atCamp bool) []string {
	xpBonus := h.CombatExperience / hero.CombatXPStepsPerBasicAttackBonus
	effective := h.EffectiveBasicAttackBonusForCombat()
	baseBonus := h.BasicAttackBonus

	var status string
	switch {
	case inReserve:
		status = "Положение: резерв (не в бою)"
	case h.CurrentHP <= 0:
		status = "Положение: строй · не может сражаться (0 ОЗ)"
	default:
		status = "Положение: строй · готов к бою"
	}

	templateLines := templateInspectLines(h)

	healTotal := 2 + h.HealPower
	lines := []string{
		status,
	}
	lines = append(lines, templateLines...)
	lines = append(lines, hero.PromotionUILine(h, atCamp))
	lines = append(lines,
		fmt.Sprintf("ОЗ: %d / %d", h.CurrentHP, h.MaxHP),
		fmt.Sprintf("Атака %d · Защита %d · Инициатива %d", h.Attack, h.Defense, h.Initiative),
		fmt.Sprintf("Бонус базовой атаки: %d (награды %d + опыт %d)", effective, baseBonus, xpBonus),
		fmt.Sprintf("Боевой опыт: %d (каждые %d → +1 к бонусу базовой атаки в бою)", h.CombatExperience, hero.CombatXPStepsPerBasicAttackBonus),
		fmt.Sprintf("Лечение: итог восстановления %d ОЗ (база 2 + бонус %d)", healTotal, h.HealPower),
		"—",
		"Способности: " + abilityListRu(h.Abilities),
	)
	return lines
}

func templateInspectLines(h *hero.Hero) []string {
	if h == nil {
		return nil
	}
	tpl, ok := unitdata.GetUnitTemplate(h.UnitID)
	if !ok {
		return []string{
			"— Шаблон —",
			"Неизвестен (пустой или устаревший unit_id); ниже — текущее состояние бойца.",
		}
	}
	out := []string{
		"— Шаблон —",
		fmt.Sprintf("ID: %s", tpl.UnitID),
		fmt.Sprintf("Фракция: %s · Линия: %s · Tier %d",
			unitdata.FactionDisplayRU(tpl.FactionID),
			unitdata.LineDisplayRU(tpl.LineID),
			tpl.Tier),
		fmt.Sprintf("Архетип: %s · Роль: %s · Тип атаки: %s",
			tpl.ArchetypeID,
			roleLabelRu(tpl.Role),
			unitdata.AttackKindDisplayRU(tpl.AttackKind)),
	}
	if tpl.InspectNote != "" {
		out = append(out, tpl.InspectNote)
	}
	return out
}

func abilityListRu(ids []battlepkg.AbilityID) string {
	if len(ids) == 0 {
		return "—"
	}
	var parts []string
	for _, id := range ids {
		parts = append(parts, abilityNameRu(id))
	}
	return strings.Join(parts, ", ")
}

func abilityNameRu(id battlepkg.AbilityID) string {
	switch id {
	case battlepkg.AbilityBasicAttack:
		return "Базовая атака"
	case battlepkg.AbilityRangedAttack:
		return "Выстрел"
	case battlepkg.AbilityHeal:
		return "Лечение"
	case battlepkg.AbilityBuff:
		return "Усиление"
	default:
		return "?"
	}
}
