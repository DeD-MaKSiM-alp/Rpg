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
// Теперь мир состоит из чанков и больше не ограничен
// фиксированными размерами по ширине и высоте.
//
// Это значит:
//   - чанки могут существовать при любых координатах;
//   - мир можно бесконечно расширять в любую сторону;
//   - реальные данные мира создаются лениво по мере необходимости.
type World struct {

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

// NewWorld создаёт бесконечный мир.
//
// Теперь мир больше не имеет фиксированной ширины и высоты.
// Мы сохраняем только seed и пустую карту чанков.
// Сами чанки будут создаваться лениво по мере необходимости.
func NewWorld(seed int) World {
	return World{
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
	chunk := newChunk(coord.x, coord.y, w.seed)

	// Сохраняем новый чанк в world,
	// чтобы в следующий раз не создавать его повторно.
	w.chunks[coord] = chunk

	return chunk
}

// IsWalkable проверяет, можно ли пройти в клетку мира.
//
// Теперь мир бесконечный, поэтому у клетки нет проверки
// на попадание "внутрь карты".
// Вместо этого мы всегда:
//  1. определяем нужный чанк;
//  2. лениво создаём его, если он ещё не существует;
//  3. читаем локальный тайл внутри чанка.
func (w *World) IsWalkable(x, y int) bool {

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

	// Определяем чанк, в котором находится указанная клетка мира.
	centerCoord, _, _ := worldToChunkLocal(worldX, worldY)

	// Проходим по квадратной области чанков вокруг центрального чанка
	// и гарантируем, что каждый из них создан.
	for chunkY := centerCoord.y - radius; chunkY <= centerCoord.y+radius; chunkY++ {
		for chunkX := centerCoord.x - radius; chunkX <= centerCoord.x+radius; chunkX++ {
			coord := ChunkCoord{x: chunkX, y: chunkY}

			w.getOrCreateChunk(coord)
		}
	}
}

// UnloadChunksFarFrom удаляет из памяти чанки,
// которые находятся слишком далеко от заданной мировой клетки.
//
// Зачем это нужно:
// сейчас мир создаёт чанки лениво,
// но без выгрузки их количество будет только расти.
// Этот метод оставляет в памяти только область вокруг игрока,
// а дальние чанки удаляет.
//
// radius задаётся в чанках:
//   - radius = 0  -> оставить только текущий чанк;
//   - radius = 1  -> оставить текущий чанк и соседей вокруг;
//   - radius = 2  -> оставить более широкую область.
//
// Важно:
// мы работаем именно с координатами чанков,
// а не с расстоянием в клетках.
func (w *World) UnloadChunksFarFrom(worldX, worldY, radius int) {

	// Определяем чанк, в котором сейчас находится игрок.
	centerCoord, _, _ := worldToChunkLocal(worldX, worldY)

	// Проходим по всем уже загруженным чанкам
	// и удаляем те, которые находятся слишком далеко.
	for coord := range w.chunks {
		dx := coord.x - centerCoord.x
		if dx < 0 {
			dx = -dx
		}

		dy := coord.y - centerCoord.y
		if dy < 0 {
			dy = -dy
		}

		// Если чанк выходит за допустимый радиус по X или Y,
		// удаляем его из памяти.
		if dx > radius || dy > radius {
			delete(w.chunks, coord)
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

// newChunk создаёт один чанк бесконечного мира.
//
// Теперь у чанка больше нет зависимости от общего размера карты.
// Любой чанк можно сгенерировать при любых координатах:
// положительных, нулевых или отрицательных.
//
// Все клетки чанка генерируются через generateTile(...),
// а значит содержимое чанка полностью определяется:
//   - координатами чанка;
//   - координатами клетки внутри мира;
//   - seed мира.
func newChunk(chunkX, chunkY, seed int) *Chunk {
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

			// Для бесконечного мира больше нет понятия
			// "клетка за пределами карты".
			// Поэтому любую клетку просто генерируем по её мировым координатам.
			chunk.tiles[localY][localX] = generateTile(worldX, worldY, seed)
		}
	}

	return chunk
}

// floorDiv выполняет целочисленное деление с округлением вниз.
//
// Это важно для бесконечного мира с отрицательными координатами.
// Обычное деление в Go для отрицательных чисел округляет к нулю,
// а нам нужно именно математическое "вниз".
//
// Пример:
//   - floorDiv(37, 16)  = 2
//   - floorDiv(-1, 16)  = -1
//   - floorDiv(-17, 16) = -2
func floorDiv(a, b int) int {
	result := a / b
	remainder := a % b

	// Если есть остаток и знак результата должен сместиться вниз,
	// уменьшаем частное на единицу.
	if remainder != 0 && ((remainder > 0) != (b > 0)) {
		result--
	}

	return result
}

// positiveMod возвращает неотрицательный остаток от деления.
//
// Для бесконечного мира это нужно,
// чтобы локальные координаты внутри чанка всегда оставались в диапазоне:
//
//	0 <= local < chunkSize
//
// Пример:
//   - positiveMod(5, 16)   = 5
//   - positiveMod(-1, 16)  = 15
//   - positiveMod(-17, 16) = 15
func positiveMod(a, b int) int {
	result := a % b
	if result < 0 {
		result += b
	}
	return result
}

// worldToChunkLocal переводит мировые координаты клетки
// в две части:
//  1. координаты чанка, в котором находится клетка;
//  2. локальные координаты клетки внутри этого чанка.
//
// Эта версия корректно работает и для отрицательных координат мира.
//
// Примеры при chunkSize = 16:
//   - worldX = 37   -> chunkX = 2,  localX = 5
//   - worldX = -1   -> chunkX = -1, localX = 15
//   - worldX = -17  -> chunkX = -2, localX = 15
func worldToChunkLocal(worldX, worldY int) (ChunkCoord, int, int) {
	// Для бесконечного мира с отрицательными координатами
	// обычных / и % недостаточно:
	// они дают корректный результат только для worldX/worldY >= 0.
	chunkX := floorDiv(worldX, chunkSize)
	chunkY := floorDiv(worldY, chunkSize)

	localX := positiveMod(worldX, chunkSize)
	localY := positiveMod(worldY, chunkSize)

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
