package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input считывает ввод игрока (движение, ожидание).
type Input struct{}

// New создаёт новый экземпляр Input.
func New() *Input {
	return &Input{}
}

// ConsumeDirection возвращает направление движения, если в этом кадре было нажатие, и true.
func (i *Input) ConsumeDirection() (Direction, bool) {
	dir := readDirectionInput()
	if dir.Dx != 0 || dir.Dy != 0 {
		return dir, true
	}
	return Direction{}, false
}

// WaitPressed возвращает true, если в этом кадре нажата клавиша ожидания (пробел).
// Использует IsKeyJustPressed, чтобы одно нажатие тратило ровно один ход.
func (i *Input) WaitPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

// readDirectionInput считывает "свежее" состояние клавиш движения.
func readDirectionInput() Direction {
	dx, dy := 0, 0
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		dx = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		dx = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		dy = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		dy = -1
	}
	return Direction{Dx: dx, Dy: dy}
}
