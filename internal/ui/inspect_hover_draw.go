package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// DrawBattleInspectHighlights — hover + active-open inspect: три режима (hover / открытая карточка / оба на одном юните).
// Рисуется поверх HUD после DrawBattleOverlay.
func DrawBattleInspectHighlights(screen *ebiten.Image, b *battlepkg.BattleContext, screenW, screenH int, hoverID battlepkg.UnitID, inspectOpen bool, openInspectID battlepkg.UnitID) {
	if b == nil {
		return
	}
	plan := BuildInspectBattleHighlightPlan(hoverID, openInspectID, inspectOpen)
	layout := b.ComputeBattleHUDLayout(screenW, screenH)

	if plan.CombinedUnitID != 0 {
		u := b.Units[plan.CombinedUnitID]
		if u != nil {
			enemy := u.Side == battlepkg.TeamEnemy
			drawRects := func(hr battlepkg.HUDRect) {
				drawBattleInspectCombinedOnRect(screen, hr, enemy)
			}
			if r, ok := layout.UnitRects[plan.CombinedUnitID]; ok {
				drawRects(r)
			}
			if layout.BattlefieldTokens != nil {
				if r, ok := layout.BattlefieldTokens[plan.CombinedUnitID]; ok {
					drawRects(r)
				}
			}
		}
		return
	}

	if plan.ActiveUnitID != 0 {
		u := b.Units[plan.ActiveUnitID]
		if u != nil {
			enemy := u.Side == battlepkg.TeamEnemy
			drawRects := func(hr battlepkg.HUDRect) {
				drawBattleInspectActiveOpenOnRect(screen, hr, enemy)
			}
			if r, ok := layout.UnitRects[plan.ActiveUnitID]; ok {
				drawRects(r)
			}
			if layout.BattlefieldTokens != nil {
				if r, ok := layout.BattlefieldTokens[plan.ActiveUnitID]; ok {
					drawRects(r)
				}
			}
		}
	}

	if plan.HoverUnitID != 0 && plan.HoverStrength > 0 {
		u := b.Units[plan.HoverUnitID]
		if u != nil {
			enemy := u.Side == battlepkg.TeamEnemy
			drawRects := func(hr battlepkg.HUDRect) {
				drawBattleHoverOnRect(screen, hr, enemy, plan.HoverStrength)
			}
			if r, ok := layout.UnitRects[plan.HoverUnitID]; ok {
				drawRects(r)
			}
			if layout.BattlefieldTokens != nil {
				if r, ok := layout.BattlefieldTokens[plan.HoverUnitID]; ok {
					drawRects(r)
				}
			}
		}
	}
}

func drawBattleInspectActiveOpenOnRect(screen *ebiten.Image, hr battlepkg.HUDRect, enemy bool) {
	if hr.W <= 0 || hr.H <= 0 {
		return
	}
	x, y, w, h := hr.X, hr.Y, hr.W, hr.H
	fill := activeInspectFillColor(enemy)
	vector.FillRect(screen, x, y, w, h, fill, false)
	brd := Theme.HoverTarget
	if enemy {
		brd = Theme.EnemyAccent
	}
	vector.StrokeRect(screen, x, y, w, h, 2.35, brd, false)
	vector.StrokeRect(screen, x-1, y-1, w+2, h+2, 1, Theme.AccentStrip, false)
	vector.FillRect(screen, x, y, 4, h, Theme.AccentStrip, false)
}

func drawBattleInspectCombinedOnRect(screen *ebiten.Image, hr battlepkg.HUDRect, enemy bool) {
	if hr.W <= 0 || hr.H <= 0 {
		return
	}
	x, y, w, h := hr.X, hr.Y, hr.W, hr.H
	fill := combinedInspectFillColor(enemy)
	vector.FillRect(screen, x, y, w, h, fill, false)
	brd := Theme.ValidTarget
	if enemy {
		brd = Theme.SelectedKill
	}
	vector.StrokeRect(screen, x, y, w, h, 2.85, brd, false)
	vector.StrokeRect(screen, x-2, y-2, w+4, h+4, 1, Theme.AccentStrip, false)
	vector.FillRect(screen, x, y, 5, h, Theme.AccentStrip, false)
}

func drawBattleHoverOnRect(screen *ebiten.Image, hr battlepkg.HUDRect, enemy bool, strength float32) {
	if hr.W <= 0 || hr.H <= 0 {
		return
	}
	x, y, w, h := hr.X, hr.Y, hr.W, hr.H
	fill := hoverInspectFillColor(enemy, strength)
	vector.FillRect(screen, x, y, w, h, fill, false)
	brd := Theme.HoverTarget
	if enemy {
		brd = Theme.EnemyAccent
	}
	vector.StrokeRect(screen, x, y, w, h, 1.75+strokeBoost(strength), brd, false)
}

func activeInspectFillColor(enemy bool) color.RGBA {
	if enemy {
		return color.RGBA{R: 115, G: 75, B: 85, A: 38}
	}
	return color.RGBA{R: 72, G: 118, B: 165, A: 38}
}

func combinedInspectFillColor(enemy bool) color.RGBA {
	if enemy {
		return color.RGBA{R: 115, G: 75, B: 85, A: 52}
	}
	return color.RGBA{R: 72, G: 118, B: 165, A: 55}
}

func hoverInspectFillColor(enemy bool, strength float32) color.RGBA {
	a := uint8(22 * strength)
	if a < 1 {
		a = 1
	}
	if enemy {
		return color.RGBA{R: 115, G: 75, B: 85, A: a}
	}
	return color.RGBA{R: 75, G: 125, B: 185, A: a}
}

func strokeBoost(strength float32) float32 {
	if strength < 0.75 {
		return 0.25
	}
	return 0.5
}
