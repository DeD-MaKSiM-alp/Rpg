// battle_roster_card.go — черновой «карточный» вид слотов ростера v2 и мини-портрет/заглушка для токена.
// Только отрисовка; логика боя и layout не меняются.

package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/unitdata"
)

// drawBattleRosterUnitCard рисует один слот ростера v2 как компактную карточку (рамка, мини-портрет/роль, HP).
func drawBattleRosterUnitCard(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext, u *battlepkg.BattleUnit, hr battlepkg.HUDRect, metrics battlepkg.HUDMetrics, inspectOpenID battlepkg.UnitID, inspectOpen bool) {
	if u == nil || screen == nil {
		return
	}
	r := battleToRect(hr)
	if r.W <= 2 || r.H <= 2 {
		return
	}

	fill := Theme.PanelBGDeep
	border := Theme.AllyAccent
	textCol := Theme.TextPrimary
	if u.Side == battlepkg.TeamEnemy {
		border = Theme.EnemyAccent
	}
	if !u.IsAlive() {
		fill = Theme.DeadFill
		textCol = Theme.DeadText
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

	strokeW := float32(2)
	if pt.SelectedTarget.Kind == battlepkg.TargetKindUnit && pt.SelectedTarget.UnitID == u.ID {
		strokeW = 2.85
	} else if pt.HoverTargetUnitID == u.ID {
		strokeW = 2.45
	}

	// Подложка карточки + лёгкая «стеклянная» внутренняя зона (как у inspect: тёмный well).
	vector.FillRect(screen, r.X, r.Y, r.W, r.H, fill, false)
	innerPad := float32(3)
	if r.W > innerPad*2+6 && r.H > innerPad*2+4 {
		vector.FillRect(screen, r.X+innerPad, r.Y+innerPad, r.W-innerPad*2, r.H-innerPad*2, Theme.RosterCardContentWell, false)
	}

	portraitW := r.H - innerPad*2 - 2
	if portraitW > 34 {
		portraitW = 34
	}
	maxPortrait := r.W * 0.38
	if portraitW > maxPortrait {
		portraitW = maxPortrait
	}
	if portraitW < 16 {
		portraitW = 16
	}
	px := r.X + innerPad + 1
	py := r.Y + innerPad + 1
	drawUnitMiniPortraitInRect(screen, hudFace, px, py, portraitW, r.H-innerPad*2-2, u)

	vector.FillRect(screen, r.X, r.Y, 4, r.H, rosterIdentityStripColor(u), false)

	// Тонкая внутренняя обводка текстового блока (связь с рамкой inspect-карточки).
	textLeft := px + portraitW + 5
	textW := r.X + r.W - textLeft - innerPad
	if textW > 8 && r.H > innerPad*2+4 {
		vector.StrokeRect(screen, textLeft-2, r.Y+innerPad, textW+2, r.H-innerPad*2, 1, Theme.RosterCardInnerStroke, false)
	}

	vector.StrokeRect(screen, r.X, r.Y, r.W, r.H, strokeW, border, false)

	if k, in := battle.FeedbackFlashIntensity(u.ID); k >= 0 && in > 0 {
		drawFeedbackOverlayRect(screen, r, k, in)
	}

	name := u.Name()
	if len([]rune(name)) > 10 {
		rs := []rune(name)
		name = string(rs[:10])
	}
	if active != nil && u.Side == battlepkg.TeamPlayer && u.IsAlive() {
		if u.ID == active.ID {
			name = "▶ " + name
		} else {
			name = "· " + name
		}
	}

	lineH := metrics.LineH
	row1 := rect{X: textLeft, Y: r.Y + innerPad + 1, W: textW, H: lineH}
	drawSingleLineInRect(screen, hudFace, row1, fitTextToWidth(hudFace, name, textW), metrics, textCol)

	hpStr := "Погиб"
	if u.IsAlive() {
		hpStr = fmt.Sprintf("ОЗ %d/%d", u.State.HP, u.MaxHP())
	}
	row2 := rect{X: textLeft, Y: r.Y + innerPad + 1 + lineH*0.92, W: textW, H: lineH}
	drawSingleLineInRect(screen, hudFace, row2, hpStr, metrics, textCol)

	// Бейдж ряда (ближний / дальний) — только визуальная подсказка из IsRanged.
	if u.IsAlive() && textW > 28 {
		badge := "Б"
		if u.IsRanged() {
			badge = "Д"
		}
		bw := minF(22, textW*0.35)
		bh := lineH * 0.78
		if bh < 11 {
			bh = 11
		}
		bx := r.X + r.W - innerPad - bw
		by := r.Y + innerPad + 2
		vector.FillRect(screen, bx, by, bw, bh, Theme.PanelBGDeep, false)
		vector.StrokeRect(screen, bx, by, bw, bh, 1, Theme.PanelBorder, false)
		br := rect{X: bx, Y: by, W: bw, H: bh}
		drawSingleLineInRect(screen, hudFace, br, badge, metrics, Theme.TextMuted)
	}

	badgeReserve := float32(0)
	if inspectOpen && inspectOpenID == u.ID && u.IsAlive() {
		badgeW := float32(22)
		badgeH := lineH * 0.82
		if badgeH < 12 {
			badgeH = 12
		}
		badgeReserve = badgeH + 3
		ix := r.X + r.W - innerPad - badgeW
		iy := r.Y + r.H - innerPad - badgeH
		vector.FillRect(screen, ix, iy, badgeW, badgeH, Theme.PanelBGDeep, false)
		vector.StrokeRect(screen, ix, iy, badgeW, badgeH, 1, Theme.AccentStrip, false)
		ir := rect{X: ix, Y: iy, W: badgeW, H: badgeH}
		drawSingleLineInRect(screen, hudFace, ir, "i", metrics, Theme.AccentStrip)
	}

	if u.IsAlive() && r.H > lineH*2+innerPad*2+6 {
		barY := r.Y + r.H - innerPad - 5 - badgeReserve
		barW := r.W - innerPad*2
		if textLeft > r.X+innerPad {
			barW = r.X + r.W - textLeft - innerPad
			DrawHPBarMicro(screen, textLeft, barY, barW, 4, u.State.HP, u.MaxHP(), true, u.Side == battlepkg.TeamEnemy)
		} else {
			DrawHPBarMicro(screen, r.X+innerPad, barY, barW, 4, u.State.HP, u.MaxHP(), true, u.Side == battlepkg.TeamEnemy)
		}
	}
}

// drawUnitMiniPortraitInRect — мини-область «портрет»: squire JPEG или заглушка с глифом роли/врага.
func drawUnitMiniPortraitInRect(screen *ebiten.Image, hudFace *text.GoTextFace, x, y, w, h float32, u *battlepkg.BattleUnit) {
	if screen == nil || u == nil || w <= 2 || h <= 2 {
		return
	}
	vector.FillRect(screen, x, y, w, h, Theme.RosterCardPortraitBG, false)
	vector.StrokeRect(screen, x, y, w, h, 1, Theme.PostBattleBorder, false)

	if u.Def.TemplateUnitID == unitdata.EmpireWarriorSquire {
		if img := SquirePortraitImage(); img != nil {
			drawImageContain(screen, img, x+1, y+1, w-2, h-2)
			if !u.IsAlive() {
				vector.FillRect(screen, x, y, w, h, color.RGBA{R: 0, G: 0, B: 0, A: 130}, false)
			}
			return
		}
	}
	glyph := ""
	if u.Side == battlepkg.TeamPlayer {
		glyph = string([]rune{battlepkg.RoleAbbrev(u.Def.Role)})
	} else {
		glyph = battlepkg.EnemyTokenGlyph(u)
	}
	if glyph != "" && hudFace != nil {
		pr := rect{X: x, Y: y, W: w, H: h}
		lh := h * 0.72
		if lh > 18 {
			lh = 18
		}
		if lh < 10 {
			lh = 10
		}
		drawSingleLineInRect(screen, hudFace, pr, glyph, battlepkg.HUDMetrics{LineH: lh}, Theme.TextSecondary)
	}

	if !u.IsAlive() {
		vector.FillRect(screen, x, y, w, h, color.RGBA{R: 0, G: 0, B: 0, A: 130}, false)
	}
}

// drawBattlefieldMiniPortraitCorner — крошечный блок в углу ячейки поля (связь с ростером / inspect).
func drawBattlefieldMiniPortraitCorner(screen *ebiten.Image, hudFace *text.GoTextFace, cell rect, u *battlepkg.BattleUnit, metrics battlepkg.HUDMetrics) {
	if u == nil || screen == nil {
		return
	}
	sz := cell.W * 0.22
	if sz > 22 {
		sz = 22
	}
	if sz < 14 {
		sz = 14
	}
	if cell.W < sz+4 || cell.H < sz+4 {
		return
	}
	drawUnitMiniPortraitInRect(screen, hudFace, cell.X+3, cell.Y+3, sz, sz, u)
}

func drawImageContain(dst *ebiten.Image, src *ebiten.Image, x, y, w, h float32) {
	if dst == nil || src == nil || w <= 0 || h <= 0 {
		return
	}
	b := src.Bounds()
	bw := float64(b.Dx())
	bh := float64(b.Dy())
	if bw <= 0 || bh <= 0 {
		return
	}
	sx := float64(w) / bw
	sy := float64(h) / bh
	s := sx
	if sy < sx {
		s = sy
	}
	nw := bw * s
	nh := bh * s
	ox := float64(x) + (float64(w)-nw)*0.5
	oy := float64(y) + (float64(h)-nh)*0.5
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(ox, oy)
	dst.DrawImage(src, op)
}

// drawBattlefieldInspectPin — маленькая отметка «открыта карточка» у токена (дублирует смысл overlay, но читается в сцене).
func drawBattlefieldInspectPin(screen *ebiten.Image, hudFace *text.GoTextFace, cell rect, u *battlepkg.BattleUnit, inspectOpenID battlepkg.UnitID, inspectOpen bool, metrics battlepkg.HUDMetrics) {
	if !inspectOpen || inspectOpenID != u.ID || u == nil || !u.IsAlive() {
		return
	}
	sz := float32(12)
	cx := cell.X + cell.W - sz - 4
	cy := cell.Y + 4
	vector.FillRect(screen, cx, cy, sz, sz, Theme.PanelBGDeep, false)
	vector.StrokeRect(screen, cx, cy, sz, sz, 1, Theme.AccentStrip, false)
	r := rect{X: cx, Y: cy, W: sz, H: sz}
	drawSingleLineInRect(screen, hudFace, r, "i", metrics, Theme.AccentStrip)
}
