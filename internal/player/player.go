package player

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mygame/internal/visualcolor"
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
// Визуально — токен «герой» в общем языке UI: золотой круг, рамки, внутренний маркер.
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY, tileSize int) {
	ts := float32(tileSize)
	cx := float32((p.GridX-cameraX)*tileSize) + ts*0.5
	cy := float32((p.GridY-cameraY)*tileSize) + ts*0.5
	r := ts * 0.38
	if r < 9 {
		r = 9
	}
	vector.FillCircle(screen, cx, cy, r, visualcolor.Foundation.ActiveTurn, false)
	vector.StrokeCircle(screen, cx, cy, r, 2.5, visualcolor.Foundation.PostBattleBorder, false)
	vector.StrokeCircle(screen, cx, cy, r-3, 1.25, visualcolor.Foundation.AccentStrip, false)
	vector.FillCircle(screen, cx, cy, r*0.32, visualcolor.Foundation.PanelBGDeep, false)
}

// Position возвращает текущие координаты игрока на сетке (GridX, GridY).
func (p *Player) Position() (x, y int) {
	return p.GridX, p.GridY
}

// Move сдвигает игрока на (dx, dy) клеток.
func (p *Player) Move(dx, dy int) {
	p.GridX += dx
	p.GridY += dy
}
