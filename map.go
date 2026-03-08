package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Это тип клетки карты.
type TileType int

// Мы вводим два типа клеток:
const (
	TileFloor TileType = iota //пол, по нему можно ходить
	TileWall                  //стена, по ней нельзя
)

/*
Карта хранит:
width
height
tiles
tiles — это двумерный массив клеток.
*/
type Map struct {
	width  int
	height int
	tiles  [][]TileType
}

func NewMap(width, height int) Map {
	//создаёт карту нужного размера
	tiles := make([][]TileType, height)
	//заполняет все клетки полом
	for y := 0; y < height; y++ {
		tiles[y] = make([]TileType, width)

		for x := 0; x < width; x++ {
			tiles[y][x] = TileFloor
		}
	}
	//делает рамку из стен по краям
	for x := 0; x < width; x++ {
		tiles[0][x] = TileWall
		tiles[height-1][x] = TileWall
	}
	//делает рамку из стен по краям
	for y := 0; y < height; y++ {
		tiles[y][0] = TileWall
		tiles[y][width-1] = TileWall
	}

	return Map{
		width:  width,
		height: height,
		tiles:  tiles,
	}
}

/*
Метод отвечает на вопрос:
клетка (x, y) вообще существует внутри карты?
*/
func (m *Map) IsInside(x, y int) bool {
	return x >= 0 && x < m.width && y >= 0 && y < m.height
}

/*
Что делает этот метод
Он проверяет:
клетка внутри карты?
если да — это пол?
Если клетка стена или вне карты, вернёт false.
*/
func (m *Map) IsWalkable(x, y int) bool {
	if !m.IsInside(x, y) {
		return false
	}

	return m.tiles[y][x] == TileFloor
}

func (m *Map) Draw(screen *ebiten.Image, cameraX, cameraY int) {
	floorColor := color.RGBA{R: 30, G: 30, B: 30, A: 255}
	wallColor := color.RGBA{R: 90, G: 90, B: 90, A: 255}

	endX := cameraX + visibleTilesX
	endY := cameraY + visibleTilesY

	if endX > m.width {
		endX = m.width
	}
	if endY > m.height {
		endY = m.height
	}

	for y := cameraY; y < endY; y++ {
		for x := cameraX; x < endX; x++ {
			screenX := float32((x - cameraX) * tileSize)
			screenY := float32((y - cameraY) * tileSize)

			tileColor := floorColor
			if m.tiles[y][x] == TileWall {
				tileColor = wallColor
			}

			vector.FillRect(screen, screenX, screenY, float32(tileSize), float32(tileSize), tileColor, false)
		}
	}
}
