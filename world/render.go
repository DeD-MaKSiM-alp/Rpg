package world

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Draw рисует только видимую часть мира.
//
// Для каждой видимой клетки мира мы:
//  1. определяем, в каком чанке она находится;
//  2. при необходимости лениво создаём этот чанк;
//  3. получаем локальный тайл внутри чанка;
//  4. выбираем цвет;
//  5. рисуем тайл относительно камеры.
func (w *World) Draw(screen *ebiten.Image, cameraX, cameraY, visibleTilesX, visibleTilesY, tileSize int) {
	floorColor := color.RGBA{R: 30, G: 30, B: 30, A: 255}
	wallColor := color.RGBA{R: 90, G: 90, B: 90, A: 255}
	grassColor := color.RGBA{R: 40, G: 110, B: 40, A: 255}
	waterColor := color.RGBA{R: 40, G: 80, B: 170, A: 255}

	// Вычисляем границы видимой области мира,
	// которую сейчас показывает камера.
	// В бесконечном мире этой области достаточно:
	// обрезать её размерами карты больше не нужно.
	endX := cameraX + visibleTilesX
	endY := cameraY + visibleTilesY

	// Проходим по всем видимым клеткам мира.
	for worldY := cameraY; worldY < endY; worldY++ {
		for worldX := cameraX; worldX < endX; worldX++ {
			// Определяем, в каком чанке находится текущая клетка,
			// и где она лежит внутри чанка.
			coord, localX, localY := worldToChunkLocal(worldX, worldY)

			// Получаем чанк для текущей клетки.
			// Если этого чанка ещё нет в памяти,
			// он будет создан прямо сейчас.
			chunk := w.getOrCreateChunk(coord)

			tile := chunk.tiles[localY][localX]

			var tileColor color.RGBA

			switch tile {
			case TileFloor:
				tileColor = floorColor
			case TileWall:
				tileColor = wallColor
			case TileGrass:
				tileColor = grassColor
			case TileWater:
				tileColor = waterColor
			default:
				tileColor = floorColor
			}

			// Переводим мировые координаты клетки в экранные,
			// вычитая смещение камеры.
			screenX := float32((worldX - cameraX) * tileSize)
			screenY := float32((worldY - cameraY) * tileSize)

			vector.FillRect(screen, screenX, screenY, float32(tileSize), float32(tileSize), tileColor, false)
		}
	}
}

// DrawChunkDebugOverlay рисует поверх мира отладочную сетку чанков.
//
// На этом этапе overlay показывает:
//   - границы чанков более толстыми линиями;
//   - только ту часть, которая попадает в видимую область камеры.
//
// Это помогает визуально увидеть:
//   - где проходят границы чанков;
//   - когда игрок пересекает чанк;
//   - как работает предзагрузка соседних чанков.
func (w *World) DrawChunkDebugOverlay(
	screen *ebiten.Image,
	cameraX, cameraY int,
	visibleTilesX, visibleTilesY int,
	tileSize int,
	screenWidth, screenHeight int,
) {
	// Цвет линий чанков делаем заметным,
	// чтобы они отличались и от обычной сетки, и от тайлов мира.
	chunkLineColor := color.RGBA{R: 220, G: 180, B: 40, A: 255}

	// Определяем видимую область мира.
	endX := cameraX + visibleTilesX
	endY := cameraY + visibleTilesY

	// Определяем диапазон чанков, попадающих в видимую область.
	// Здесь важно использовать floorDiv(...),
	// чтобы отрицательные координаты камеры тоже работали корректно.
	startChunkX := floorDiv(cameraX, chunkSize)
	startChunkY := floorDiv(cameraY, chunkSize)
	endChunkX := floorDiv(endX-1, chunkSize)
	endChunkY := floorDiv(endY-1, chunkSize)

	// Рисуем вертикальные границы чанков.
	for chunkX := startChunkX; chunkX <= endChunkX+1; chunkX++ {
		worldX := chunkX * chunkSize

		// Граница может оказаться за пределами мира,
		// поэтому ограничиваем её.
		if worldX < cameraX || worldX > endX {
			continue
		}

		screenX := float32((worldX - cameraX) * tileSize)
		vector.StrokeLine(screen, screenX, 0, screenX, float32(screenHeight), 2, chunkLineColor, false)
	}

	// Рисуем горизонтальные границы чанков.
	for chunkY := startChunkY; chunkY <= endChunkY+1; chunkY++ {
		worldY := chunkY * chunkSize

		// Граница может оказаться за пределами мира,
		// поэтому ограничиваем её.
		if worldY < cameraY || worldY > endY {
			continue
		}

		screenY := float32((worldY - cameraY) * tileSize)
		vector.StrokeLine(screen, 0, screenY, float32(screenWidth), screenY, 2, chunkLineColor, false)
	}
}
