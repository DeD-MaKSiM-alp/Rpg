package main

// TileType — тип клетки мира.
// Через этот тип мы описываем, что лежит в каждой клетке:
// пол, стена и позже другие виды тайлов.
type TileType int

const (
	// TileFloor — обычный проходимый пол.
	TileFloor TileType = iota

	// TileWall — стена, непроходимая клетка.
	TileWall

	// TileGrass — трава, тоже проходимая клетка.
	TileGrass

	// TileWater — вода, непроходимая клетка.
	TileWater
)
