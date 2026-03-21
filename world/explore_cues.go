package world

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"mygame/world/entity"
	"mygame/world/mapdata"
	"mygame/world/render"
)

// PickupPreviewAt — read-only: есть ли несобранный пикап на клетке и какого вида (для UX, без сбора).
func (w *World) PickupPreviewAt(worldX, worldY int) (entity.PickupKind, bool) {
	coord, _, _ := mapdata.WorldToChunkLocal(worldX, worldY)
	chunk := w.getOrCreateChunk(coord)
	for i := range chunk.Pickups {
		p := &chunk.Pickups[i]
		if p.Collected {
			continue
		}
		if p.X != worldX || p.Y != worldY {
			continue
		}
		return p.Kind, true
	}
	return 0, false
}

// ExploreHUDHintLine — одна строка для нижней панели explore: что рядом по сторонам света.
func (w *World) ExploreHUDHintLine(px, py int) string {
	var parts []string
	dirs := []struct {
		dx, dy int
		name   string
	}{
		{0, -1, "↑"},
		{0, 1, "↓"},
		{-1, 0, "←"},
		{1, 0, "→"},
	}
	for _, d := range dirs {
		tx, ty := px+d.dx, py+d.dy
		if !w.IsWalkable(tx, ty) {
			continue
		}
		if w.GetEnemyAt(tx, ty) != nil {
			parts = append(parts, d.name+" бой")
			continue
		}
		if k, ok := w.PickupPreviewAt(tx, ty); ok {
			if k == entity.PickupKindRecruitCamp {
				parts = append(parts, d.name+" лагерь")
			} else {
				parts = append(parts, d.name+" добыча")
			}
		}
	}
	if k, ok := w.PickupPreviewAt(px, py); ok {
		if k == entity.PickupKindRecruitCamp {
			parts = append(parts, "здесь лагерь")
		} else {
			parts = append(parts, "здесь добыча")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "Шаг: " + strings.Join(parts, " · ")
}

// DrawExploreCues — кольца интерактива (см. render.DrawExploreInteractionCues).
func (w *World) DrawExploreCues(screen *ebiten.Image, px, py, cameraX, cameraY, visX, visY, tileSize int) {
	render.DrawExploreInteractionCues(screen, w, px, py, cameraX, cameraY, cameraX+visX, cameraY+visY, tileSize)
}
