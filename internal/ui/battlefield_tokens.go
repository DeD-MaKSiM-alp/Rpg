// battlefield_tokens.go — визуальные токены юнитов на поле боя (v2), без спрайтов: vector + Theme.

package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// DrawBattlefieldV2Scene — сетка слотов, разделитель сторон, токены юнитов (после фона battlefield, до ростеров).
func DrawBattlefieldV2Scene(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, layout battlepkg.BattleHUDLayout) {
	if battle == nil || layout.Style != battlepkg.LayoutStyleV2Disciples {
		return
	}
	bf := layout.Battlefield
	if bf.W <= 0 || bf.H <= 0 {
		return
	}

	metrics := layout.Metrics
	inner := inset(battleToRect(bf), metrics.Pad*1.2)
	midGap := metrics.Gap * 2.5
	halfW := (inner.W - midGap) * 0.5
	midX := inner.X + halfW + midGap*0.5

	drawRowBand := func(hr battlepkg.HUDRect, front bool) {
		if hr.W <= 0 {
			return
		}
		r := battleToRect(hr)
		fill := Theme.BattlefieldBackRowBand
		if front {
			fill = Theme.BattlefieldFrontRowBand
		}
		vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
	}
	// Подложки рядов: задний дальше от центра, передний ближе к линии столкновения.
	drawRowBand(layout.BattlefieldPlayerBack, false)
	drawRowBand(layout.BattlefieldPlayerFront, true)
	drawRowBand(layout.BattlefieldEnemyFront, true)
	drawRowBand(layout.BattlefieldEnemyBack, false)

	// Лёгкий акцент на внутренней кромке переднего ряда (к центру боя).
	pf := battleToRect(layout.BattlefieldPlayerFront)
	if pf.W > 0 {
		vector.StrokeLine(screen, pf.X+pf.W, inner.Y+3, pf.X+pf.W, inner.Y+inner.H-3, 1, Theme.BattlefieldFrontRowBorder, false)
	}
	ef := battleToRect(layout.BattlefieldEnemyFront)
	if ef.W > 0 {
		vector.StrokeLine(screen, ef.X, inner.Y+3, ef.X, inner.Y+inner.H-3, 1, Theme.BattlefieldFrontRowBorder, false)
	}

	// Разделитель «линия столкновения» между сторонами
	vector.StrokeLine(screen, midX, inner.Y+4, midX, inner.Y+inner.H-4, 1.25, Theme.PanelTitleSep, false)

	// Ячейки сетки — слабая обводка, чтобы не перебивать токены и группировку рядов
	for _, cell := range layout.BattlefieldSlotCells {
		cr := battleToRect(cell)
		vector.StrokeRect(screen, cr.X+1, cr.Y+1, cr.W-2, cr.H-2, 1, Theme.BattlefieldEmptyCellBorder, false)
	}

	for _, lb := range layout.BattlefieldRowLabels {
		if lb.Rect.W <= 0 {
			continue
		}
		lr := battleToRect(lb.Rect)
		drawSingleLineInRect(screen, hudFace, lr, lb.Text, metrics, Theme.TextMuted)
	}

	// Токены поверх ячеек
	for id, hr := range layout.BattlefieldTokens {
		u := battle.Units[id]
		if u == nil {
			continue
		}
		drawBattlefieldUnitToken(screen, hudFace, battle, u, battleToRect(hr), metrics)
	}

	DrawBattleFeedbackFloats(screen, hudFace, battle, layout, metrics)
}

func drawBattlefieldUnitToken(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, u *battlepkg.BattleUnit, r rect, metrics battlepkg.HUDMetrics) {
	cx := r.X + r.W*0.5
	radius := r.W * 0.36
	if h := r.H * 0.38; h < radius {
		radius = h
	}
	if radius < 6 {
		radius = 6
	}
	cy := r.Y + r.H*0.42

	fill := tokenFillForUnit(u, Theme.DeadFill)
	shape := tokenShapeKind(u)

	border := Theme.AllyAccent
	if u.Side == battlepkg.TeamEnemy {
		border = Theme.EnemyAccent
	}
	active := battle.ActiveUnit()
	if active != nil && active.ID == u.ID {
		border = Theme.ActiveTurn
	} else if u.IsAlive() && u.Side == battlepkg.TeamPlayer && battle.Phase == battlepkg.PhaseAwaitAction &&
		active != nil && active.Side == battlepkg.TeamPlayer && active.ID != u.ID {
		border = Theme.WaitAlly
	}
	pt := &battle.PlayerTurn
	if pt.HoverTargetUnitID == u.ID {
		border = Theme.HoverTarget
	}
	if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && pt.SelectedTarget.UnitID == u.ID {
		border = Theme.SelectedKill
	}

	strokeW := float32(2.5)
	if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && pt.SelectedTarget.UnitID == u.ID {
		strokeW = 3.8
	} else if pt.HoverTargetUnitID == u.ID {
		strokeW = 3.2
	}

	drawTokenBody(screen, cx, cy, radius, fill, shape)
	if k, in := battle.FeedbackFlashIntensity(u.ID); k >= 0 && in > 0 {
		clr := Theme.FeedbackDamageOverlay
		switch k {
		case battlepkg.FeedbackKindHeal:
			clr = Theme.FeedbackHealOverlay
		case battlepkg.FeedbackKindDeath:
			clr = Theme.FeedbackDeathOverlay
		}
		a := float32(clr.A) * in
		ov := color.RGBA{R: clr.R, G: clr.G, B: clr.B, A: uint8(a)}
		drawTokenBody(screen, cx, cy, radius, ov, shape)
	}
	strokeTokenBody(screen, cx, cy, radius, border, shape, strokeW)

	if !u.IsAlive() {
		d := radius * 0.55
		vector.StrokeLine(screen, cx-d, cy-d, cx+d, cy+d, 2, Theme.DeadText, false)
		vector.StrokeLine(screen, cx-d, cy+d, cx+d, cy-d, 2, Theme.DeadText, false)
	}

	// Индекс в строю (союзники) / условный номер врага
	label := ""
	if u.Side == battlepkg.TeamPlayer && u.Origin.PartyActiveIndex >= 0 {
		label = fmt.Sprintf("%d", u.Origin.PartyActiveIndex+1)
	} else {
		label = fmt.Sprintf("%d", int(u.ID)%10)
	}
	if hudFace != nil && label != "" && u.IsAlive() {
		lr := rect{X: cx - 14, Y: cy - radius*0.85, W: 28, H: metrics.LineH * 0.9}
		drawSingleLineInRect(screen, hudFace, lr, label, metrics, Theme.TextPrimary)
	}
	if u.IsAlive() {
		drawTokenIdentityBadge(screen, hudFace, cx, cy, radius, u, metrics)
	}

	// Микро-HP под токеном
	if u.IsAlive() {
		barY := r.Y + r.H - 5
		DrawHPBarMicro(screen, r.X+4, barY, r.W-8, 4, u.State.HP, u.MaxHP(), true, u.Side == battlepkg.TeamEnemy)
	}

	// Маркер хода — пульсирующее кольцо (круг, читаемо для любых силуэтов)
	if active != nil && active.ID == u.ID && u.IsAlive() {
		pulse := float32(1 + math.Sin(float64(battle.Feedback.FrameTick)*0.11)*0.08)
		vector.StrokeCircle(screen, cx, cy, (radius+5)*pulse, 1.5, Theme.ActiveTurn, false)
	}

}
