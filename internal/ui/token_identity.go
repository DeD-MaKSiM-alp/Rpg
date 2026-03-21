// token_identity.go — лёгкая визуальная идентичность токенов (vector, без спрайтов).

package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// tokenShape: 0 circle, 1 square, 2 triangle, 3 diamond
func tokenShapeKind(u *battlepkg.BattleUnit) int {
	if u == nil {
		return 0
	}
	if u.Side == battlepkg.TeamPlayer {
		switch u.Def.Role {
		case battlepkg.RoleFighter:
			return 1
		case battlepkg.RoleArcher:
			return 2
		case battlepkg.RoleHealer:
			return 0
		case battlepkg.RoleMage:
			return 3
		}
	}
	h := int(u.ID)*31 + len(u.Def.ArchetypeID)*13
	for _, c := range u.Def.ArchetypeID {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h % 4
}

func tintFill(base color.RGBA, dr, dg, db int) color.RGBA {
	r := int(base.R) + dr
	g := int(base.G) + dg
	b := int(base.B) + db
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	}
	if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	}
	if b > 255 {
		b = 255
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: base.A}
}

// allySlotTint — различимые оттенки союзников по индексу в партии.
func allySlotTint(base color.RGBA, partyIndex int) color.RGBA {
	if partyIndex < 0 {
		partyIndex = 0
	}
	tints := []struct{ dr, dg, db int }{
		{10, 14, 0},
		{-8, 10, 18},
		{14, -6, 8},
		{0, 12, 20},
		{-6, 8, 14},
		{12, 8, -8},
	}
	t := tints[partyIndex%len(tints)]
	return tintFill(base, t.dr, t.dg, t.db)
}

func enemyVariantTint(base color.RGBA, u *battlepkg.BattleUnit) color.RGBA {
	if u == nil {
		return base
	}
	h := int(u.ID) * 31
	for _, c := range u.Def.ArchetypeID {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	v := h % 5
	tints := []struct{ dr, dg, db int }{
		{12, -4, -4},
		{-6, 8, 14},
		{16, 4, -8},
		{0, -6, 12},
		{10, 10, -6},
	}
	t := tints[v%len(tints)]
	return tintFill(base, t.dr, t.dg, t.db)
}

func tokenFillForUnit(u *battlepkg.BattleUnit, deadFill color.RGBA) color.RGBA {
	if u == nil || !u.IsAlive() {
		return deadFill
	}
	base := Theme.BattlefieldTokenAlly
	if u.Side == battlepkg.TeamEnemy {
		base = Theme.BattlefieldTokenEnemy
		return enemyVariantTint(base, u)
	}
	idx := u.Origin.PartyActiveIndex
	if idx < 0 {
		idx = int(u.ID) % 6
	}
	return allySlotTint(base, idx)
}

// drawTokenBody рисует силуэт токена внутри радиуса от (cx, cy).
func drawTokenBody(screen *ebiten.Image, cx, cy, radius float32, fill color.RGBA, shape int) {
	switch shape {
	case 1:
		s := radius * 1.45
		vector.FillRect(screen, cx-s*0.5, cy-s*0.5, s, s, fill, false)
	case 2:
		drawTokenTriangle(screen, cx, cy, radius*1.35, fill)
	case 3:
		drawTokenDiamond(screen, cx, cy, radius*1.25, fill)
	default:
		vector.FillCircle(screen, cx, cy, radius, fill, false)
	}
}

func drawTokenTriangle(screen *ebiten.Image, cx, cy, r float32, fill color.RGBA) {
	var path vector.Path
	path.MoveTo(cx, cy-r)
	path.LineTo(cx+r*0.9, cy+r*0.75)
	path.LineTo(cx-r*0.9, cy+r*0.75)
	path.Close()
	fo := &vector.FillOptions{FillRule: vector.FillRuleEvenOdd}
	do := &vector.DrawPathOptions{}
	do.ColorScale.ScaleWithColor(fill)
	vector.FillPath(screen, &path, fo, do)
}

func drawTokenDiamond(screen *ebiten.Image, cx, cy, r float32, fill color.RGBA) {
	var path vector.Path
	path.MoveTo(cx, cy-r)
	path.LineTo(cx+r, cy)
	path.LineTo(cx, cy+r)
	path.LineTo(cx-r, cy)
	path.Close()
	fo := &vector.FillOptions{FillRule: vector.FillRuleEvenOdd}
	do := &vector.DrawPathOptions{}
	do.ColorScale.ScaleWithColor(fill)
	vector.FillPath(screen, &path, fo, do)
}

func strokeTokenBody(screen *ebiten.Image, cx, cy, radius float32, clr color.RGBA, shape int, strokeW float32) {
	switch shape {
	case 1:
		s := radius * 1.45
		vector.StrokeRect(screen, cx-s*0.5, cy-s*0.5, s, s, strokeW, clr, false)
	case 2:
		strokeTokenTriangle(screen, cx, cy, radius*1.35, clr, strokeW)
	case 3:
		strokeTokenDiamond(screen, cx, cy, radius*1.25, clr, strokeW)
	default:
		vector.StrokeCircle(screen, cx, cy, radius, strokeW, clr, false)
	}
}

func strokeTokenTriangle(screen *ebiten.Image, cx, cy, r float32, clr color.RGBA, w float32) {
	var path vector.Path
	path.MoveTo(cx, cy-r)
	path.LineTo(cx+r*0.9, cy+r*0.75)
	path.LineTo(cx-r*0.9, cy+r*0.75)
	path.Close()
	so := &vector.StrokeOptions{Width: w, LineJoin: vector.LineJoinMiter}
	do := &vector.DrawPathOptions{}
	do.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(screen, &path, so, do)
}

func strokeTokenDiamond(screen *ebiten.Image, cx, cy, r float32, clr color.RGBA, w float32) {
	var path vector.Path
	path.MoveTo(cx, cy-r)
	path.LineTo(cx+r, cy)
	path.LineTo(cx, cy+r)
	path.LineTo(cx-r, cy)
	path.Close()
	so := &vector.StrokeOptions{Width: w, LineJoin: vector.LineJoinMiter}
	do := &vector.DrawPathOptions{}
	do.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(screen, &path, so, do)
}

func drawTokenIdentityBadge(screen *ebiten.Image, hudFace *text.GoTextFace, cx, cy, radius float32, u *battlepkg.BattleUnit, metrics battlepkg.HUDMetrics) {
	if u == nil || hudFace == nil {
		return
	}
	badgeW := radius * 1.1
	if badgeW < 14 {
		badgeW = 14
	}
	badgeH := metrics.LineH * 0.85
	if badgeH < 12 {
		badgeH = 12
	}
	bx := cx + radius*0.55
	by := cy + radius*0.38
	vector.FillRect(screen, bx-badgeW*0.5, by-badgeH*0.5, badgeW, badgeH, Theme.PanelBGDeep, false)
	vector.StrokeRect(screen, bx-badgeW*0.5, by-badgeH*0.5, badgeW, badgeH, 1, Theme.PanelBorder, false)
	glyph := ""
	if u.Side == battlepkg.TeamPlayer {
		glyph = string([]rune{battlepkg.RoleAbbrev(u.Def.Role)})
	} else {
		glyph = battlepkg.EnemyTokenGlyph(u)
	}
	br := rect{X: bx - badgeW*0.5, Y: by - badgeH*0.5, W: badgeW, H: badgeH}
	drawSingleLineInRect(screen, hudFace, br, glyph, metrics, Theme.TextSecondary)
}

// rosterIdentityStripColor согласует цвет полоски ростера с токеном.
func rosterIdentityStripColor(u *battlepkg.BattleUnit) color.RGBA {
	if u == nil {
		return Theme.PanelBorder
	}
	if !u.IsAlive() {
		return Theme.DeadText
	}
	return tokenFillForUnit(u, Theme.DeadFill)
}
