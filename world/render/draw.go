package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mygame/internal/visualcolor"
	"mygame/world/entity"
	"mygame/world/generation"
	"mygame/world/mapdata"
)

// DrawSource — источник данных для отрисовки мира (реализуется world.World).
type DrawSource interface {
	GetTileAt(worldX, worldY int) mapdata.TileType
	Chunks() map[mapdata.ChunkCoord]*mapdata.Chunk
	Entities() map[entity.EntityID]*entity.Entity
}

// Draw рисует видимую часть мира: тайлы, пикапы, врагов.
func Draw(
	screen *ebiten.Image,
	source DrawSource,
	cameraX, cameraY, visibleTilesX, visibleTilesY, tileSize int,
) {
	endX := cameraX + visibleTilesX
	endY := cameraY + visibleTilesY
	drawTiles(screen, source, cameraX, cameraY, endX, endY, tileSize)
	drawPickups(screen, source, cameraX, cameraY, endX, endY, tileSize)
	drawEnemies(screen, source, cameraX, cameraY, endX, endY, tileSize)
}

func drawTiles(screen *ebiten.Image, source DrawSource, cameraX, cameraY, endX, endY, tileSize int) {
	ts := float32(tileSize)
	for worldY := cameraY; worldY < endY; worldY++ {
		for worldX := cameraX; worldX < endX; worldX++ {
			tile := source.GetTileAt(worldX, worldY)
			tileColor := tileBaseColor(tile, worldX, worldY)
			screenX := float32((worldX - cameraX) * tileSize)
			screenY := float32((worldY - cameraY) * tileSize)
			vector.FillRect(screen, screenX, screenY, ts, ts, tileColor, false)

			if tile == mapdata.TileWall {
				// Лёгкий «объём» стены: верхняя кромка светлее.
				hl := wallHighlight()
				vector.StrokeLine(screen, screenX+1, screenY+1, screenX+ts-1, screenY+1, 1, hl, false)
				vector.StrokeRect(screen, screenX, screenY, ts, ts, 1, visualcolor.Foundation.PanelTitleSep, false)
			} else if generation.IsTileWalkable(tile) {
				// Тонкая кромка клетки пола — чуть лучше читается сетка без тяжёлой сетки.
				vector.StrokeRect(screen, screenX+0.5, screenY+0.5, ts-1, ts-1, 0.75, floorGridLine(), false)
			}
		}
	}
}

func tileBaseColor(tile mapdata.TileType, worldX, worldY int) color.RGBA {
	parity := (worldX+worldY)&1 == 0
	switch tile {
	case mapdata.TileFloor:
		return blendParity(visualcolor.Foundation.PanelBGDeep, visualcolor.Foundation.SceneTint, parity)
	case mapdata.TileWall:
		return visualcolor.Foundation.PanelBorder
	case mapdata.TileGrass:
		return blendParity(tintRGBA(visualcolor.Foundation.BattlefieldTokenAlly, -25, 15, -20), tintRGBA(visualcolor.Foundation.BattlefieldTokenAlly, -10, 28, -8), parity)
	case mapdata.TileWater:
		return blendParity(tintRGBA(visualcolor.Foundation.ValidTarget, -40, -20, 10), tintRGBA(visualcolor.Foundation.HoverTarget, -55, -35, -5), parity)
	default:
		return visualcolor.Foundation.PanelBGDeep
	}
}

func blendParity(a, b color.RGBA, useA bool) color.RGBA {
	if useA {
		return a
	}
	return b
}

func tintRGBA(c color.RGBA, dr, dg, db int) color.RGBA {
	r := int(c.R) + dr
	g := int(c.G) + dg
	bl := int(c.B) + db
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
	if bl < 0 {
		bl = 0
	}
	if bl > 255 {
		bl = 255
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(bl), A: c.A}
}

func wallHighlight() color.RGBA {
	return color.RGBA{R: 130, G: 136, B: 155, A: 200}
}

func floorGridLine() color.RGBA {
	return color.RGBA{R: 32, G: 36, B: 46, A: 90}
}

func drawPickups(screen *ebiten.Image, source DrawSource, cameraX, cameraY, endX, endY, tileSize int) {
	ts := float32(tileSize)
	for _, chunk := range source.Chunks() {
		for _, pickup := range chunk.Pickups {
			if pickup.Collected {
				continue
			}
			if pickup.X < cameraX || pickup.X >= endX || pickup.Y < cameraY || pickup.Y >= endY {
				continue
			}
			cx := float32((pickup.X-cameraX)*tileSize) + ts*0.5
			cy := float32((pickup.Y-cameraY)*tileSize) + ts*0.5

			switch {
			case pickup.Kind == entity.PickupKindRecruitCamp:
				drawRecruitCampMarker(screen, cx, cy, ts)
			case entity.IsPOIKind(pickup.Kind):
				drawPOIMarker(screen, cx, cy, ts, pickup.Kind)
			default:
				drawResourcePickupMarker(screen, cx, cy, ts)
			}
		}
	}
}

func drawResourcePickupMarker(screen *ebiten.Image, cx, cy, ts float32) {
	r := ts * 0.26
	if r < 6 {
		r = 6
	}
	vector.FillCircle(screen, cx, cy, r, visualcolor.Foundation.AccentStrip, false)
	vector.StrokeCircle(screen, cx, cy, r, 2, visualcolor.Foundation.PostBattleBorder, false)
	vector.FillCircle(screen, cx, cy, r*0.35, visualcolor.Foundation.PanelBGDeep, false)
}

func drawRecruitCampMarker(screen *ebiten.Image, cx, cy, ts float32) {
	r := ts * 0.32
	if r < 8 {
		r = 8
	}
	vector.FillCircle(screen, cx, cy, r, visualcolor.Foundation.AbilityHoverBG, false)
	vector.StrokeCircle(screen, cx, cy, r, 2.25, visualcolor.Foundation.HoverTarget, false)
	vector.StrokeCircle(screen, cx, cy, r-3, 1, visualcolor.Foundation.AccentStrip, false)
	// Упрощённый «шатёр»: треугольник + основание (лагерь наёмников).
	s := ts * 0.22
	vector.StrokeLine(screen, cx-s, cy+s*0.35, cx, cy-s*0.95, 2, visualcolor.Foundation.AccentStrip, false)
	vector.StrokeLine(screen, cx+s, cy+s*0.35, cx, cy-s*0.95, 2, visualcolor.Foundation.AccentStrip, false)
	vector.StrokeLine(screen, cx-s*1.1, cy+s*0.4, cx+s*1.1, cy+s*0.4, 1.5, visualcolor.Foundation.TextPrimary, false)
}

// drawPOIMarker — маркер точки интереса; отличается от ресурса и лагеря.
func drawPOIMarker(screen *ebiten.Image, cx, cy, ts float32, k entity.PickupKind) {
	r := ts * 0.28
	if r < 7 {
		r = 7
	}
	switch k {
	case entity.PickupKindPOIAltar:
		vector.FillCircle(screen, cx, cy, r, color.RGBA{R: 90, G: 70, B: 120, A: 255}, false)
		vector.StrokeCircle(screen, cx, cy, r, 2, color.RGBA{R: 200, G: 160, B: 255, A: 255}, false)
		vector.StrokeLine(screen, cx, cy-r*0.9, cx, cy+r*0.5, 2, color.RGBA{R: 240, G: 220, B: 255, A: 255}, false)
	case entity.PickupKindPOISpring:
		vector.FillCircle(screen, cx, cy-r*0.15, r*0.85, color.RGBA{R: 70, G: 150, B: 220, A: 255}, false)
		vector.FillCircle(screen, cx+r*0.35, cy+r*0.25, r*0.45, color.RGBA{R: 120, G: 200, B: 255, A: 255}, false)
		vector.StrokeCircle(screen, cx, cy-r*0.15, r*0.85, 1.5, color.RGBA{R: 200, G: 240, B: 255, A: 255}, false)
	case entity.PickupKindPOICache:
		sq := r * 1.1
		vector.FillRect(screen, cx-sq*0.5, cy-sq*0.45, sq, sq*0.9, color.RGBA{R: 140, G: 100, B: 40, A: 255}, false)
		vector.StrokeRect(screen, cx-sq*0.5, cy-sq*0.45, sq, sq*0.9, 1.5, color.RGBA{R: 255, G: 210, B: 100, A: 255}, false)
	case entity.PickupKindPOIRuins:
		vector.StrokeRect(screen, cx-r, cy-r*0.7, r*2, r*1.4, 2, color.RGBA{R: 140, G: 130, B: 120, A: 255}, false)
		vector.StrokeLine(screen, cx-r*0.6, cy+r*0.5, cx+r*0.5, cy-r*0.4, 1.5, color.RGBA{R: 100, G: 95, B: 88, A: 220}, false)
	case entity.PickupKindPOICampfire:
		vector.FillCircle(screen, cx, cy+r*0.35, r*0.35, color.RGBA{R: 60, G: 45, B: 30, A: 255}, false)
		vector.StrokeLine(screen, cx-r*0.5, cy+r*0.2, cx, cy-r*0.85, 2.5, color.RGBA{R: 255, G: 140, B: 60, A: 255}, false)
		vector.StrokeLine(screen, cx+r*0.45, cy+r*0.15, cx-r*0.2, cy-r*0.6, 2, color.RGBA{R: 255, G: 200, B: 80, A: 255}, false)
	default:
		drawResourcePickupMarker(screen, cx, cy, ts)
	}
}

func drawEnemies(screen *ebiten.Image, source DrawSource, cameraX, cameraY, endX, endY, tileSize int) {
	ts := float32(tileSize)
	for _, e := range source.Entities() {
		if !e.Alive || e.Type != entity.EntityEnemy {
			continue
		}
		if e.X < cameraX || e.X >= endX || e.Y < cameraY || e.Y >= endY {
			continue
		}
		cx := float32((e.X-cameraX)*tileSize) + ts*0.5
		cy := float32((e.Y-cameraY)*tileSize) + ts*0.5
		r := ts * 0.3
		if r < 8 {
			r = 8
		}
		vector.FillCircle(screen, cx, cy, r, visualcolor.Foundation.HPEnemyFill, false)
		vector.StrokeCircle(screen, cx, cy, r, 2.25, visualcolor.Foundation.EnemyAccent, false)
		vector.StrokeCircle(screen, cx, cy, r-2.5, 1, visualcolor.Foundation.SelectedKill, false)
		// Крест-глиф «враждебность» (черновой маркер).
		d := r * 0.45
		vector.StrokeLine(screen, cx-d, cy-d, cx+d, cy+d, 1.5, visualcolor.Foundation.PanelBGDeep, false)
		vector.StrokeLine(screen, cx-d, cy+d, cx+d, cy-d, 1.5, visualcolor.Foundation.PanelBGDeep, false)
	}
}

// DrawChunkDebugOverlay рисует отладочную сетку границ чанков.
func DrawChunkDebugOverlay(
	screen *ebiten.Image,
	cameraX, cameraY int,
	visibleTilesX, visibleTilesY int,
	tileSize int,
	screenWidth, screenHeight int,
) {
	chunkLineColor := visualcolor.Foundation.AccentStrip
	endX := cameraX + visibleTilesX
	endY := cameraY + visibleTilesY

	startChunkX := mapdata.FloorDiv(cameraX, mapdata.ChunkSize)
	startChunkY := mapdata.FloorDiv(cameraY, mapdata.ChunkSize)
	endChunkX := mapdata.FloorDiv(endX-1, mapdata.ChunkSize)
	endChunkY := mapdata.FloorDiv(endY-1, mapdata.ChunkSize)

	for chunkX := startChunkX; chunkX <= endChunkX+1; chunkX++ {
		worldX := chunkX * mapdata.ChunkSize
		if worldX < cameraX || worldX > endX {
			continue
		}
		screenX := float32((worldX - cameraX) * tileSize)
		vector.StrokeLine(screen, screenX, 0, screenX, float32(screenHeight), 2, chunkLineColor, false)
	}
	for chunkY := startChunkY; chunkY <= endChunkY+1; chunkY++ {
		worldY := chunkY * mapdata.ChunkSize
		if worldY < cameraY || worldY > endY {
			continue
		}
		screenY := float32((worldY - cameraY) * tileSize)
		vector.StrokeLine(screen, 0, screenY, float32(screenWidth), screenY, 2, chunkLineColor, false)
	}
}
