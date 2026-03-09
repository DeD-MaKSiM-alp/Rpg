package world

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
func (w *World) newChunk(chunkX, chunkY, seed int) *Chunk {
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

	chunk.pickups = generatePickupsForChunk(w, chunkX, chunkY, seed, chunk.tiles)
	w.generateEnemiesForChunk(chunkX, chunkY, seed, chunk.tiles)
	return chunk
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
	chunk := w.newChunk(coord.X, coord.Y, w.seed)

	// Сохраняем новый чанк в world,
	// чтобы в следующий раз не создавать его повторно.
	w.chunks[coord] = chunk

	return chunk
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
	for chunkY := centerCoord.Y - radius; chunkY <= centerCoord.Y+radius; chunkY++ {
		for chunkX := centerCoord.X - radius; chunkX <= centerCoord.X+radius; chunkX++ {
			coord := ChunkCoord{X: chunkX, Y: chunkY}

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
		dx := coord.X - centerCoord.X
		if dx < 0 {
			dx = -dx
		}

		dy := coord.Y - centerCoord.Y
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
