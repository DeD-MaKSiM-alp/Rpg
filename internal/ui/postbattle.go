package ui

import (
	"image/color"

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
	ScreenW, ScreenH int
	PanelX, PanelY, PanelW, PanelH float32
	InnerX, InnerY, InnerW         float32
	LineH                          float32
	Pad                            float32
	// RewardOptionRects — области клика и подсветки строк награды (совпадают с отрисовкой).
	RewardOptionRects []PostBattleRect
}

// postBattle metrics — одна точка правды для чисел layout (draw + hit-test).
const (
	postBattleLineH   = float32(22)
	postBattlePad     = float32(24)
	postBattlePanelW  = float32(400)
	postBattleRowH    = float32(32)
	postBattleRowGap  = float32(4)
	postBattleBaseH   = float32(220)
	postBattleRewardH = float32(120) // высота панели без строк опций (заголовок + choose + hint)
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
	panelH := postBattleBaseH
	if isRewardStep && optionCount > 0 {
		panelH = postBattleRewardH + float32(optionCount)*36 // 36 = rowH + gap (как в цикле отрисовки)
	}
	panelX := (w - panelW) / 2
	panelY := (h - panelH) / 2

	l.PanelX, l.PanelY, l.PanelW, l.PanelH = panelX, panelY, panelW, panelH
	l.InnerX = panelX + postBattlePad
	l.InnerY = panelY + postBattlePad
	l.InnerW = panelW - postBattlePad*2

	if isRewardStep && optionCount > 0 {
		// Первая строка награды: y = innerY + lineH*3.5; подсветка: y-2, высота rowH+4 (как в DrawPostBattleOverlay).
		firstY := l.InnerY + postBattleLineH*3.5
		l.RewardOptionRects = make([]PostBattleRect, optionCount)
		for i := 0; i < optionCount; i++ {
			y := firstY + float32(i)*(postBattleRowH+postBattleRowGap)
			l.RewardOptionRects[i] = PostBattleRect{
				X: l.InnerX,
				Y: y - 2,
				W: l.InnerW,
				H: postBattleRowH + 4,
			}
		}
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

// PostBattleParams — параметры для отрисовки post-battle overlay (game передаёт готовые строки).
type PostBattleParams struct {
	ResultText    string
	IsRewardStep  bool
	OptionLabels  []string
	OptionDescs   []string
	SelectedIndex int
	ScreenWidth   int
	ScreenHeight  int
}

// DrawPostBattleOverlay рисует полупрозрачный overlay: результат боя и (при победе) выбор награды.
func DrawPostBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p PostBattleParams) {
	if hudFace == nil || p.ScreenWidth < 100 || p.ScreenHeight < 100 {
		return
	}
	layout := ComputePostBattleLayout(p.ScreenWidth, p.ScreenHeight, p.IsRewardStep, len(p.OptionLabels))
	w := float32(p.ScreenWidth)
	h := float32(p.ScreenHeight)
	// Dim background
	vector.FillRect(screen, 0, 0, w, h, color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)

	// Panel background
	vector.FillRect(screen, layout.PanelX, layout.PanelY, layout.PanelW, layout.PanelH, color.RGBA{R: 28, G: 28, B: 34, A: 255}, false)
	vector.StrokeRect(screen, layout.PanelX, layout.PanelY, layout.PanelW, layout.PanelH, 2, color.RGBA{R: 100, G: 100, B: 120, A: 255}, false)

	innerX := layout.InnerX
	innerY := layout.InnerY
	innerW := layout.InnerW
	lineH := layout.LineH
	metrics := battlepkg.HUDMetrics{LineH: lineH}

	// Result line
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY, W: innerW, H: lineH * 1.5}, p.ResultText, metrics, color.White)

	if !p.IsRewardStep {
		drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*2, W: innerW, H: lineH}, "SPACE / ENTER — continue", metrics, color.RGBA{R: 180, G: 180, B: 180, A: 255})
		return
	}

	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*2, W: innerW, H: lineH}, "Choose reward:", metrics, color.RGBA{R: 220, G: 220, B: 220, A: 255})

	y := innerY + lineH*3.5
	for i := 0; i < len(p.OptionLabels); i++ {
		rowH := postBattleRowH
		label := p.OptionLabels[i]
		if i < len(p.OptionDescs) && p.OptionDescs[i] != "" {
			label = label + " — " + p.OptionDescs[i]
		}
		textCol := color.RGBA{R: 200, G: 200, B: 200, A: 255}
		if i == p.SelectedIndex && i < len(layout.RewardOptionRects) {
			textCol = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			fill := color.RGBA{R: 55, G: 65, B: 90, A: 255}
			r := layout.RewardOptionRects[i]
			vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
			vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, 1, color.RGBA{R: 120, G: 140, B: 200, A: 255}, false)
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 8, Y: y, W: innerW - 16, H: rowH}, label, metrics, textCol)
		y += rowH + postBattleRowGap
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y + 4, W: innerW, H: lineH}, "ARROWS — select   SPACE / ENTER — confirm", metrics, color.RGBA{R: 140, G: 140, B: 150, A: 255})
}
