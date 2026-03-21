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

// battleInspectPanelWidth — шире обычной карточки: место под крупный портрет battle inspect.
func battleInspectPanelWidth(screenW int) float32 {
	w := float32(620)
	if float32(screenW)-40 < w {
		w = float32(screenW) - 40
	}
	return w
}

// DrawBattleInspectOverlay — карточка по ПКМ в бою: союзник (hero + текущее ОЗ) или враг.
// promotionHeadline — готовность повышения (в бою лагерь недоступен — строка обычно про «нужен лагерь»).
func DrawBattleInspectOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, u *battlepkg.CombatUnit, screenW, screenH int, trainingMarks int, promoteTargets []string, promoteCosts []int, branchIdx int, promotionHeadline string) {
	if hudFace == nil || p == nil || u == nil {
		return
	}
	sw := float32(screenW)
	sh := float32(screenH)
	panelW := battleInspectPanelWidth(screenW)

	var m InspectCardModel
	if u.Side == battlepkg.TeamPlayer && u.Origin.PartyActiveIndex >= 0 && u.Origin.PartyActiveIndex < len(p.Active) {
		idx := u.Origin.PartyActiveIndex
		h := p.HeroAtGlobalIndex(idx)
		if h == nil {
			return
		}
		snap := heroSnapshotBattleInspect(h, u)
		m = buildBattleInspectAllyModel(&snap, u, idx, len(p.Active), trainingMarks, promoteTargets, promoteCosts, branchIdx, promotionHeadline)
	} else {
		m = buildBattleInspectEnemyModel(u)
	}

	panelH := EstimateInspectCardHeight(m)
	px := (sw - panelW) / 2
	py := (sh - panelH) * 0.45
	if py < 16 {
		py = 16
	}

	DrawInspectCardChrome(screen, px, py, panelW, panelH, m.IsEnemy)
	DrawInspectCardContent(screen, hudFace, px, py, panelW, m)
}

func battleInspectCardFooter() string {
	return "Esc или ПКМ мимо — закрыть · ПКМ по другому юниту — переключить"
}

func battlePortraitImageForInspectAlly(h *hero.Hero) *ebiten.Image {
	if h == nil {
		return nil
	}
	if h.UnitID == unitdata.EmpireWarriorSquire {
		return SquirePortraitImage()
	}
	return nil
}

func battlePortraitImageForInspectEnemy(u *battlepkg.CombatUnit) *ebiten.Image {
	if u == nil {
		return nil
	}
	if u.Def.TemplateUnitID == unitdata.EmpireWarriorSquire {
		return SquirePortraitImage()
	}
	return nil
}

func loreParagraphBattleInspectAlly(h *hero.Hero) string {
	if h == nil {
		return ""
	}
	if tpl, ok := unitdata.GetUnitTemplate(h.UnitID); ok && tpl.InspectNote != "" {
		return tpl.InspectNote
	}
	if h.RecruitLabel != "" {
		return fmt.Sprintf("%s — черновая запись в отряде; полная хроника появится позже.", h.RecruitLabel)
	}
	return "Черновая заглушка: для этого бойца ещё нет готовой записи в хронике кампании."
}

func loreParagraphBattleInspectEnemy(u *battlepkg.CombatUnit) string {
	if u == nil {
		return ""
	}
	if u.Def.TemplateUnitID != "" {
		if tpl, ok := unitdata.GetUnitTemplate(u.Def.TemplateUnitID); ok && tpl.InspectNote != "" {
			return tpl.InspectNote
		}
	}
	return "Разведданные неполные. Облик и происхождение противника будут уточнены в следующих билдах."
}

func heroSnapshotBattleInspect(h *hero.Hero, u *battlepkg.CombatUnit) hero.Hero {
	if h == nil || u == nil {
		return hero.Hero{}
	}
	snap := *h
	snap.CurrentHP = u.State.HP
	if snap.CurrentHP < 0 {
		snap.CurrentHP = 0
	}
	snap.MaxHP = u.MaxHP()
	return snap
}

func buildBattleInspectAllyModel(h *hero.Hero, u *battlepkg.CombatUnit, idx, na int, trainingMarks int, promoteTargets []string, promoteCosts []int, branchIdx int, promotionHeadline string) InspectCardModel {
	m := InspectCardModel{
		BattlePortraitLayout: true,
		BattlePortrait:       battlePortraitImageForInspectAlly(h),
		LoreParagraph:        loreParagraphBattleInspectAlly(h),
		RoleIcon:             InspectRoleIconFromHero(h),
		Title:                inspectPrimaryTitle(h, idx, na),
		ContextLine:          fmt.Sprintf("В бою · %s", party.FormationSlotCaption(idx)),
		HPCur:                h.CurrentHP,
		HPMax:                h.MaxHP,
		Alive:                u.IsAlive(),
		IsEnemy:              false,
		Footer:               battleInspectCardFooter(),
	}
	m.Badges = compactTierRangeBadgesFromHero(h)
	m.ProfileLines = templateProfileShortLines(h)
	healTotal := 2 + h.HealPower
	m.StatsLine = fmt.Sprintf("Атака %d · Защита %d · Инициатива %d · Лечение +%d", h.Attack, h.Defense, h.Initiative, healTotal)
	m.ExtraStatLine = ""

	m.AbilityLines = abilityLinesBullet(h.Abilities)

	atCamp := false
	lines := battleInspectAllyProgressLines(h, atCamp, trainingMarks, promoteTargets, promoteCosts, branchIdx)
	if strings.TrimSpace(promotionHeadline) != "" {
		lines = append([]string{strings.TrimSpace(promotionHeadline)}, lines...)
	}
	m.ProgressLines = lines
	return m
}

func buildBattleInspectEnemyModel(u *battlepkg.CombatUnit) InspectCardModel {
	m := InspectCardModel{
		BattlePortraitLayout: true,
		BattlePortrait:       battlePortraitImageForInspectEnemy(u),
		LoreParagraph:        loreParagraphBattleInspectEnemy(u),
		RoleIcon:             InspectRoleIconFromCombatUnit(u),
		Title:                u.Name(),
		ContextLine:          "Противник",
		HPCur:                u.State.HP,
		HPMax:                u.MaxHP(),
		Alive:                u.IsAlive(),
		IsEnemy:              true,
		Footer:               battleInspectCardFooter(),
	}
	m.Badges = compactTierRangeBadgesFromEnemy(u)
	m.ProfileLines = enemyProfileLines(u)
	m.StatsLine = fmt.Sprintf("Атака %d · Защита %d · Инициатива %d", u.Attack(), u.Defense(), u.Initiative())
	m.ExtraStatLine = ""
	m.AbilityLines = abilityLinesBullet(u.Abilities())
	return m
}

func compactTierRangeBadgesFromHero(h *hero.Hero) []string {
	if h == nil {
		return nil
	}
	tpl, ok := unitdata.GetUnitTemplate(h.UnitID)
	if !ok {
		return nil
	}
	return []string{fmt.Sprintf("Ранг %d · %s", tpl.Tier, attackKindShortRu(tpl.AttackKind))}
}

func compactTierRangeBadgesFromEnemy(u *battlepkg.CombatUnit) []string {
	if u == nil {
		return nil
	}
	if u.Def.TemplateUnitID != "" {
		if tpl, ok := unitdata.GetUnitTemplate(u.Def.TemplateUnitID); ok {
			return []string{fmt.Sprintf("Ранг %d · %s", tpl.Tier, attackKindShortRu(tpl.AttackKind))}
		}
	}
	kind := attackKindShortRuFromLegacy(u.Def.IsRanged)
	return []string{kind}
}

func attackKindShortRu(k unitdata.AttackKind) string {
	switch k {
	case unitdata.AttackMelee:
		return "ближний бой"
	case unitdata.AttackRanged:
		return "дальний бой"
	case unitdata.AttackHeal:
		return "поддержка"
	default:
		return "ближний бой"
	}
}

func attackKindShortRuFromLegacy(isRanged bool) string {
	if isRanged {
		return "дальний бой"
	}
	return "ближний бой"
}

func templateProfileShortLines(h *hero.Hero) []string {
	if h == nil {
		return nil
	}
	tpl, ok := unitdata.GetUnitTemplate(h.UnitID)
	if !ok {
		if h.RecruitLabel != "" {
			return []string{h.RecruitLabel}
		}
		return []string{"Профиль недоступен — ниже актуальные показатели."}
	}
	// Имя уже в заголовке карточки — здесь линия фракции и линии без повтора DisplayName.
	lines := []string{fmt.Sprintf("%s · %s", unitdata.LineDisplayRU(tpl.LineID), unitdata.FactionDisplayRU(tpl.FactionID))}
	if tpl.InspectNote != "" {
		n := tpl.InspectNote
		if len([]rune(n)) > 72 {
			rs := []rune(n)
			n = string(rs[:69]) + "…"
		}
		lines = append(lines, n)
	}
	return lines
}

func enemyProfileLines(u *battlepkg.CombatUnit) []string {
	if u == nil {
		return nil
	}
	if u.Def.TemplateUnitID != "" {
		if tpl, ok := unitdata.GetUnitTemplate(u.Def.TemplateUnitID); ok {
			lines := []string{fmt.Sprintf("%s · %s", unitdata.LineDisplayRU(tpl.LineID), unitdata.FactionDisplayRU(tpl.FactionID))}
			if tpl.InspectNote != "" {
				n := tpl.InspectNote
				if len([]rune(n)) > 72 {
					rs := []rune(n)
					n = string(rs[:69]) + "…"
				}
				lines = append(lines, n)
			}
			return lines
		}
		return []string{"Сведения о профиле временно недоступны."}
	}
	return []string{fmt.Sprintf("%s · %s", roleLabelRu(u.Def.Role), archetypeShort(u.Def.ArchetypeID))}
}

func archetypeShort(s string) string {
	if s == "" {
		return "—"
	}
	if len([]rune(s)) > 28 {
		rs := []rune(s)
		return string(rs[:25]) + "…"
	}
	return s
}

func abilityLinesBullet(ids []battlepkg.AbilityID) []string {
	if len(ids) == 0 {
		return []string{"нет способностей"}
	}
	var out []string
	for _, id := range ids {
		out = append(out, battlepkg.PlayerAbilityLabelRU(id))
	}
	return out
}

const maxBattleInspectProgressLines = 7

func battleInspectAllyProgressLines(h *hero.Hero, atCamp bool, trainingMarks int, promoteTargets []string, promoteCosts []int, branchIdx int) []string {
	if h == nil {
		return nil
	}
	var out []string
	out = append(out, FormatCombatXPInspectLines(h)...)
	promo := inspectPromotionLines(h, atCamp, trainingMarks, promoteTargets, promoteCosts, branchIdx)
	for _, ln := range promo {
		out = append(out, ln)
		if len(out) >= maxBattleInspectProgressLines {
			break
		}
	}
	return out
}
