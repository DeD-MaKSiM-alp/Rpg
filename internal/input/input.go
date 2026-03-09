package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Параметры окна добора второй оси для диагонали (в кадрах).
const graceWindowFrames = 4

// Параметры controlled repeat: задержка до первого повтора и интервал повторов (в кадрах).
const (
	initialDelayFrames   = 15 // ~250 ms при 60 FPS
	repeatIntervalFrames = 5  // ~83 ms между повторами
)

// Input считывает ввод игрока для режима исследования (движение, ожидание).
// Внутри: physical axes → grace window → effective direction → repeat logic.
type Input struct {
	// Счётчик кадров (увеличивается при каждом вызове ReadExploreInput).
	frameCounter int

	// Последнее выданное эффективное направление и сколько кадров оно удерживается.
	lastEffectiveDX, lastEffectiveDY int
	holdFrames                       int

	// Grace window: последнее значение и кадр активации по каждой оси.
	lastHorizontalValue      int
	lastHorizontalPressFrame int
	lastVerticalValue        int
	lastVerticalPressFrame   int
}

// New создаёт новый экземпляр Input.
func New() *Input {
	return &Input{}
}

// readRawAxes возвращает текущие физические оси по удержанию стрелок.
// Right=+1, Left=-1, обе=0; Down=+1, Up=-1, обе=0.
func (i *Input) readRawAxes() (rawX, rawY int) {
	if ebiten.IsKeyPressed(ebiten.KeyRight) && !ebiten.IsKeyPressed(ebiten.KeyLeft) {
		rawX = 1
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) && !ebiten.IsKeyPressed(ebiten.KeyRight) {
		rawX = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) && !ebiten.IsKeyPressed(ebiten.KeyUp) {
		rawY = 1
	} else if ebiten.IsKeyPressed(ebiten.KeyUp) && !ebiten.IsKeyPressed(ebiten.KeyDown) {
		rawY = -1
	}
	return rawX, rawY
}

// updateAxisPressState обновляет last*Value и last*PressFrame для оси, если она активна.
func (i *Input) updateAxisPressState(rawX, rawY int) {
	if rawX != 0 {
		i.lastHorizontalValue = rawX
		i.lastHorizontalPressFrame = i.frameCounter
	}
	if rawY != 0 {
		i.lastVerticalValue = rawY
		i.lastVerticalPressFrame = i.frameCounter
	}
}

// applyGraceWindow возвращает эффективное направление: raw оси + добор второй оси в пределах окна.
// Если одна ось нуль, но вторая была активна недавно (в пределах graceWindowFrames), подставляем её.
func (i *Input) applyGraceWindow(rawX, rawY int) (effX, effY int) {
	effX = rawX
	effY = rawY
	if rawX == 0 && rawY != 0 && (i.frameCounter-i.lastHorizontalPressFrame) <= graceWindowFrames {
		effX = i.lastHorizontalValue
	}
	if rawY == 0 && rawX != 0 && (i.frameCounter-i.lastVerticalPressFrame) <= graceWindowFrames {
		effY = i.lastVerticalValue
	}
	return effX, effY
}

// computeEffectiveDirection возвращает эффективное направление: raw axes + grace window.
func (i *Input) computeEffectiveDirection() (effX, effY int) {
	rawX, rawY := i.readRawAxes()
	i.updateAxisPressState(rawX, rawY)
	return i.applyGraceWindow(rawX, rawY)
}

// directionChanged возвращает true, если эффективное направление отличается от последнего выданного.
func (i *Input) directionChanged(effX, effY int) bool {
	return effX != i.lastEffectiveDX || effY != i.lastEffectiveDY
}

// shouldEmitRepeat возвращает true, если при удержании того же направления пора выдать повтор:
// первый повтор после initialDelayFrames, далее каждые repeatIntervalFrames.
func (i *Input) shouldEmitRepeat() bool {
	firstRepeatFrame := initialDelayFrames + 1
	if i.holdFrames < firstRepeatFrame {
		return false
	}
	if i.holdFrames == firstRepeatFrame {
		return true
	}
	return (i.holdFrames-firstRepeatFrame)%repeatIntervalFrames == 0
}

// ReadExploreInput — единственная точка чтения ввода для режима исследования (explore).
// Контракт: (dx, dy int, wait bool). Детали grace window и repeat скрыты внутри.
func (i *Input) ReadExploreInput() (dx, dy int, wait bool) {
	i.frameCounter++

	effX, effY := i.computeEffectiveDirection()

	// Нет движения: сброс state, wait только по just-pressed Space.
	if effX == 0 && effY == 0 {
		i.lastEffectiveDX, i.lastEffectiveDY = 0, 0
		i.holdFrames = 0
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			return 0, 0, true
		}
		return 0, 0, false
	}

	// Приоритет: движение важнее wait. Effective direction изменился — срабатывает сразу, repeat сбрасывается.
	if i.directionChanged(effX, effY) {
		i.lastEffectiveDX, i.lastEffectiveDY = effX, effY
		i.holdFrames = 1
		return effX, effY, false
	}

	// То же направление удерживается — повтор только по правилам repeat.
	i.holdFrames++
	if i.shouldEmitRepeat() {
		return i.lastEffectiveDX, i.lastEffectiveDY, false
	}
	return 0, 0, false
}
