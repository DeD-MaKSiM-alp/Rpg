package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mygame/world/entity"
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
	floorColor := color.RGBA{R: 30, G: 30, B: 30, A: 255}
	wallColor := color.RGBA{R: 90, G: 90, B: 90, A: 255}
	grassColor := color.RGBA{R: 40, G: 110, B: 40, A: 255}
	waterColor := color.RGBA{R: 40, G: 80, B: 170, A: 255}

	for worldY := cameraY; worldY < endY; worldY++ {
		for worldX := cameraX; worldX < endX; worldX++ {
			tile := source.GetTileAt(worldX, worldY)
			var tileColor color.RGBA
			switch tile {
			case mapdata.TileFloor:
				tileColor = floorColor
			case mapdata.TileWall:
				tileColor = wallColor
			case mapdata.TileGrass:
				tileColor = grassColor
			case mapdata.TileWater:
				tileColor = waterColor
			default:
				tileColor = floorColor
			}
			screenX := float32((worldX-cameraX)*tileSize)
			screenY := float32((worldY-cameraY)*tileSize)
			vector.FillRect(screen, screenX, screenY, float32(tileSize), float32(tileSize), tileColor, false)
		}
	}
}

func drawPickups(screen *ebiten.Image, source DrawSource, cameraX, cameraY, endX, endY, tileSize int) {
	pickupColor := color.RGBA{R: 240, G: 220, B: 60, A: 255}
	pickupSize := float32(tileSize / 2)

	for _, chunk := range source.Chunks() {
		for _, pickup := range chunk.Pickups {
			if pickup.Collected {
				continue
			}
			if pickup.X < cameraX || pickup.X >= endX || pickup.Y < cameraY || pickup.Y >= endY {
				continue
			}
			screenX := float32((pickup.X-cameraX)*tileSize) + float32(tileSize)/4
			screenY := float32((pickup.Y-cameraY)*tileSize) + float32(tileSize)/4
			vector.FillRect(screen, screenX, screenY, pickupSize, pickupSize, pickupColor, false)
		}
	}
}

func drawEnemies(screen *ebiten.Image, source DrawSource, cameraX, cameraY, endX, endY, tileSize int) {
	enemyColor := color.RGBA{R: 200, G: 60, B: 60, A: 255}
	enemySize := float32(tileSize / 2)

	for _, e := range source.Entities() {
		if !e.Alive || e.Type != entity.EntityEnemy {
			continue
		}
		if e.X < cameraX || e.X >= endX || e.Y < cameraY || e.Y >= endY {
			continue
		}
		screenX := float32((e.X-cameraX)*tileSize) + float32(tileSize)/4
		screenY := float32((e.Y-cameraY)*tileSize) + float32(tileSize)/4
		vector.FillRect(screen, screenX, screenY, enemySize, enemySize, enemyColor, false)
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
	chunkLineColor := color.RGBA{R: 220, G: 180, B: 40, A: 255}
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
