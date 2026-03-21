package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// drawFeedbackOverlayRect — полупрозрачная вспышка на прямоугольнике (v1 слоты, v2 карточки).
func drawFeedbackOverlayRect(screen *ebiten.Image, r rect, kind int, intensity float32) {
	if kind < 0 || intensity <= 0 || r.W <= 0 || r.H <= 0 {
		return
	}
	var base color.RGBA
	switch kind {
	case battlepkg.FeedbackKindDamage:
		base = Theme.FeedbackDamageOverlay
	case battlepkg.FeedbackKindHeal:
		base = Theme.FeedbackHealOverlay
	case battlepkg.FeedbackKindDeath:
		base = Theme.FeedbackDeathOverlay
	default:
		return
	}
	a := float32(base.A) * intensity
	if a > 255 {
		a = 255
	}
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, color.RGBA{R: base.R, G: base.G, B: base.B, A: uint8(a)}, false)
}

// DrawBattleFeedbackFloats — всплывающие числа урона/лечения (v2: привязка к токену или карточке).
func DrawBattleFeedbackFloats(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, layout battlepkg.BattleHUDLayout, metrics battlepkg.HUDMetrics) {
	if battle == nil || hudFace == nil || len(battle.Feedback.Floats) == 0 {
		return
	}
	stack := map[battlepkg.UnitID]int{}
	for _, f := range battle.Feedback.Floats {
		cx, cy, ok := feedbackFloatAnchor(layout, f.UnitID)
		if !ok {
			continue
		}
		idx := stack[f.UnitID]
		stack[f.UnitID] = idx + 1
		age := 1 - float32(f.TicksLeft)/float32(f.TotalTicks)
		if f.TotalTicks <= 0 {
			age = 0
		}
		y := cy + (-18 - age*38) - float32(idx)*14

		s := fmt.Sprintf("-%d", f.Value)
		col := Theme.TextDanger
		if f.Heal {
			s = fmt.Sprintf("+%d", f.Value)
			col = Theme.TextSuccess
		}
		alpha := float32(f.TicksLeft) / float32(f.TotalTicks)
		if f.TotalTicks <= 0 {
			alpha = 1
		}
		fr := rect{X: cx - 48, Y: y, W: 96, H: metrics.LineH * 1.1}
		pad := float32(3)
		vector.FillRect(screen, fr.X-pad, fr.Y-pad*0.5, fr.W+pad*2, fr.H+pad, color.RGBA{R: 8, G: 10, B: 14, A: 175}, false)
		drawSingleLineInRect(screen, hudFace, fr, s, metrics, feedbackTextAlpha(col, alpha))
	}
}

func feedbackTextAlpha(base color.RGBA, alpha float32) color.RGBA {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	a := float32(base.A) * alpha
	if a > 255 {
		a = 255
	}
	return color.RGBA{R: base.R, G: base.G, B: base.B, A: uint8(a)}
}

func feedbackFloatAnchor(layout battlepkg.BattleHUDLayout, id battlepkg.UnitID) (cx, cy float32, ok bool) {
	if r, ok := layout.BattlefieldTokens[id]; ok && r.W > 0 {
		return r.X + r.W*0.5, r.Y + r.H*0.28, true
	}
	if r, ok := layout.UnitRects[id]; ok && r.W > 0 {
		return r.X + r.W*0.5, r.Y + r.H*0.35, true
	}
	return 0, 0, false
}
