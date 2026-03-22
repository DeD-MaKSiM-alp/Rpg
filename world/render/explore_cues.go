// explore_cues.go — UX-подсказки exploration: кольца на соседних клетках с интерактивом (только отрисовка).

package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mygame/internal/visualcolor"
	"mygame/world/entity"
)

// ExploreCueSource — read-only доступ к миру для подсказок (без смены состояния).
type ExploreCueSource interface {
	GetEnemyAt(x, y int) *entity.Entity
	IsWalkable(x, y int) bool
	PickupPreviewAt(x, y int) (entity.PickupKind, bool)
}

// DrawExploreInteractionCues рисует кольца на клетках, куда можно шагнуть с эффектом (бой / подбор / лагерь),
// и мягкое кольцо под ногами, если на текущей клетке есть пикап. Рисовать до отрисовки игрока.
func DrawExploreInteractionCues(screen *ebiten.Image, src ExploreCueSource, px, py, cameraX, cameraY, endX, endY, tileSize int) {
	if screen == nil || src == nil || tileSize < 8 {
		return
	}
	ts := float32(tileSize)
	dirs := [4]struct{ dx, dy int }{
		{0, -1}, {0, 1}, {-1, 0}, {1, 0},
	}
	for _, d := range dirs {
		tx, ty := px+d.dx, py+d.dy
		if tx < cameraX || tx >= endX || ty < cameraY || ty >= endY {
			continue
		}
		if !src.IsWalkable(tx, ty) {
			continue
		}
		cx := float32((tx-cameraX)*tileSize) + ts*0.5
		cy := float32((ty-cameraY)*tileSize) + ts*0.5
		if e := src.GetEnemyAt(tx, ty); e != nil && e.Alive && e.Type == entity.EntityEnemy {
			drawNeighborCueRing(screen, cx, cy, ts, cueEnemy)
			continue
		}
		if k, ok := src.PickupPreviewAt(tx, ty); ok {
			drawNeighborCueRing(screen, cx, cy, ts, cueFromPickupKind(k))
		}
	}
	// Под ногами: пикап на текущей клетке (ещё не собран — для лагеря до подтверждения).
	if k, ok := src.PickupPreviewAt(px, py); ok {
		cx := float32((px-cameraX)*tileSize) + ts*0.5
		cy := float32((py-cameraY)*tileSize) + ts*0.5
		drawFootCueRing(screen, cx, cy, ts, cueFromPickupKind(k))
	}
}

type cueStyle int

const (
	cueEnemy cueStyle = iota
	cueResource
	cueRecruitCamp
	cuePOIMystic
	cuePOIWater
	cuePOIGold
	cuePOIRuins
	cuePOIWarm
)

func cueFromPickupKind(k entity.PickupKind) cueStyle {
	switch k {
	case entity.PickupKindRecruitCamp:
		return cueRecruitCamp
	case entity.PickupKindPOIAltar:
		return cuePOIMystic
	case entity.PickupKindPOISpring:
		return cuePOIWater
	case entity.PickupKindPOICache:
		return cuePOIGold
	case entity.PickupKindPOIRuins:
		return cuePOIRuins
	case entity.PickupKindPOICampfire:
		return cuePOIWarm
	default:
		return cueResource
	}
}

func drawNeighborCueRing(screen *ebiten.Image, cx, cy, ts float32, st cueStyle) {
	r := ts * 0.44
	if r < 10 {
		r = 10
	}
	outer, inner, w := cueColors(st)
	vector.StrokeCircle(screen, cx, cy, r, w, outer, false)
	vector.StrokeCircle(screen, cx, cy, r-2.5, 1, inner, false)
}

func drawFootCueRing(screen *ebiten.Image, cx, cy, ts float32, st cueStyle) {
	r := ts * 0.36
	if r < 8 {
		r = 8
	}
	outer, inner, w := cueColors(st)
	a := outer
	a.A = 140
	vector.StrokeCircle(screen, cx, cy, r, w*0.85, a, false)
	vector.StrokeCircle(screen, cx, cy, r-2, 1, inner, false)
}

func cueColors(st cueStyle) (outer, inner color.RGBA, stroke float32) {
	switch st {
	case cueEnemy:
		return visualcolor.Foundation.SelectedKill, visualcolor.Foundation.EnemyAccent, 2.75
	case cueRecruitCamp:
		return visualcolor.Foundation.HoverTarget, visualcolor.Foundation.AccentStrip, 2.5
	case cuePOIMystic:
		return color.RGBA{R: 200, G: 160, B: 255, A: 255}, color.RGBA{R: 120, G: 90, B: 180, A: 255}, 2.45
	case cuePOIWater:
		return color.RGBA{R: 100, G: 200, B: 255, A: 255}, color.RGBA{R: 40, G: 120, B: 200, A: 255}, 2.4
	case cuePOIGold:
		return color.RGBA{R: 255, G: 210, B: 100, A: 255}, color.RGBA{R: 180, G: 120, B: 40, A: 255}, 2.35
	case cuePOIRuins:
		return color.RGBA{R: 160, G: 150, B: 140, A: 255}, color.RGBA{R: 90, G: 85, B: 78, A: 255}, 2.3
	case cuePOIWarm:
		return color.RGBA{R: 255, G: 140, B: 80, A: 255}, color.RGBA{R: 200, G: 90, B: 40, A: 255}, 2.4
	default:
		return visualcolor.Foundation.AccentStrip, visualcolor.Foundation.PostBattleBorder, 2.35
	}
}
