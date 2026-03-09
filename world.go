package main

import (
	"image/color"
	"math"

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

func (w *World) IsWalkable(x, y int) bool {
	coord, localX, localY := worldToChunkLocal(x, y)
	chunk := w.getOrCreateChunk(coord)

	tile := chunk.tiles[localY][localX]
	return isTileWalkable(tile)
}

// isTileWalkable определяет, можно ли пройти по конкретному типу тайла.
//
// Это важный шаг для расширения мира:
// теперь игровая логика не зависит от проверки "tile == TileFloor".
// Мы можем добавлять новые типы клеток и отдельно решать,
// какие из них проходимы, а какие нет.
func isTileWalkable(tile TileType) bool {
	switch tile {
	case TileFloor, TileGrass:
		return true
	case TileWall, TileWater:
		return false
	default:
		return false
	}
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

// fade сглаживает t в диапазоне 0..1.
// Используется для плавной интерполяции между значениями шума.
func fade(t float64) float64 {
	return t * t * (3 - 2*t)
}

// lerp выполняет линейную интерполяцию между a и b.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// hash2D детерминированно превращает целочисленные координаты и seed
// в псевдослучайное число.
func hash2D(x, y, seed int) int {
	h := x*374761393 + y*668265263 + seed*69069
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16

	if h < 0 {
		h = -h
	}

	return h
}

// randomValue2D возвращает детерминированное значение в диапазоне 0..1
// для узла сетки (x, y).
func randomValue2D(x, y, seed int) float64 {
	return float64(hash2D(x, y, seed)%10000) / 10000.0
}

// valueNoise2D возвращает сглаженное noise-значение в диапазоне 0..1
// для вещественных координат x/y.
//
// Идея такая:
//  1. берём 4 соседних узла сетки;
//  2. у каждого есть фиксированное псевдослучайное значение;
//  3. плавно интерполируем между ними.
//
// Это даёт связные области вместо "шахматного" шума.
func valueNoise2D(x, y float64, seed int) float64 {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1

	sx := x - float64(x0)
	sy := y - float64(y0)

	n00 := randomValue2D(x0, y0, seed)
	n10 := randomValue2D(x1, y0, seed)
	n01 := randomValue2D(x0, y1, seed)
	n11 := randomValue2D(x1, y1, seed)

	ux := fade(sx)
	uy := fade(sy)

	ix0 := lerp(n00, n10, ux)
	ix1 := lerp(n01, n11, ux)

	return lerp(ix0, ix1, uy)
}

// fractalNoise2D суммирует несколько октав value noise
// и возвращает итоговое значение в диапазоне 0..1.
//
// octaves     — сколько слоёв шума суммировать;
// persistence — насколько быстро уменьшается вклад каждой следующей октавы;
// lacunarity  — насколько быстро растёт частота каждой следующей октавы.
//
// Идея такая:
//   - первая октава задаёт крупную форму;
//   - следующие добавляют всё более мелкие детали.
func fractalNoise2D(x, y float64, seed, octaves int, persistence, lacunarity float64) float64 {
	total := 0.0
	amplitude := 1.0
	frequency := 1.0
	maxAmplitude := 0.0

	for i := 0; i < octaves; i++ {
		n := valueNoise2D(x*frequency, y*frequency, seed+i*101)

		total += n * amplitude
		maxAmplitude += amplitude

		amplitude *= persistence
		frequency *= lacunarity
	}

	if maxAmplitude == 0 {
		return 0
	}

	return total / maxAmplitude
}

// terrainValue возвращает базовое значение местности в диапазоне 0..99.
//
// Для формы мира используем многослойный noise:
// крупная октава задаёт большие области,
// дополнительные октавы делают края и переходы менее примитивными.
func terrainValue(worldX, worldY, seed int) int {
	scale := 24.0

	n := fractalNoise2D(
		float64(worldX)/scale,
		float64(worldY)/scale,
		seed,
		4,   // octaves
		0.5, // persistence
		2.0, // lacunarity
	)

	return int(n * 100)
}

// detailValue возвращает дополнительное значение в диапазоне 0..99.
//
// Этот noise используется для вариаций внутри уже выбранного слоя:
// например, где будет обычный пол, а где трава.
func detailValue(worldX, worldY, seed int) int {
	scale := 10.0

	n := fractalNoise2D(
		float64(worldX)/scale,
		float64(worldY)/scale,
		seed+1337,
		3,    // octaves
		0.55, // persistence
		2.0,  // lacunarity
	)

	return int(n * 100)
}

// isBlockedTile решает, должна ли клетка быть непроходимой.
//
// На этом этапе мир должен быть в основном проходимым,
// поэтому blocked-зона должна занимать меньшую часть карты.
// Здесь мы сознательно делаем широкую "сушу" посередине
// и только по краям диапазона получаем препятствия.
func isBlockedTile(terrain int) bool {
	return terrain < 26 || terrain > 75
}

// blockedTileType выбирает тип непроходимого тайла.
//
// Низкие значения terrain превращаем в воду,
// высокие — в стены.
func blockedTileType(terrain int) TileType {
	if terrain < 26 {
		return TileWater
	}

	return TileWall
}

// walkableTileType выбирает тип проходимого тайла.
//
// Основная масса — обычный пол,
// часть клеток — трава как проходимая вариация поверхности.
func walkableTileType(detail int) TileType {
	if detail > 68 {
		return TileGrass
	}

	return TileFloor
}

// generateTile определяет итоговый тайл мировой клетки.
//
// Новый подход двухслойный:
//  1. сначала определяем, клетка проходимая или нет;
//  2. затем выбираем конкретный тип тайла внутри этого слоя.
//
// Это удобнее старой схемы, где мы сразу выбирали TileFloor/TileGrass/TileWater/TileWall.
// Теперь гораздо проще контролировать:
//   - сколько в мире проходимого пространства;
//   - сколько препятствий;
//   - какие именно тайлы должны появляться внутри каждой группы.
func generateTile(worldX, worldY, seed int) TileType {
	// Стартовую область вокруг игрока оставляем свободной,
	// чтобы старт всегда был безопасным и удобным для движения.
	if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
		return TileFloor
	}

	terrain := terrainValue(worldX, worldY, seed)

	if isBlockedTile(terrain) {
		return blockedTileType(terrain)
	}

	detail := detailValue(worldX, worldY, seed)
	return walkableTileType(detail)
}
