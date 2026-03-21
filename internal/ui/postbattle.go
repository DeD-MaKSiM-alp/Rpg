package ui

import (
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
	PanelX, PanelY, PanelW, PanelH float32
	InnerX, InnerY, InnerW         float32
	LineH                          float32
	Pad                            float32
	// RewardOptionRects — области клика и подсветки строк награды (совпадают с отрисовкой).
	RewardOptionRects []PostBattleRect
	// ResultContinueButton — шаг результата боя: явное продолжение (мышь + тот же layout, что и draw).
	ResultContinueButton PostBattleRect
	// RewardConfirmButton — шаг выбора награды: подтвердить текущий выбор (аналог Space/Enter).
	RewardConfirmButton PostBattleRect
}

// postBattle metrics — одна точка правды для чисел layout (draw + hit-test).
const (
	postBattleLineH   = float32(22)
	postBattlePad     = float32(24)
	postBattlePanelW  = float32(400)
	postBattleRowH    = float32(32)
	postBattleRowGap  = float32(4)
	postBattleButtonH = float32(36)
	postBattleButtonW = float32(200)
)

// ComputePostBattleLayout вычисляет layout post-battle экрана для заданного размера окна.
// isRewardStep и optionCount задают высоту панели и число прямоугольников наград.
// resultSummaryLines — число строк сводки прогрессии на шаге результата (победа); иначе 0.
func ComputePostBattleLayout(screenW, screenH int, isRewardStep bool, optionCount int, resultSummaryLines int) PostBattleLayout {
	l := PostBattleLayout{
		ScreenW: screenW,
		ScreenH: screenH,
		LineH:   postBattleLineH,
		Pad:     postBattlePad,
	}
	if screenW < 100 || screenH < 100 {
		return l
	}
	w := float32(screenW)
	h := float32(screenH)

	panelW := postBattlePanelW
	if panelW > w-postBattlePad*2 {
		panelW = w - postBattlePad*2
	}
	lineH := postBattleLineH

	var panelH float32
	if isRewardStep {
		n := optionCount
		if n < 0 {
			n = 0
		}
		// Заголовок + две строки пояснения + строки награды + подсказка + кнопка «Подтвердить».
		innerH := lineH*4.5 + float32(n)*(postBattleRowH+postBattleRowGap) + 4 + lineH + 8 + postBattleButtonH + 8
		panelH = postBattlePad*2 + innerH
	} else {
		// Результат боя: заголовок + опционально сводка + подсказка + кнопка.
		ns := resultSummaryLines
		if ns < 0 {
			ns = 0
		}
		summaryBlock := float32(0)
		if ns > 0 {
			summaryBlock = lineH*1.05*float32(ns) + 8
		}
		panelH = postBattlePad*2 + lineH*1.5 + summaryBlock + lineH + 12 + postBattleButtonH + 16
	}
	panelX := (w - panelW) / 2
	panelY := (h - panelH) / 2

	l.PanelX, l.PanelY, l.PanelW, l.PanelH = panelX, panelY, panelW, panelH
	l.InnerX = panelX + postBattlePad
	l.InnerY = panelY + postBattlePad
	l.InnerW = panelW - postBattlePad*2

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
			summaryBlock = lineH*1.05*float32(ns) + 8
		}
		l.ResultContinueButton = PostBattleRect{
			X: btnX,
			Y: l.InnerY + lineH*1.5 + summaryBlock + lineH + 12,
			W: btnW,
			H: postBattleButtonH,
		}
		return l
	}

	n := optionCount
	if n < 0 {
		n = 0
	}
	firstY := l.InnerY + lineH*4.0
	l.RewardOptionRects = make([]PostBattleRect, n)
	for i := 0; i < n; i++ {
		y := firstY + float32(i)*(postBattleRowH+postBattleRowGap)
		l.RewardOptionRects[i] = PostBattleRect{
			X: l.InnerX,
			Y: y - 2,
			W: l.InnerW,
			H: postBattleRowH + 4,
		}
	}
	var hintY float32
	if n > 0 {
		yAfter := firstY + float32(n)*(postBattleRowH+postBattleRowGap)
		hintY = yAfter + 4
	} else {
		hintY = l.InnerY + lineH*4.0 + 4
	}
	confirmY := hintY + lineH + 8
	l.RewardConfirmButton = PostBattleRect{
		X: btnX,
		Y: confirmY,
		W: btnW,
		H: postBattleButtonH,
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
	vector.FillRect(screen, 0, 0, w, h, Theme.OverlayDim, false)

	drawUnifiedModalPanelChrome(screen, layout.PanelX, layout.PanelY, layout.PanelW, layout.PanelH)

	innerX := layout.InnerX
	innerY := layout.InnerY
	innerW := layout.InnerW
	lineH := layout.LineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}

	// Result line
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY, W: innerW, H: lineH * 1.5}, p.ResultText, metrics, Theme.TextPrimary)
	DrawThinAccentLine(screen, innerX+4, innerY+lineH*1.48, innerW-8)

	if !p.IsRewardStep {
		y := innerY + lineH*1.55
		for _, s := range p.VictorySummaryLines {
			if s == "" {
				continue
			}
			drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y, W: innerW, H: lineH * 1.05}, s, metrics, Theme.TextSecondary)
			y += lineH * 1.08
		}
		hint := p.ResultHintLine
		if hint == "" {
			hint = "Пробел / Enter — продолжить или кнопка ниже"
		}
		hintY := innerY + lineH*1.5
		if len(p.VictorySummaryLines) > 0 {
			hintY += lineH*1.05*float32(len(p.VictorySummaryLines)) + 8
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: hintY, W: innerW, H: lineH}, hint, metrics, Theme.TextMuted)
		lbl := p.ContinueButtonLabel
		if lbl == "" {
			lbl = "Продолжить"
		}
		rc := layout.ResultContinueButton
		drawPostBattlePrimaryButton(screen, hudFace, rc, lbl, p.HoverContinue, metrics)
		return
	}

	sub := p.RewardPreambleLine
	if sub == "" {
		sub = "Награда лидеру — отдельно от боевого опыта отряда."
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*2, W: innerW, H: lineH}, sub, metrics, Theme.TextMuted)
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*3.1, W: innerW, H: lineH}, "Выберите вариант:", metrics, Theme.TextSecondary)
	DrawThinAccentLine(screen, innerX+4, innerY+lineH*4.05, innerW-8)

	y := innerY + lineH*4.0
	for i := 0; i < len(p.OptionLabels); i++ {
		rowH := postBattleRowH
		label := p.OptionLabels[i]
		if i < len(p.OptionDescs) && p.OptionDescs[i] != "" {
			label = label + " — " + p.OptionDescs[i]
		}
		textCol := Theme.TextSecondary
		if i < len(layout.RewardOptionRects) {
			r := layout.RewardOptionRects[i]
			switch {
			case i == p.SelectedIndex:
				textCol = Theme.TextPrimary
				vector.FillRect(screen, r.X, r.Y, r.W, r.H, Theme.PostBattleRowSelect, false)
				vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1.25, Theme.PostBattleRowBrd, false)
				vector.FillRect(screen, r.X, r.Y, 4, r.H, Theme.AccentStrip, false)
			case p.HoverRewardIndex >= 0 && i == p.HoverRewardIndex:
				textCol = Theme.TextPrimary
				vector.FillRect(screen, r.X, r.Y, r.W, r.H, Theme.AbilityHoverBG, false)
				vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, Theme.HoverTarget, false)
			default:
				vector.FillRect(screen, r.X, r.Y, r.W, r.H, Theme.RosterCardContentWell, false)
				vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, Theme.PanelBorder, false)
			}
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 10, Y: y, W: innerW - 20, H: rowH}, label, metrics, textCol)
		y += rowH + postBattleRowGap
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y + 4, W: innerW, H: lineH}, "Стрелки — выбор · Пробел / Enter или кнопка ниже", metrics, Theme.TextMuted)
	cl := p.ConfirmRewardLabel
	if cl == "" {
		cl = "Подтвердить"
	}
	drawPostBattlePrimaryButton(screen, hudFace, layout.RewardConfirmButton, cl, p.HoverRewardConfirm, metrics)
}

func drawPostBattlePrimaryButton(screen *ebiten.Image, hudFace *text.GoTextFace, r PostBattleRect, label string, hover bool, metrics battlepkg.HUDMetrics) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	fill := Theme.AbilityBG
	border := Theme.AbilityBorder
	if hover {
		fill = Theme.ButtonHoverBG
		border = Theme.AccentStrip
	}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1.25, border, false)
	if hover {
		vector.FillRect(screen, r.X, r.Y, 3, r.H, Theme.AccentStrip, false)
	}
	rr := rect{X: r.X, Y: r.Y, W: r.W, H: r.H}
	drawSingleLineInRect(screen, hudFace, rr, label, metrics, Theme.TextPrimary)
}
