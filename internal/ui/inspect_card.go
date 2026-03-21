package ui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// InspectCardModel — единая структурированная карточка inspect (бой и состав).
type InspectCardModel struct {
	RoleIcon InspectRoleIcon

	Title       string
	Badges      []string
	ContextLine string

	HPCur, HPMax int
	Alive        bool
	IsEnemy      bool

	ProfileLines []string

	StatsLine     string
	ExtraStatLine string

	AbilityLines []string

	ProgressLines []string

	Footer         string
	FeedbackBanner string

	// Battle inspect overlay only (formation оставляет нули).
	BattlePortraitLayout bool           // двухколоночный блок с крупным портретом
	BattlePortrait       *ebiten.Image  // nil — рамка-заглушка
	LoreParagraph        string         // секция «Описание» (лор / заглушка)
}

const inspectCardPanelW = float32(480)
const inspectCardIconSize = float32(30)

// battleInspectPortraitBoxW/H — область под портрет в battle inspect.
const battleInspectPortraitBoxW = float32(252)
const battleInspectPortraitBoxH = float32(280)

// DefaultInspectCardPanelWidth — ширина карточки (единая для battle / formation).
func DefaultInspectCardPanelWidth(screenW int) float32 {
	w := inspectCardPanelW
	if float32(screenW)-40 < w {
		w = float32(screenW) - 40
	}
	return w
}

// DrawInspectCardChrome — фон, рамка, боковая полоса ally/enemy, верхний акцент.
func DrawInspectCardChrome(screen *ebiten.Image, px, py, panelW, panelH float32, isEnemy bool) {
	vector.FillRect(screen, px, py, panelW, panelH, Theme.PostBattlePanelBG, false)
	vector.StrokeRect(screen, px, py, panelW, panelH, 2, Theme.PostBattleBorder, false)
	strip := Theme.AllyAccent
	if isEnemy {
		strip = Theme.EnemyAccent
	}
	vector.FillRect(screen, px, py, 4, panelH, strip, false)
}

// EstimateInspectCardHeight — высота контента карточки (без внешних полей экрана).
func EstimateInspectCardHeight(m InspectCardModel) float32 {
	lineH := uiLineH
	var h float32 = 14 + 3
	if m.BattlePortraitLayout {
		h += estimateBattleInspectHeaderRowHeight(m, lineH)
	} else {
		h += maxF(lineH*1.95, inspectCardIconSize) + 6
		for range m.Badges {
			h += lineH * 1.12
		}
	}
	h += lineH*1.2 + 8
	h += lineH*1.35 + 4 + 6 + 10
	if m.BattlePortraitLayout && strings.TrimSpace(m.LoreParagraph) != "" {
		h += lineH*0.9 + 8 + lineH*1.15*3 + 10
	}
	h += inspectCardSectionBlockHeight(len(m.ProfileLines), lineH)
	if m.StatsLine != "" {
		h += lineH*0.9 + 6 + lineH*1.1
		if m.ExtraStatLine != "" {
			h += lineH * 1.08
		}
		h += 10
	}
	h += inspectCardSectionBlockHeight(len(m.AbilityLines), lineH)
	if len(m.ProgressLines) > 0 {
		h += inspectCardSectionBlockHeight(len(m.ProgressLines), lineH)
	}
	if m.FeedbackBanner != "" {
		h += lineH*1.15 + 6
	}
	h += lineH*1.15 + 14
	return h
}

func inspectCardSectionBlockHeight(nLines int, lineH float32) float32 {
	if nLines == 0 {
		return 0
	}
	h := lineH*0.9 + 6
	h += float32(nLines) * lineH * 1.1
	h += 10
	return h
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// DrawInspectCardContent — содержимое карточки (после DrawInspectCardChrome).
func DrawInspectCardContent(screen *ebiten.Image, hudFace *text.GoTextFace, px, py, panelW float32, m InspectCardModel) {
	lineH := uiLineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}
	ix := px + 16
	innerW := panelW - 32
	y := py + 14

	vector.FillRect(screen, px, py, panelW, 3, Theme.AccentStrip, false)

	if m.BattlePortraitLayout {
		y = drawBattleInspectPortraitHeader(screen, hudFace, ix, y, innerW, m)
	} else {
		iconCol := Theme.AccentStrip
		titleRowH := maxF(lineH*1.95, inspectCardIconSize)
		DrawInspectRoleIcon(screen, ix, y+(titleRowH-inspectCardIconSize)*0.5, inspectCardIconSize, m.RoleIcon, iconCol)

		titleMetrics := battlepkg.HUDMetrics{LineH: lineH * 1.12}
		titleTextW := innerW - inspectCardIconSize - 8
		drawSingleLineInRect(screen, hudFace, rect{X: ix + inspectCardIconSize + 8, Y: y, W: titleTextW, H: titleRowH}, m.Title, titleMetrics, Theme.TextPrimary)
		y += titleRowH + 6
	}

	if !m.BattlePortraitLayout {
		for _, b := range m.Badges {
			if strings.TrimSpace(b) == "" {
				continue
			}
			drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.05}, b, metrics, Theme.TextSecondary)
			y += lineH * 1.12
		}
	}

	drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.05}, m.ContextLine, metrics, Theme.TextMuted)
	y += lineH*1.15 + 8

	hpStr := fmt.Sprintf("ОЗ  %d / %d", m.HPCur, m.HPMax)
	hpCol := Theme.TextSuccess
	if m.HPCur*2 < m.HPMax && m.HPCur > 0 {
		hpCol = Theme.TextSecondary
	}
	if m.HPCur <= 0 || !m.Alive {
		hpCol = Theme.TextDanger
	}
	hpMetrics := battlepkg.HUDMetrics{LineH: lineH * 1.18}
	drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.4}, hpStr, hpMetrics, hpCol)
	y += lineH*1.4 + 4
	DrawHPBarMicro(screen, ix, y, innerW, 6, m.HPCur, m.HPMax, m.Alive && m.HPCur > 0, m.IsEnemy)
	y += 6 + 12

	if m.BattlePortraitLayout && strings.TrimSpace(m.LoreParagraph) != "" {
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 0.9}, "Описание", metrics, Theme.TextMuted)
		y += lineH * 0.9
		DrawThinAccentLine(screen, ix, y, innerW)
		y += 8
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 3.5}, m.LoreParagraph, metrics, Theme.TextSecondary)
		y += lineH*3.5 + 10
	}

	y = drawInspectCardSection(screen, hudFace, ix, y, innerW, "Профиль", m.ProfileLines, metrics)

	if m.StatsLine != "" {
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 0.9}, "Показатели", metrics, Theme.TextMuted)
		y += lineH * 0.9
		DrawThinAccentLine(screen, ix, y, innerW)
		y += 8
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.1}, m.StatsLine, metrics, Theme.TextPrimary)
		y += lineH * 1.12
		if m.ExtraStatLine != "" {
			drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.05}, m.ExtraStatLine, metrics, Theme.TextSecondary)
			y += lineH * 1.08
		}
		y += 8
	}

	y = drawInspectCardSection(screen, hudFace, ix, y, innerW, "Способности", m.AbilityLines, metrics)

	if len(m.ProgressLines) > 0 {
		y = drawInspectCardSection(screen, hudFace, ix, y, innerW, "Развитие", m.ProgressLines, metrics)
	}

	if m.FeedbackBanner != "" {
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.12}, m.FeedbackBanner, metrics, Theme.TextSecondary)
		y += lineH * 1.15 + 4
	}

	drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.1}, m.Footer, metrics, Theme.TextMuted)
}

func drawInspectCardSection(screen *ebiten.Image, hudFace *text.GoTextFace, ix, y, innerW float32, heading string, lines []string, metrics battlepkg.HUDMetrics) float32 {
	lineH := metrics.LineH
	if heading == "" || len(lines) == 0 {
		return y
	}
	drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 0.9}, heading, metrics, Theme.TextMuted)
	y += lineH * 0.9
	DrawThinAccentLine(screen, ix, y, innerW)
	y += 8
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		drawSingleLineInRect(screen, hudFace, rect{X: ix, Y: y, W: innerW, H: lineH * 1.08}, ln, metrics, Theme.TextSecondary)
		y += lineH * 1.1
	}
	y += 10
	return y
}

func estimateBattleInspectHeaderRowHeight(m InspectCardModel, lineH float32) float32 {
	var badgeH float32
	for _, b := range m.Badges {
		if strings.TrimSpace(b) == "" {
			continue
		}
		badgeH += lineH * 1.12
	}
	titleH := maxF(lineH*1.95, inspectCardIconSize)
	rightCol := titleH + 6 + badgeH + 4
	row := maxF(battleInspectPortraitBoxH, rightCol)
	return row + 6
}

// drawBattleInspectPortraitHeader — крупный портрет слева, роль/имя/бейджи справа.
func drawBattleInspectPortraitHeader(screen *ebiten.Image, hudFace *text.GoTextFace, ix, y, innerW float32, m InspectCardModel) float32 {
	lineH := uiLineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}
	titleMetrics := battlepkg.HUDMetrics{LineH: lineH * 1.12}
	iconCol := Theme.AccentStrip

	drawBattlePortraitCell(screen, hudFace, ix, y, battleInspectPortraitBoxW, battleInspectPortraitBoxH, m.BattlePortrait)

	textX := ix + battleInspectPortraitBoxW + 12
	textW := innerW - battleInspectPortraitBoxW - 12
	ty := y + 4
	DrawInspectRoleIcon(screen, textX, ty, inspectCardIconSize, m.RoleIcon, iconCol)
	titleRowH := maxF(lineH*1.95, inspectCardIconSize)
	drawSingleLineInRect(screen, hudFace, rect{X: textX + inspectCardIconSize + 8, Y: ty, W: textW - inspectCardIconSize - 8, H: titleRowH}, m.Title, titleMetrics, Theme.TextPrimary)
	badgeY := ty + titleRowH + 6
	for _, b := range m.Badges {
		if strings.TrimSpace(b) == "" {
			continue
		}
		drawSingleLineInRect(screen, hudFace, rect{X: textX, Y: badgeY, W: textW, H: lineH * 1.05}, b, metrics, Theme.TextSecondary)
		badgeY += lineH * 1.12
	}
	headerBottom := y + battleInspectPortraitBoxH
	if badgeY+8 > headerBottom {
		headerBottom = badgeY + 8
	}
	return headerBottom + 8
}

func drawBattlePortraitCell(screen *ebiten.Image, hudFace *text.GoTextFace, x, y, w, h float32, img *ebiten.Image) {
	vector.FillRect(screen, x, y, w, h, Theme.PanelBGDeep, false)
	vector.StrokeRect(screen, x, y, w, h, 2, Theme.PanelBorder, false)
	pad := float32(4)
	lineH := uiLineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}
	if img == nil {
		drawSingleLineInRect(screen, hudFace, rect{X: x + pad, Y: y + h*0.42, W: w - 2*pad, H: lineH * 1.3}, "Портрет · скоро", metrics, Theme.TextMuted)
		return
	}
	iw := float32(img.Bounds().Dx())
	ih := float32(img.Bounds().Dy())
	if iw <= 0 || ih <= 0 {
		return
	}
	scale := (w - 2*pad) / iw
	if sh := (h - 2*pad) / ih; sh < scale {
		scale = sh
	}
	dw := iw * scale
	dh := ih * scale
	offX := x + (w-dw)/2
	offY := y + (h-dh)/2
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(offX), float64(offY))
	screen.DrawImage(img, op)
}

// FlattenInspectCardText склеивает текст карточки для тестов (без иконки).
func FlattenInspectCardText(m InspectCardModel) string {
	var b strings.Builder
	b.WriteString(m.Title)
	b.WriteString(m.ContextLine)
	for _, s := range m.Badges {
		b.WriteString(s)
	}
	fmt.Fprintf(&b, "%d%d", m.HPCur, m.HPMax)
	for _, s := range m.ProfileLines {
		b.WriteString(s)
	}
	b.WriteString(m.StatsLine)
	b.WriteString(m.ExtraStatLine)
	for _, s := range m.AbilityLines {
		b.WriteString(s)
	}
	for _, s := range m.ProgressLines {
		b.WriteString(s)
	}
	b.WriteString(m.Footer)
	b.WriteString(m.FeedbackBanner)
	b.WriteString(m.LoreParagraph)
	return b.String()
}
