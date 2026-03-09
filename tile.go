package main

// TileType — тип клетки мира.
// Через этот тип мы описываем, что лежит в каждой клетке:
// пол, стена и позже другие виды тайлов.
type TileType int

const (
	// TileFloor — проходимая клетка.
	TileFloor TileType = iota

	// TileWall — непроходимая клетка.
	TileWall
)
