package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Direction задаёт направление движения (смещение по сетке).
// Нулевое значение Direction{} означает отсутствие направления.
type Direction struct {
	Dx, Dy int
}

// Input считывает ввод игрока для режима исследования (движение, ожидание).
type Input struct{}

// New создаёт новый экземпляр Input.
func New() *Input {
	return &Input{}
}

// ReadExploreInput возвращает действие игрока в этом кадре: смещение (dx, dy) при нажатии стрелки или wait=true при SPACE.
// Одно нажатие — одно действие за кадр. Приоритет: стрелка, затем SPACE.
func (i *Input) ReadExploreInput() (dx, dy int, wait bool) {
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		return 1, 0, false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		return -1, 0, false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		return 0, 1, false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		return 0, -1, false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return 0, 0, true
	}
	return 0, 0, false
}
