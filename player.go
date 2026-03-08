package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Константы, относящиеся к игроку.
const (
	// playerSize — логический размер квадрата игрока в пикселях.
	// Фактическая позиция на экране зависит от gridX/gridY и размера клетки tileSize.
	playerSize = 32
)

// Player — сущность, которая описывает состояние игрока в игре.
// Она знает:
//   - в какой клетке сетки находится (gridX, gridY);
//   - какого размера его визуальное представление (size).
//
// Логику ввода игрок НЕ содержит — ей занимается структура Game,
// которая уже решает, когда вызвать методы Update/Move/Draw у Player.
type Player struct {
	gridX int     // позиция игрока по оси X в координатах сетки (номер клетки)
	gridY int     // позиция игрока по оси Y в координатах сетки (номер клетки)
	size  float64 // размер квадрата игрока в пикселях
}

// NewPlayer — фабричная функция для создания нового игрока.
// На вход она получает координаты в сетке (gridX, gridY),
// а на выходе отдаёт готовую структуру Player с заданным размером.
func NewPlayer(gridX int, gridY int) *Player {
	return &Player{
		gridX: gridX,
		gridY: gridY,
		size:  playerSize,
	}
}

// Update — метод для обновления внутреннего состояния игрока каждый кадр.
// Сейчас он пустой, но сюда можно будет добавлять, например, анимации или эффекты.
func (p *Player) Update() {

}

func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY int) {
	screenX := float32((p.gridX-cameraX)*tileSize) + float32(tileSize)/2 - float32(p.size)/2
	screenY := float32((p.gridY-cameraY)*tileSize) + float32(tileSize)/2 - float32(p.size)/2

	vector.FillRect(screen, screenX, screenY, float32(p.size), float32(p.size), color.White, false)
}

// Move сдвигает игрока по сетке.
// dx и dy — смещения в клетках: например, (1, 0) — один шаг вправо,
// (‑1, 1) — шаг влево и вниз (диагональ).
func (p *Player) Move(dx, dy int) {
	p.gridX += dx
	p.gridY += dy
}
