package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Direction представляет намерение игрока двигаться по сетке.
// Dx — смещение по горизонтали (‑1, 0, 1),
// Dy — смещение по вертикали (‑1, 0, 1).
type Direction struct {
	Dx int
	Dy int
}

// ReadDirectionInput считывает "свежее" состояние клавиш движения
// и возвращает направление, в котором игрок хочет сдвинуться в этом кадре.
func ReadDirectionInput() Direction {
	dx := 0
	dy := 0

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		dx += 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		dx -= 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		dy += 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		dy -= 1
	}

	return Direction{Dx: dx, Dy: dy}
}

// MergeDirections объединяет два направления движения.
func MergeDirections(a, b Direction) Direction {
	result := a
	if b.Dx != 0 {
		result.Dx = b.Dx
	}
	if b.Dy != 0 {
		result.Dy = b.Dy
	}
	return result
}

// StartInputBufferIfNeeded запускает новый буфер направления,
// если игрок только что нажал кнопку движения.
func StartInputBufferIfNeeded(newDirection Direction, buffered *Direction, ticksLeft *int, hasBuffered *bool, bufferTicks int) {
	if newDirection.Dx == 0 && newDirection.Dy == 0 {
		return
	}
	*buffered = newDirection
	*ticksLeft = bufferTicks
	*hasBuffered = true
}

// UpdateBufferedInput обновляет уже активный буфер направления
// и при необходимости вызывает tryMove с накопленным направлением.
func UpdateBufferedInput(newDirection Direction, buffered *Direction, ticksLeft *int, hasBuffered *bool, tryMove func(dx, dy int)) {
	if newDirection.Dx != 0 || newDirection.Dy != 0 {
		*buffered = MergeDirections(*buffered, newDirection)
	}
	*ticksLeft--
	if *ticksLeft <= 0 {
		tryMove(buffered.Dx, buffered.Dy)
		*hasBuffered = false
	}
}
