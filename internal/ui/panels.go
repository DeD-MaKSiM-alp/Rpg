// Package ui: shared panel/rect/text toolkit. Battle HUD rendering lives in battle_panels.go.
package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

const (
	uiLineH = float32(18)
	uiPad   = float32(12)
	uiGap   = float32(10)
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampF(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func inset(r rect, pad float32) rect {
	n := rect{X: r.X + pad, Y: r.Y + pad, W: r.W - pad*2, H: r.H - pad*2}
	if n.W < 0 {
		n.W = 0
	}
	if n.H < 0 {
		n.H = 0
	}
	return n
}

func splitH(r rect, leftW, gap float32) (rect, rect) {
	left := rect{X: r.X, Y: r.Y, W: leftW, H: r.H}
	right := rect{X: r.X + leftW + gap, Y: r.Y, W: r.W - leftW - gap, H: r.H}
	if right.W < 0 {
		right.W = 0
	}
	return left, right
}

func splitV(r rect, topH, gap float32) (rect, rect) {
	top := rect{X: r.X, Y: r.Y, W: r.W, H: topH}
	bot := rect{X: r.X, Y: r.Y + topH + gap, W: r.W, H: r.H - topH - gap}
	if bot.H < 0 {
		bot.H = 0
	}
	return top, bot
}

// drawHUDText рисует текстовые блоки HUD (счётчик собранных предметов и т.п.).
func drawHUDText(screen *ebiten.Image, pickupCount int, hudFace *text.GoTextFace) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, 20)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Pickups: %d", pickupCount), hudFace, op)
}

type rect struct {
	X, Y, W, H float32
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// measureTextWidth returns the rendered width of a string for the given face.
func measureTextWidth(face *text.GoTextFace, s string) float32 {
	if s == "" || face == nil {
		return 0
	}
	adv := text.Advance(s, face)
	return float32(adv)
}

// trimTextToWidth returns a single-line string that fits into maxW pixels,
// appending "..." when trimming is required.
func trimTextToWidth(face *text.GoTextFace, s string, maxW float32) string {
	if maxW <= 0 || s == "" || face == nil {
		return ""
	}
	if measureTextWidth(face, s) <= maxW {
		return s
	}

	const ellipsis = "..."
	ellW := measureTextWidth(face, ellipsis)
	if ellW >= maxW {
		return ""
	}

	rs := []rune(s)
	lo, hi := 0, len(rs)
	best := ""
	for lo <= hi {
		mid := (lo + hi) / 2
		cand := string(rs[:mid])
		if measureTextWidth(face, cand)+ellW <= maxW {
			best = cand
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	if best == "" {
		return ellipsis
	}
	return best + ellipsis
}

// fitTextToWidth is a convenience alias for single-line trimming.
func fitTextToWidth(face *text.GoTextFace, s string, maxW float32) string {
	return trimTextToWidth(face, s, maxW)
}

// baselineYForLineInRect returns the Y position (top of line) for a single line of text
// so that a line of height metrics.LineH is vertically centered in the rect.
// Used by drawSingleLineInRect; no manual baseline math elsewhere.
func baselineYForLineInRect(r rect, metrics battlepkg.HUDMetrics) float32 {
	if metrics.LineH <= 0 || r.H <= 0 {
		return r.Y
	}
	off := (r.H - metrics.LineH) * 0.5
	if off < 0 {
		off = 0
	}
	return r.Y + off
}

// drawSingleLineInRect draws one line of text inside r, using the shared baseline helper.
// Text is trimmed to fit r.W and vertically placed within r. Does not draw if r has no area.
func drawSingleLineInRect(screen *ebiten.Image, face *text.GoTextFace, r rect, line string, metrics battlepkg.HUDMetrics, col color.Color) {
	if r.W <= 0 || r.H <= 0 || face == nil {
		return
	}
	line = fitTextToWidth(face, line, r.W)
	if line == "" {
		return
	}
	y := baselineYForLineInRect(r, metrics)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(r.X), float64(y))
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, line, face, op)
}

// isCompactForRect decides whether a given rect should use compact wording
// based on its width and the current HUD metrics.
func isCompactForRect(metrics battlepkg.HUDMetrics, r rect) bool {
	// Treat narrow panels or small line heights as candidates for compact text.
	if r.W <= 0 {
		return false
	}
	if r.W < 260 {
		return true
	}
	if metrics.LineH <= 16 {
		return true
	}
	return false
}

// maxLinesForRect returns how many lines (at the given lineStep) can fit
// vertically into the rect, respecting top and bottom padding.
func maxLinesForRect(metrics battlepkg.HUDMetrics, r rect, topPad, bottomPad, lineStep float32) int {
	usableH := r.H - topPad - bottomPad
	if usableH <= 0 || lineStep <= 0 {
		return 0
	}
	n := int(usableH / lineStep)
	if n < 0 {
		return 0
	}
	return n
}

// drawLinesInRect draws a list of lines inside r on a strict grid (lineStep = metrics.LineH).
// Single-line zones use drawSingleLineInRect; multi-line content uses this. maxLines 0 = use capacity.
func drawLinesInRect(screen *ebiten.Image, face *text.GoTextFace, r rect, lines []string, metrics battlepkg.HUDMetrics, col color.Color, maxLines int) int {
	if len(lines) == 0 || r.W <= 0 || r.H <= 0 || face == nil {
		return 0
	}
	lineStep := metrics.LineH
	capacity := int(r.H / lineStep)
	if capacity <= 0 {
		return 0
	}
	if maxLines > 0 && maxLines < capacity {
		capacity = maxLines
	}
	linesToDraw := capacity
	if linesToDraw > len(lines) {
		linesToDraw = len(lines)
	}

	drawn := 0
	for i := 0; i < linesToDraw; i++ {
		y := r.Y + float32(i)*lineStep
		if y+lineStep > r.Y+r.H {
			break
		}
		line := lines[i]
		row := rect{X: r.X, Y: y, W: r.W, H: lineStep}
		drawSingleLineInRect(screen, face, row, line, metrics, col)
		drawn++
	}
	return drawn
}

// drawPanelBox draws panel background, border, optional title-row separator, and title text.
// Title is drawn in a slightly brighter tone so it reads as header; a thin line under title separates it from content.
func drawPanelBox(screen *ebiten.Image, panelRect rect, titleRow rect, title string, hudFace *text.GoTextFace, metrics battlepkg.HUDMetrics) {
	bg := color.RGBA{R: 28, G: 28, B: 28, A: 255}
	border := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	titleSepColor := color.RGBA{R: 55, G: 55, B: 55, A: 255}
	vector.FillRect(screen, panelRect.X, panelRect.Y, panelRect.W, panelRect.H, bg, false)
	vector.StrokeRect(screen, panelRect.X, panelRect.Y, panelRect.W, panelRect.H, 1, border, false)

	if title != "" && titleRow.W > 0 && titleRow.H > 0 {
		drawSingleLineInRect(screen, hudFace, titleRow, title, metrics, color.RGBA{R: 230, G: 230, B: 230, A: 255})
		sepY := titleRow.Y + titleRow.H
		if sepY < panelRect.Y+panelRect.H-1 {
			vector.StrokeLine(screen, panelRect.X, sepY, panelRect.X+panelRect.W, sepY, 1, titleSepColor, false)
		}
	}
}

