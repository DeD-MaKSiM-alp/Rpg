package player

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const playerSize = 32

// Player — сущность, которая описывает состояние игрока в игре.
type Player struct {
	GridX int
	GridY int
	size  float64
}

// NewPlayer создаёт нового игрока в заданной клетке сетки.
func NewPlayer(gridX, gridY int) *Player {
	return &Player{
		GridX: gridX,
		GridY: gridY,
		size:  playerSize,
	}
}

// Update обновляет внутреннее состояние игрока каждый кадр.
func (p *Player) Update() {}

// Draw рисует игрока на экране (cameraX, cameraY и tileSize в клетках/пикселях).
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY, tileSize int) {
	screenX := float32((p.GridX-cameraX)*tileSize) + float32(tileSize)/2 - float32(p.size)/2
	screenY := float32((p.GridY-cameraY)*tileSize) + float32(tileSize)/2 - float32(p.size)/2
	vector.FillRect(screen, screenX, screenY, float32(p.size), float32(p.size), color.White, false)
}

// Move сдвигает игрока на (dx, dy) клеток.
func (p *Player) Move(dx, dy int) {
	p.GridX += dx
	p.GridY += dy
}
