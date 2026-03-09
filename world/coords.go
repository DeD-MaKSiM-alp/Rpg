package world

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

	return ChunkCoord{X: chunkX, Y: chunkY}, localX, localY
}
