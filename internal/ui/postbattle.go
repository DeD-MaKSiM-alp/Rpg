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
	postBattleLineH    = float32(22)
	postBattlePad      = float32(24)
	postBattlePanelW   = float32(400)
	postBattleRowH     = float32(32)
	postBattleRowGap   = float32(4)
	postBattleButtonH  = float32(36)
	postBattleButtonW  = float32(200)
)

// ComputePostBattleLayout вычисляет layout post-battle экрана для заданного размера окна.
// isRewardStep и optionCount задают высоту панели и число прямоугольников наград.
func ComputePostBattleLayout(screenW, screenH int, isRewardStep bool, optionCount int) PostBattleLayout {
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
		// Высота внутренней области: заголовок + строки награды + подсказка + кнопка «Подтвердить».
		innerH := lineH*3.5 + float32(n)*(postBattleRowH+postBattleRowGap) + 4 + lineH + 8 + postBattleButtonH + 8
		panelH = postBattlePad*2 + innerH
	} else {
		// Результат боя: заголовок + подсказка + кнопка «Продолжить» / «В мир».
		panelH = postBattlePad*2 + lineH*3 + 12 + postBattleButtonH + 16
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
		l.ResultContinueButton = PostBattleRect{
			X: btnX,
			Y: l.InnerY + lineH*3 + 12,
			W: btnW,
			H: postBattleButtonH,
		}
		return l
	}

	n := optionCount
	if n < 0 {
		n = 0
	}
	firstY := l.InnerY + lineH*3.5
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
		hintY = l.InnerY + lineH*3.5 + 4
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
	// Кнопки (явный mouse path; Space/Enter остаётся альтернативой).
	ContinueButtonLabel string
	ConfirmRewardLabel  string
	HoverContinue       bool
	HoverRewardConfirm  bool
}

// DrawPostBattleOverlay рисует полупрозрачный overlay: результат боя и (при победе) выбор награды.
func DrawPostBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p PostBattleParams) {
	if hudFace == nil || p.ScreenWidth < 100 || p.ScreenHeight < 100 {
		return
	}
	layout := ComputePostBattleLayout(p.ScreenWidth, p.ScreenHeight, p.IsRewardStep, len(p.OptionLabels))
	w := float32(p.ScreenWidth)
	h := float32(p.ScreenHeight)
	vector.FillRect(screen, 0, 0, w, h, Theme.OverlayDim, false)

	vector.FillRect(screen, layout.PanelX, layout.PanelY, layout.PanelW, layout.PanelH, Theme.PostBattlePanelBG, false)
	vector.StrokeRect(screen, layout.PanelX, layout.PanelY, layout.PanelW, layout.PanelH, 2, Theme.PostBattleBorder, false)

	innerX := layout.InnerX
	innerY := layout.InnerY
	innerW := layout.InnerW
	lineH := layout.LineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}

	// Result line
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY, W: innerW, H: lineH * 1.5}, p.ResultText, metrics, Theme.TextPrimary)

	if !p.IsRewardStep {
		drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*2, W: innerW, H: lineH}, "Пробел / Enter — продолжить или кнопка ниже", metrics, Theme.TextMuted)
		lbl := p.ContinueButtonLabel
		if lbl == "" {
			lbl = "Продолжить"
		}
		rc := layout.ResultContinueButton
		drawPostBattlePrimaryButton(screen, hudFace, rc, lbl, p.HoverContinue, metrics)
		return
	}

	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*2, W: innerW, H: lineH}, "Выберите награду:", metrics, Theme.TextSecondary)

	y := innerY + lineH*3.5
	for i := 0; i < len(p.OptionLabels); i++ {
		rowH := postBattleRowH
		label := p.OptionLabels[i]
		if i < len(p.OptionDescs) && p.OptionDescs[i] != "" {
			label = label + " — " + p.OptionDescs[i]
		}
		textCol := Theme.TextSecondary
		if i == p.SelectedIndex && i < len(layout.RewardOptionRects) {
			textCol = Theme.TextPrimary
			r := layout.RewardOptionRects[i]
			vector.FillRect(screen, r.X, r.Y, r.W, r.H, Theme.PostBattleRowSelect, false)
			vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, Theme.PostBattleRowBrd, false)
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 8, Y: y, W: innerW - 16, H: rowH}, label, metrics, textCol)
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
	fill := Theme.ButtonBG
	border := Theme.ButtonBorder
	if hover {
		fill = Theme.ButtonHoverBG
		border = Theme.ButtonHoverBorder
	}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, border, false)
	rr := rect{X: r.X, Y: r.Y, W: r.W, H: r.H}
	drawSingleLineInRect(screen, hudFace, rr, label, metrics, Theme.TextPrimary)
}
