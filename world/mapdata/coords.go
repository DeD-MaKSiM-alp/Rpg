package mapdata

// FloorDiv — целочисленное деление с округлением вниз (для отрицательных координат).
func FloorDiv(a, b int) int {
	result := a / b
	remainder := a % b
	if remainder != 0 && ((remainder > 0) != (b > 0)) {
		result--
	}
	return result
}

// PositiveMod — неотрицательный остаток от деления.
func PositiveMod(a, b int) int {
	result := a % b
	if result < 0 {
		result += b
	}
	return result
}

// WorldToChunkLocal переводит мировые координаты в координаты чанка и локальные в чанке.
func WorldToChunkLocal(worldX, worldY int) (ChunkCoord, int, int) {
	chunkX := FloorDiv(worldX, ChunkSize)
	chunkY := FloorDiv(worldY, ChunkSize)
	localX := PositiveMod(worldX, ChunkSize)
	localY := PositiveMod(worldY, ChunkSize)
	return ChunkCoord{X: chunkX, Y: chunkY}, localX, localY
}
