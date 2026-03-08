package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	// chunkSize — размер одного чанка в клетках.
	// Например, при значении 16 один чанк содержит 16x16 тайлов.
	// Дальше весь мир будет состоять из таких одинаковых блоков.
	chunkSize = 16
)

// World хранит мир целиком на уровне логики.
// Теперь мир состоит не из одной большой карты,
// а из набора чанков, которые лежат в chunks.
//
// width и height пока оставляем как общий размер мира в клетках.
// Это удобно на переходном этапе, пока мир ещё не бесконечный.
type World struct {
	// width — полная ширина мира в клетках.
	width int

	// height — полная высота мира в клетках.
	height int

	// seed — числовое зерно мира.
	// Оно влияет на процедурную генерацию:
	// один и тот же seed даёт один и тот же мир,
	// а другой seed — уже другой вариант мира.
	seed int

	// chunks — все чанки мира, доступные по их координатам.
	// Ключом является ChunkCoord, значением — указатель на Chunk.
	chunks map[ChunkCoord]*Chunk
}

// ChunkCoord — координаты чанка в мире.
// Это не координаты клетки, а именно номер чанка.
// Например, chunk (0, 0), chunk (1, 0), chunk (-1, 2).
type ChunkCoord struct {
	x int
	y int
}

// Chunk — отдельный кусок мира фиксированного размера.
// Он хранит:
//   - свои координаты в сетке чанков;
//   - локальные тайлы внутри себя.
//
// Важно:
// chunkX/chunkY — это не координаты клетки игрока,
// а позиция чанка среди других чанков мира.
type Chunk struct {
	chunkX int
	chunkY int
	tiles  [][]TileType
}

// NewWorld создаёт мир заданного размера в клетках.
//
// seed определяет конкретный вариант процедурно сгенерированного мира.
// При одинаковых width, height и seed мир будет одинаковым.
// При другом seed мир изменится, даже если размеры останутся теми же.
func NewWorld(width, height, seed int) World {
	return World{
		width:  width,
		height: height,
		seed:   seed,
		chunks: make(map[ChunkCoord]*Chunk),
	}
}

// getOrCreateChunk возвращает чанк по его координатам.
//
// Если чанк уже существует в карте w.chunks,
// функция просто возвращает его.
//
// Если чанка ещё нет, функция создаёт его через newChunk(...),
// сохраняет в карту чанков и только потом возвращает.
//
// Это и есть основа ленивого создания мира:
// мы не создаём все чанки заранее,
// а подгружаем их только по мере необходимости.
func (w *World) getOrCreateChunk(coord ChunkCoord) *Chunk {
	// Сначала пытаемся найти уже существующий чанк.
	if chunk, exists := w.chunks[coord]; exists {
		return chunk
	}

	// Если чанка ещё нет — создаём его.
	chunk := newChunk(coord.x, coord.y, w.width, w.height, w.seed)

	// Сохраняем новый чанк в world,
	// чтобы в следующий раз не создавать его повторно.
	w.chunks[coord] = chunk

	return chunk
}

// IsWalkable проверяет, можно ли пройти в клетку мира.
//
// Теперь проверка выполняется через:
//  1. проверку границ мира;
//  2. определение нужного чанка;
//  3. ленивое создание чанка, если он ещё не существует;
//  4. чтение локального тайла внутри чанка.
func (w *World) IsWalkable(x, y int) bool {
	// Если клетка вне мира — пройти туда нельзя.
	if !w.IsInside(x, y) {
		return false
	}

	// Определяем, в каком чанке находится клетка,
	// и где именно она лежит внутри чанка.
	coord, localX, localY := worldToChunkLocal(x, y)

	// Получаем чанк по координатам.
	// Если он ещё не был создан раньше,
	// world создаст его автоматически.
	chunk := w.getOrCreateChunk(coord)

	return chunk.tiles[localY][localX] == TileFloor
}

// Draw рисует только видимую часть мира.
//
// Для каждой видимой клетки мира мы:
//  1. определяем, в каком чанке она находится;
//  2. при необходимости лениво создаём этот чанк;
//  3. получаем локальный тайл внутри чанка;
//  4. выбираем цвет;
//  5. рисуем тайл относительно камеры.
func (w *World) Draw(screen *ebiten.Image, cameraX, cameraY int) {
	floorColor := color.RGBA{R: 30, G: 30, B: 30, A: 255}
	wallColor := color.RGBA{R: 90, G: 90, B: 90, A: 255}

	// Вычисляем границы видимой области мира,
	// которую сейчас показывает камера.
	endX := cameraX + visibleTilesX
	endY := cameraY + visibleTilesY

	if endX > w.width {
		endX = w.width
	}
	if endY > w.height {
		endY = w.height
	}

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

			tileColor := floorColor
			if chunk.tiles[localY][localX] == TileWall {
				tileColor = wallColor
			}

			// Переводим мировые координаты клетки в экранные,
			// вычитая смещение камеры.
			screenX := float32((worldX - cameraX) * tileSize)
			screenY := float32((worldY - cameraY) * tileSize)

			vector.FillRect(screen, screenX, screenY, float32(tileSize), float32(tileSize), tileColor, false)
		}
	}
}

// PreloadChunksAround заранее создаёт чанки вокруг заданной клетки мира.
//
// Это нужно не для обязательной логики движения,
// а для более аккуратной подготовки окружающей области:
// например, чтобы соседние чанки уже существовали к моменту,
// когда игрок подойдёт к их границе.
//
// radius задаётся в чанках:
//   - radius = 0  -> только текущий чанк;
//   - radius = 1  -> текущий чанк и все соседи вокруг;
//   - radius = 2  -> ещё более широкая область.
func (w *World) PreloadChunksAround(worldX, worldY, radius int) {
	// Если точка вообще вне мира, ничего не делаем.
	if !w.IsInside(worldX, worldY) {
		return
	}

	// Определяем чанк, в котором находится указанная клетка мира.
	centerCoord, _, _ := worldToChunkLocal(worldX, worldY)

	// Проходим по квадратной области чанков вокруг центрального чанка
	// и гарантируем, что каждый из них создан.
	for chunkY := centerCoord.y - radius; chunkY <= centerCoord.y+radius; chunkY++ {
		for chunkX := centerCoord.x - radius; chunkX <= centerCoord.x+radius; chunkX++ {
			coord := ChunkCoord{x: chunkX, y: chunkY}

			// Для конечного мира с координатами от 0 и выше
			// отрицательные чанки нам не нужны.
			if coord.x < 0 || coord.y < 0 {
				continue
			}

			w.getOrCreateChunk(coord)
		}
	}
}

// ChunkCoordAt возвращает координаты чанка,
// в котором находится мировая клетка worldX/worldY.
//
// Это удобно для debug-отображения и любых систем,
// которым нужно быстро понять, в каком чанке сейчас находится игрок
// или любой другой объект мира.
func (w *World) ChunkCoordAt(worldX, worldY int) ChunkCoord {
	coord, _, _ := worldToChunkLocal(worldX, worldY)
	return coord
}

// LoadedChunkCount возвращает текущее количество чанков,
// которые уже были созданы и находятся в памяти.
//
// Это полезно для отладки:
// можно сразу видеть, как работает ленивое создание чанков
// и не растёт ли их число слишком быстро.
func (w *World) LoadedChunkCount() int {
	return len(w.chunks)
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
func (w *World) DrawChunkDebugOverlay(screen *ebiten.Image, cameraX, cameraY int) {
	// Цвет линий чанков делаем заметным,
	// чтобы они отличались и от обычной сетки, и от тайлов мира.
	chunkLineColor := color.RGBA{R: 220, G: 180, B: 40, A: 255}

	// Определяем видимую область мира.
	endX := cameraX + visibleTilesX
	endY := cameraY + visibleTilesY

	if endX > w.width {
		endX = w.width
	}
	if endY > w.height {
		endY = w.height
	}

	// Находим диапазон чанков, которые видны на экране.
	startChunkX := cameraX / chunkSize
	startChunkY := cameraY / chunkSize
	endChunkX := (endX - 1) / chunkSize
	endChunkY := (endY - 1) / chunkSize

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

// Width возвращает ширину мира в клетках.
func (w *World) Width() int {
	return w.width
}

// Height возвращает высоту мира в клетках.
func (w *World) Height() int {
	return w.height
}

// newChunk создаёт один чанк мира.
//
// Теперь чанк больше не копирует старую прямоугольную карту.
// Вместо этого каждая клетка внутри мира генерируется
// через отдельную функцию generateTile(...).
//
// Что важно на этом этапе:
//   - генерация детерминированная;
//   - один и тот же участок мира всегда выглядит одинаково;
//   - клетки за пределами конечного мира по-прежнему считаются стенами.
//
// Позже сюда можно будет добавить seed,
// более сложные правила генерации, биомы и структуры.
func newChunk(chunkX, chunkY, worldWidth, worldHeight, seed int) *Chunk {
	chunk := &Chunk{
		chunkX: chunkX,
		chunkY: chunkY,
		tiles:  make([][]TileType, chunkSize),
	}

	// Создаём строки массива тайлов чанка.
	for y := 0; y < chunkSize; y++ {
		chunk.tiles[y] = make([]TileType, chunkSize)
	}

	// Заполняем тайлы чанка.
	for localY := 0; localY < chunkSize; localY++ {
		for localX := 0; localX < chunkSize; localX++ {
			// Переводим локальные координаты внутри чанка
			// в мировые координаты клетки.
			worldX := chunkX*chunkSize + localX
			worldY := chunkY*chunkSize + localY

			// Если клетка уже вышла за фактический размер мира,
			// считаем её стеной.
			// Это нужно для крайних чанков,
			// которые могут частично выходить за размеры конечного мира.
			if worldX >= worldWidth || worldY >= worldHeight {
				chunk.tiles[localY][localX] = TileWall
				continue
			}

			// Для всех клеток внутри мира используем отдельную функцию генерации.
			// Так логика содержимого чанка становится независимой
			// от самой структуры чанка и её будет проще развивать дальше.
			chunk.tiles[localY][localX] = generateTile(worldX, worldY, seed)
		}
	}

	return chunk
}

// worldToChunkLocal переводит мировые координаты клетки
// в две части:
//  1. координаты чанка, в котором находится клетка;
//  2. локальные координаты клетки внутри этого чанка.
//
// Пример:
// при chunkSize = 16 и worldX = 37:
//   - chunkX = 2
//   - localX = 5
func worldToChunkLocal(worldX, worldY int) (ChunkCoord, int, int) {
	// Находим номер чанка,
	// в который попадает клетка мира.
	chunkX := worldX / chunkSize
	chunkY := worldY / chunkSize

	// Находим локальную позицию клетки внутри чанка.
	localX := worldX % chunkSize
	localY := worldY % chunkSize

	return ChunkCoord{x: chunkX, y: chunkY}, localX, localY
}

// generateTile определяет, каким должен быть тайл
// в мировой клетке worldX/worldY.
//
// В отличие от предыдущей версии, теперь генерация зависит не только
// от координат клетки, но и от seed мира.
// Это значит:
//   - один и тот же seed всегда даёт одинаковую карту;
//   - другой seed даёт уже другой вариант мира.
//
// Пока логика остаётся простой:
//   - стартовую область игрока сохраняем свободной;
//   - основная масса клеток остаётся проходимой;
//   - часть клеток превращается в стены по детерминированной формуле.
func generateTile(worldX, worldY, seed int) TileType {
	// Стартовую область вокруг игрока оставляем свободной,
	// чтобы игрок не появился внутри стены
	// и мог спокойно начать движение.
	if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
		return TileFloor
	}

	// Добавляем seed в формулу так,
	// чтобы один и тот же мир был воспроизводим,
	// но при смене seed распределение препятствий менялось.
	value := (worldX*37 + worldY*57 + worldX*worldY*13 + seed*71 + (worldX+seed)*(worldY+11)) % 100

	// Небольшой процент клеток делаем стенами.
	// Это значение можно регулировать:
	// меньше — мир свободнее,
	// больше — мир плотнее и сложнее для перемещения.
	if value < 18 {
		return TileWall
	}

	return TileFloor
}

// IsInside проверяет, лежит ли клетка внутри границ мира.
// Здесь мы работаем именно с мировыми координатами,
// а не с координатами экрана и не с координатами внутри чанка.
func (w *World) IsInside(x, y int) bool {
	return x >= 0 && x < w.width && y >= 0 && y < w.height
}
