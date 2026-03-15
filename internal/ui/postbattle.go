package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// PostBattleParams — параметры для отрисовки post-battle overlay (game передаёт готовые строки).
type PostBattleParams struct {
	ResultText      string
	IsRewardStep    bool
	OptionLabels    []string
	OptionDescs     []string
	SelectedIndex   int
	ScreenWidth     int
	ScreenHeight    int
}

// DrawPostBattleOverlay рисует полупрозрачный overlay: результат боя и (при победе) выбор награды.
func DrawPostBattleOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p PostBattleParams) {
	if hudFace == nil || p.ScreenWidth < 100 || p.ScreenHeight < 100 {
		return
	}
	w := float32(p.ScreenWidth)
	h := float32(p.ScreenHeight)
	// Dim background
	vector.FillRect(screen, 0, 0, w, h, color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)

	lineH := float32(22)
	pad := float32(24)
	panelW := float32(400)
	if panelW > w-pad*2 {
		panelW = w - pad*2
	}
	panelH := float32(220)
	if p.IsRewardStep && len(p.OptionLabels) > 0 {
		panelH = float32(120 + len(p.OptionLabels)*36)
	}
	panelX := (w - panelW) / 2
	panelY := (h - panelH) / 2

	// Panel background
	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 28, G: 28, B: 34, A: 255}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2, color.RGBA{R: 100, G: 100, B: 120, A: 255}, false)

	innerX := panelX + pad
	innerY := panelY + pad
	innerW := panelW - pad*2
	metrics := battlepkg.HUDMetrics{LineH: lineH}

	// Result line
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY, W: innerW, H: lineH * 1.5}, p.ResultText, metrics, color.White)

	if !p.IsRewardStep {
		// "Press Space to continue"
		drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*2, W: innerW, H: lineH}, "SPACE / ENTER — continue", metrics, color.RGBA{R: 180, G: 180, B: 180, A: 255})
		return
	}

	// Reward step: "Choose reward:"
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: innerY + lineH*2, W: innerW, H: lineH}, "Choose reward:", metrics, color.RGBA{R: 220, G: 220, B: 220, A: 255})
	y := innerY + lineH*3.5
	for i := 0; i < len(p.OptionLabels); i++ {
		rowH := float32(32)
		label := p.OptionLabels[i]
		if i < len(p.OptionDescs) && p.OptionDescs[i] != "" {
			label = label + " — " + p.OptionDescs[i]
		}
		fill := color.RGBA{R: 40, G: 40, B: 48, A: 255}
		textCol := color.RGBA{R: 200, G: 200, B: 200, A: 255}
		if i == p.SelectedIndex {
			fill = color.RGBA{R: 55, G: 65, B: 90, A: 255}
			textCol = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			vector.FillRect(screen, innerX, y-2, innerW, rowH+4, fill, false)
			vector.StrokeRect(screen, innerX, y-2, innerW, rowH+4, 1, color.RGBA{R: 120, G: 140, B: 200, A: 255}, false)
		}
		drawSingleLineInRect(screen, hudFace, rect{X: innerX + 8, Y: y, W: innerW - 16, H: rowH}, label, metrics, textCol)
		y += rowH + 4
	}
	drawSingleLineInRect(screen, hudFace, rect{X: innerX, Y: y + 4, W: innerW, H: lineH}, "ARROWS — select   SPACE / ENTER — confirm", metrics, color.RGBA{R: 140, G: 140, B: 150, A: 255})
}
