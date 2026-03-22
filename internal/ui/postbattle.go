package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// PostBattleRect — прямоугольник в координатах экрана (логические пиксели).
type PostBattleRect struct {
	X, Y, W, H float32
}

// PostBattleLayout — каноническая геометрия post-battle overlay (отрисовка и hit-test).
// Единственный источник истины: ComputePostBattleLayout.
type PostBattleLayout struct {
	ScreenW, ScreenH               int
	Tier                           ResolutionTier
	PanelX, PanelY, PanelW, PanelH float32
	InnerX, InnerY, InnerW         float32
	LineH                          float32
	Pad                            float32
	RowH                           float32
	RowGap                         float32
	ButtonH                        float32
	// RewardRowH — высота строки варианта награды (две строки текста: заголовок + описание).
	RewardRowH float32
	// RewardOptionRects — области клика и подсветки строк награды (совпадают с отрисовкой).
	RewardOptionRects []PostBattleRect
	// ResultContinueButton — шаг результата боя: явное продолжение (мышь + тот же layout, что и draw).
	ResultContinueButton PostBattleRect
	// RewardConfirmButton — шаг выбора награды: подтвердить текущий выбор (аналог Space/Enter).
	RewardConfirmButton PostBattleRect
}

const postBattleButtonW = float32(200)

// ComputePostBattleLayout вычисляет layout post-battle экрана для заданного размера окна.
// isRewardStep и optionCount задают высоту панели и число прямоугольников наград.
// resultSummaryLines — число строк сводки прогрессии на шаге результата (победа); иначе 0.
func ComputePostBattleLayout(screenW, screenH int, isRewardStep bool, optionCount int, resultSummaryLines int) PostBattleLayout {
	l := PostBattleLayout{
		ScreenW: screenW,
		ScreenH: screenH,
	}
	if screenW < 100 || screenH < 100 {
		return l
	}

	sl := BattleOverlayScreenLayout(screenW, screenH)
	l.Tier = sl.Tier
	lineH, pad, rowH, rowGap, buttonH := PostBattleMetrics(sl.Tier)
	panelW := PostBattlePanelMaxWidth(screenW, sl.Tier)
	if panelW > sl.Modal.W {
		panelW = sl.Modal.W
	}

	computePanelH := func(lh, p, rh, rg, bh float32) float32 {
		if isRewardStep {
			n := optionCount
			if n < 0 {
				n = 0
			}
			// Компактнее «шапка», строки награды выше (две линии внутри rh).
			innerH := lh*3.35 + float32(n)*(rh+rg) + lh + 6 + bh + 10
			return p*2 + innerH
		}
		ns := resultSummaryLines
		if ns < 0 {
			ns = 0
		}
		summaryBlock := float32(0)
		if ns > 0 {
			summaryBlock = lh * 1.06 * float32(ns)
		}
		// Заголовок + блок «ИТОГИ» + сводка + подсказка + кнопка (см. DrawPostBattleOverlay).
		return p*2 + lh*2.9 + summaryBlock + lh*1.15 + bh + 22
	}

	for iter := 0; iter < 40; iter++ {
		rhPanel := rowH
		if isRewardStep {
			rhPanel = lineH * 2.5
			if rhPanel < 48 {
				rhPanel = 48
			}
		}
		panelH := computePanelH(lineH, pad, rhPanel, rowGap, buttonH)
		rect := CenterPanelInModal(sl, panelW, panelH)
		if panelH <= rect.H+0.5 || iter == 39 {
			l.PanelX, l.PanelY, l.PanelW, l.PanelH = rect.X, rect.Y, rect.W, rect.H
			l.LineH = lineH
			l.Pad = pad
			l.RowH = rowH
			l.RowGap = rowGap
			l.ButtonH = buttonH
			l.InnerX = l.PanelX + pad
			l.InnerY = l.PanelY + pad
			l.InnerW = l.PanelW - pad*2
			break
		}
		scale := rect.H / panelH
		lineH *= scale
		pad *= scale
		if pad < 12 {
			pad = 12
		}
		if lineH < 14 {
			lineH = 14
		}
		rowH = lineH * 1.35
		if rowH < 24 {
			rowH = 24
		}
		rowGap = float32(3)
		if sl.Tier == TierLarge {
			rowGap = 5
		}
		buttonH = lineH * 1.55
		if buttonH < 30 {
			buttonH = 30
		}
	}

	lineH = l.LineH
	pad = l.Pad
	rowH = l.RowH
	rowGap = l.RowGap
	buttonH = l.ButtonH

	if isRewardStep {
		l.RewardRowH = lineH * 2.5
		if l.RewardRowH < 48 {
			l.RewardRowH = 48
		}
	}

	btnW := postBattleButtonW
	if btnW > l.InnerW-16 {
		btnW = l.InnerW - 16
	}
	if btnW < 100 {
		btnW = l.InnerW - 16
	}
	btnX := l.InnerX + (l.InnerW-btnW)*0.5

	if !isRewardStep {
		ns := resultSummaryLines
		if ns < 0 {
			ns = 0
		}
		summaryBlock := float32(0)
		if ns > 0 {
			summaryBlock = lineH * 1.06 * float32(ns)
		}
		hintY := l.InnerY + lineH*2.85 + summaryBlock + 6
		btnY := hintY + lineH*0.95 + 12
		l.ResultContinueButton = PostBattleRect{
			X: btnX,
			Y: btnY,
			W: btnW,
			H: buttonH,
		}
		return l
	}

	n := optionCount
	if n < 0 {
		n = 0
	}
	rh := l.RewardRowH
	if rh <= 0 {
		rh = rowH
	}
	firstY := l.InnerY + lineH*3.25
	l.RewardOptionRects = make([]PostBattleRect, n)
	for i := 0; i < n; i++ {
		y := firstY + float32(i)*(rh+rowGap)
		l.RewardOptionRects[i] = PostBattleRect{
			X: l.InnerX,
			Y: y - 2,
			W: l.InnerW,
			H: rh + 4,
		}
	}
	var hintY float32
	if n > 0 {
		yAfter := firstY + float32(n)*(rh+rowGap)
		hintY = yAfter + 4
	} else {
		hintY = l.InnerY + lineH*3.25 + 4
	}
	confirmY := hintY + lineH + 8
	l.RewardConfirmButton = PostBattleRect{
		X: btnX,
		Y: confirmY,
		W: btnW,
		H: buttonH,
	}
	return l
}

// RewardOptionIndexAt возвращает индекс строки награды под курсором или -1.
func (l PostBattleLayout) RewardOptionIndexAt(mx, my int) int {
	mxf := float32(mx)
	myf := float32(my)
	for i, r := range l.RewardOptionRects {
		if mxf >= r.X && mxf <= r.X+r.W && myf >= r.Y && myf <= r.Y+r.H {
			return i
		}
	}
	return -1
}

func postBattlePointInRect(mx, my int, r PostBattleRect) bool {
	if r.W <= 0 || r.H <= 0 {
		return false
	}
	mxf := float32(mx)
	myf := float32(my)
	return mxf >= r.X && mxf <= r.X+r.W && myf >= r.Y && myf <= r.Y+r.H
}

// HitResultContinue — клик по кнопке продолжения на экране результата.
func (l PostBattleLayout) HitResultContinue(mx, my int) bool {
	return postBattlePointInRect(mx, my, l.ResultContinueButton)
}

// HitRewardConfirm — клик по кнопке подтверждения выбранной награды.
func (l PostBattleLayout) HitRewardConfirm(mx, my int) bool {
	return postBattlePointInRect(mx, my, l.RewardConfirmButton)
}

// PostBattleParams — параметры для отрисовки post-battle overlay (game передаёт готовые строки).
type PostBattleParams struct {
	ResultText    string
	IsRewardStep  bool
	OptionLabels  []string
	OptionDescs   []string
	SelectedIndex int
	ScreenWidth   int
	ScreenHeight  int
	// ResultHintLine — подсказка под заголовком на шаге результата; пусто = дефолтная строка про Пробел/Enter.
	ResultHintLine string
	// VictorySummaryLines — компактная сводка (победа); между заголовком и подсказкой.
	VictorySummaryLines []string
	// RewardPreambleLine — одна строка над выбором награды (разделение с боевым опытом).
	RewardPreambleLine string
	// Кнопки (явный mouse path; Space/Enter остаётся альтернативой).
	ContinueButtonLabel string
	ConfirmRewardLabel  string
	HoverContinue       bool
	HoverRewardConfirm  bool
	// HoverRewardIndex — индекс строки награды под курсором (шаг награды); -1 = нет наведения.
	HoverRewardIndex int
}

// DrawPostBattleOverlay рисует полупрозрачный overlay: результат боя и (при победе) выбор награды.
func DrawPostBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p PostBattleParams) {
	if hudFace == nil || p.ScreenWidth < 100 || p.ScreenHeight < 100 {
		return
	}
	summaryN := len(p.VictorySummaryLines)
	if p.IsRewardStep {
		summaryN = 0
	}
	layout := ComputePostBattleLayout(p.ScreenWidth, p.ScreenHeight, p.IsRewardStep, len(p.OptionLabels), summaryN)
	w := float32(p.ScreenWidth)
	h := float32(p.ScreenHeight)
	vector.FillRect(screen, 0, 0, w, h, Theme.OverlayDimHeavy, false)

	drawPostBattleEventChrome(screen, layout.PanelX, layout.PanelY, layout.PanelW, layout.PanelH)

	innerX := layout.InnerX
	innerY := layout.InnerY
	innerW := layout.InnerW
	lineH := layout.LineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}

	if !p.IsRewardStep {
		vector.FillRect(screen, innerX, innerY, innerW, lineH*2.1, Theme.PostBattleTitleGlow, false)
	}

	rt := strings.TrimSpace(p.ResultText)
	if strings.HasPrefix(rt, "Победа") {
		drawPostBattleHeadlineScaled(screen, hudFace, innerX+6, innerY+6, innerW-12, rt)
	} else {
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 4, Y: innerY + 6, W: innerW - 8, H: lineH * 1.6}, rt, metrics, Theme.TextHeadline)
	}
	vector.FillRect(screen, innerX+4, innerY+lineH*1.95, innerW-8, 2, Theme.PanelTitleSep, false)

	if !p.IsRewardStep {
		sec := rect{X: innerX + 4, Y: innerY + lineH*2.15, W: innerW - 8, H: lineH * 0.85}
		drawSingleLineInRect(screen, hudFace, sec, "ИТОГИ", metrics, Theme.TextMuted)
		y := innerY + lineH*2.85
		for _, s := range p.VictorySummaryLines {
			if s == "" {
				continue
			}
			drawSingleLineInRect(screen, hudFace, rect{X: innerX + 4, Y: y, W: innerW - 8, H: lineH * 1.08}, s, metrics, Theme.TextSecondary)
			y += lineH * 1.06
		}
		hint := p.ResultHintLine
		if hint == "" {
			hint = "Enter / Пробел — далее"
		}
		hintY := y + 6
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 4, Y: hintY, W: innerW - 8, H: lineH * 0.95}, hint, metrics, Theme.TextMuted)
		lbl := p.ContinueButtonLabel
		if lbl == "" {
			lbl = "Далее"
		}
		rc := layout.ResultContinueButton
		drawPostBattlePrimaryButton(screen, hudFace, rc, lbl, p.HoverContinue, metrics)
		return
	}

	head := rect{X: innerX + 4, Y: innerY + 6, W: innerW - 8, H: lineH * 1.2}
	drawSingleLineInRect(screen, hudFace, head, rt, metrics, Theme.TextHeadline)
	sub := p.RewardPreambleLine
	if sub == "" {
		sub = "Отдельно от опыта отряда — усиление лидера."
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX + 4, Y: innerY + lineH*1.45, W: innerW - 8, H: lineH}, sub, metrics, Theme.TextMuted)
	drawSingleLineInRect(screen, hudFace, rect{X: innerX + 4, Y: innerY + lineH*2.35, W: innerW - 8, H: lineH * 0.9}, "ВЫБЕРИТЕ НАГРАДУ", metrics, Theme.TextSecondary)
	vector.FillRect(screen, innerX+4, innerY+lineH*3.05, innerW-8, 2, Theme.PanelTitleSep, false)

	y := innerY + lineH*3.25
	rowRH := layout.RewardRowH
	if rowRH <= 0 {
		rowRH = layout.RowH
	}
	rowRG := layout.RowGap
	line1H := rowRH * 0.48
	line2H := rowRH * 0.42
	for i := 0; i < len(p.OptionLabels); i++ {
		label := p.OptionLabels[i]
		desc := ""
		if i < len(p.OptionDescs) {
			desc = strings.TrimSpace(p.OptionDescs[i])
		}
		textPri := Theme.TextPrimary
		textSec := Theme.TextSecondary
		if i < len(layout.RewardOptionRects) {
			r := layout.RewardOptionRects[i]
			switch {
			case i == p.SelectedIndex:
				textPri = Theme.TextHeadline
				textSec = Theme.TextSecondary
				vector.FillRect(screen, r.X, r.Y, r.W, r.H, Theme.PostBattleRowSelect, false)
				vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 2, Theme.PostBattleRowBrd, false)
				vector.FillRect(screen, r.X, r.Y, 5, r.H, Theme.AccentStrip, false)
			case p.HoverRewardIndex >= 0 && i == p.HoverRewardIndex:
				textPri = Theme.TextPrimary
				vector.FillRect(screen, r.X, r.Y, r.W, r.H, Theme.AbilityHoverBG, false)
				vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, Theme.HoverTarget, false)
			default:
				vector.FillRect(screen, r.X, r.Y, r.W, r.H, Theme.RosterCardContentWell, false)
				vector.FillRect(screen, r.X, r.Y, 3, r.H, Theme.ExploreModuleEdge, false)
			}
		}
		lbl := PrimaryLine(hudFace, label, innerW-28)
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 14, Y: y + 4, W: innerW - 28, H: line1H}, lbl, metrics, textPri)
		if desc != "" {
			drawSingleLineInRect(screen, hudFace, rect{X: innerX + 14, Y: y + 6 + line1H, W: innerW - 28, H: line2H}, PrimaryLine(hudFace, desc, innerW-28), metrics, textSec)
		}
		y += rowRH + rowRG
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX + 4, Y: y + 4, W: innerW - 8, H: lineH * 0.9}, "↑↓ выбор · Enter — подтвердить", metrics, Theme.TextMuted)
	cl := p.ConfirmRewardLabel
	if cl == "" {
		cl = "Взять награду"
	}
	drawPostBattlePrimaryButton(screen, hudFace, layout.RewardConfirmButton, cl, p.HoverRewardConfirm, metrics)
}

func drawPostBattleHeadlineScaled(screen *ebiten.Image, hudFace *text.GoTextFace, x, y, maxW float32, s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	const scale = 1.26
	avail := maxW / scale
	s = fitTextToWidth(hudFace, s, avail)
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(Theme.TextHeadline)
	text.Draw(screen, s, hudFace, op)
}

func drawPostBattlePrimaryButton(screen *ebiten.Image, hudFace *text.GoTextFace, r PostBattleRect, label string, hover bool, metrics battlepkg.HUDMetrics) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	fill := Theme.ButtonBG
	border := Theme.ButtonBorder
	if hover {
		fill = Theme.ButtonHoverBG
		border = Theme.AccentStrip
	}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 2, border, false)
	vector.FillRect(screen, r.X, r.Y, 5, r.H, Theme.AccentStrip, false)
	rr := rect{X: r.X, Y: r.Y, W: r.W, H: r.H}
	drawSingleLineInRect(screen, hudFace, rr, label, metrics, Theme.TextPrimary)
}
