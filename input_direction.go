package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// readDirectionInput считывает "свежее" состояние клавиш движения
// и возвращает направление, в котором игрок хочет сдвинуться в этом кадре.
// Здесь мы ещё никого не двигаем — только собираем вектор dx/dy.
func (g *Game) readDirectionInput() Direction {
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

	return Direction{dx: dx, dy: dy}
}

// startInputBufferIfNeeded запускает новый буфер направления,
// если игрок только что нажал кнопку движения.
func (g *Game) startInputBufferIfNeeded(newDirection Direction) {
	if newDirection.dx == 0 && newDirection.dy == 0 {
		return
	}

	g.bufferedDirection = newDirection
	g.bufferTicksLeft = inputBufferTicks
	g.hasBufferedInput = true
}

// updateBufferedInput обновляет уже активный буфер направления
// и при необходимости двигает игрока.
func (g *Game) updateBufferedInput(newDirection Direction) {
	if newDirection.dx != 0 || newDirection.dy != 0 {
		g.bufferedDirection = mergeDirections(g.bufferedDirection, newDirection)
	}

	g.bufferTicksLeft--

	if g.bufferTicksLeft <= 0 {
		g.TryMovePlayer(g.bufferedDirection.dx, g.bufferedDirection.dy)
		g.hasBufferedInput = false
	}
}
