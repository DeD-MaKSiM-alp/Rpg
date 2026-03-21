package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

// InspectRoleIcon — минимальный набор пиктограмм роли/типа боя (вектор, без ассетов).
type InspectRoleIcon int

const (
	InspectRoleIconUnknown InspectRoleIcon = iota
	InspectRoleIconMelee                   // ближний / боец
	InspectRoleIconRanged                  // дальний / лучник
	InspectRoleIconHeal                    // целитель / поддержка
	InspectRoleIconArcane                  // маг / магия
)

// InspectRoleIconFromUnitTemplate — маппинг по данным шаблона (канонично для героев).
func InspectRoleIconFromUnitTemplate(tpl *unitdata.UnitTemplate) InspectRoleIcon {
	if tpl == nil {
		return InspectRoleIconUnknown
	}
	switch tpl.AttackKind {
	case unitdata.AttackHeal:
		return InspectRoleIconHeal
	case unitdata.AttackRanged:
		return InspectRoleIconRanged
	case unitdata.AttackMelee:
		if tpl.Role == battlepkg.RoleMage {
			return InspectRoleIconArcane
		}
		return InspectRoleIconMelee
	default:
		return InspectRoleIconMelee
	}
}

// InspectRoleIconFromHero — по UnitID героя.
func InspectRoleIconFromHero(h *hero.Hero) InspectRoleIcon {
	if h == nil {
		return InspectRoleIconUnknown
	}
	tpl, ok := unitdata.GetUnitTemplate(h.UnitID)
	if !ok {
		return InspectRoleIconUnknown
	}
	return InspectRoleIconFromUnitTemplate(&tpl)
}

// InspectRoleIconFromCombatUnit — для врагов / юнитов боя без полного шаблона.
func InspectRoleIconFromCombatUnit(u *battlepkg.CombatUnit) InspectRoleIcon {
	if u == nil {
		return InspectRoleIconUnknown
	}
	if u.Def.TemplateUnitID != "" {
		if tpl, ok := unitdata.GetUnitTemplate(u.Def.TemplateUnitID); ok {
			return InspectRoleIconFromUnitTemplate(&tpl)
		}
	}
	if u.Def.IsRanged {
		return InspectRoleIconRanged
	}
	switch u.Def.Role {
	case battlepkg.RoleHealer:
		return InspectRoleIconHeal
	case battlepkg.RoleMage:
		return InspectRoleIconArcane
	case battlepkg.RoleArcher:
		return InspectRoleIconRanged
	default:
		return InspectRoleIconMelee
	}
}

// DrawInspectRoleIcon рисует иконку в квадрате (x,y) размера size×size.
func DrawInspectRoleIcon(screen *ebiten.Image, x, y, size float32, icon InspectRoleIcon, col color.Color) {
	if size < 8 {
		return
	}
	cx := x + size*0.5
	cy := y + size*0.5
	w := size * 0.08
	if w < 1 {
		w = 1
	}
	switch icon {
	case InspectRoleIconMelee:
		drawIconMelee(screen, x, y, size, col, w)
	case InspectRoleIconRanged:
		drawIconRanged(screen, x, y, size, col, w)
	case InspectRoleIconHeal:
		drawIconHeal(screen, cx, cy, size*0.42, col, w)
	case InspectRoleIconArcane:
		drawIconArcane(screen, cx, cy, size*0.38, col, w)
	default:
		drawIconUnknown(screen, cx, cy, size*0.35, col, w)
	}
}

func drawIconMelee(screen *ebiten.Image, x, y, size float32, col color.Color, w float32) {
	x1 := x + size*0.15
	y1 := y + size*0.82
	x2 := x + size*0.88
	y2 := y + size*0.12
	vector.StrokeLine(screen, x1, y1, x2, y2, w*2.2, col, false)
	vector.StrokeLine(screen, x+size*0.12, y+size*0.78, x+size*0.32, y+size*0.72, w*1.4, col, false)
}

func drawIconRanged(screen *ebiten.Image, x, y, size float32, col color.Color, w float32) {
	// Стрелка вправо — быстро читается как «дальний бой».
	mid := y + size*0.5
	vector.StrokeLine(screen, x+size*0.1, mid, x+size*0.58, mid, w*2, col, false)
	vector.StrokeLine(screen, x+size*0.48, y+size*0.28, x+size*0.88, mid, w*2, col, false)
	vector.StrokeLine(screen, x+size*0.48, y+size*0.72, x+size*0.88, mid, w*2, col, false)
}

func drawIconHeal(screen *ebiten.Image, cx, cy, arm float32, col color.Color, w float32) {
	vector.StrokeLine(screen, cx-arm, cy, cx+arm, cy, w*2, col, false)
	vector.StrokeLine(screen, cx, cy-arm, cx, cy+arm, w*2, col, false)
}

func drawIconArcane(screen *ebiten.Image, cx, cy, r float32, col color.Color, w float32) {
	for i := 0; i < 4; i++ {
		a0 := float64(i) * math.Pi / 2
		a1 := float64(i+1) * math.Pi / 2
		xa := cx + float32(math.Cos(a0))*r
		ya := cy + float32(math.Sin(a0))*r
		xb := cx + float32(math.Cos(a1))*r
		yb := cy + float32(math.Sin(a1))*r
		vector.StrokeLine(screen, xa, ya, xb, yb, w*1.6, col, false)
	}
}

func drawIconUnknown(screen *ebiten.Image, cx, cy, r float32, col color.Color, w float32) {
	steps := 24
	for i := 0; i < steps; i++ {
		a0 := float64(i) / float64(steps) * math.Pi * 2
		a1 := float64(i+1) / float64(steps) * math.Pi * 2
		xa := cx + float32(math.Cos(a0))*r
		ya := cy + float32(math.Sin(a0))*r
		xb := cx + float32(math.Cos(a1))*r
		yb := cy + float32(math.Sin(a1))*r
		vector.StrokeLine(screen, xa, ya, xb, yb, w*1.2, col, false)
	}
}
