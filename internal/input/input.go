package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input считывает ввод игрока для режима исследования (движение, ожидание).
// Цепочка: physical axes (raw) → effective direction (= raw, без подстановки оси) →
// смена направления = немедленный emit; то же направление = controlled repeat.
type Input struct {
	// Счётчик кадров (увеличивается при каждом вызове ReadExploreInput).
	frameCounter int

	// Последнее выданное эффективное направление и сколько кадров оно удерживается.
	// Repeat: первый шаг сразу, следующий — после initialDelayFrames, далее каждые repeatIntervalFrames.
	lastEffectiveDX, lastEffectiveDY int
	holdFrames                       int

	// Момент последнего изменения значения по каждой оси и последнее ненулевое значение.
	// Нужно для grace: при кратковременном отпускании клавиши сохраняем направление в пределах окна.
	lastHorizontalChangeFrame int
	lastVerticalChangeFrame   int
	lastHorizontalValue      int // последнее ненулевое rawX
	lastVerticalValue        int // последнее ненулевое rawY
	prevRawDX, prevRawDY      int

	// Параметры repeat и grace: задаются в New(), легко подстроить.
	InitialDelayFrames   int // кадров до первого повтора при удержании
	RepeatIntervalFrames int // кадров между повторами
	DiagonalGraceFrames  int // окно добора второй оси (в кадрах); смена на диагональ без штрафа по repeat

	// Временный debug: последние raw и выданное направление (для overlay).
	debugRawDX, debugRawDY   int
	debugEmitDX, debugEmitDY int
}

// New создаёт новый экземпляр Input с дефолтными параметрами repeat и grace.
func New() *Input {
	return &Input{
		InitialDelayFrames:   12, // ~200 ms при 60 FPS до первого повтора
		RepeatIntervalFrames: 5,  // ~83 ms между повторами
		DiagonalGraceFrames:  10, // окно: добор второй оси и устойчивость при кратком отпускании (~167 ms при 60 FPS)
	}
}

// readRawAxes возвращает текущие физические оси по удержанию стрелок.
// Right=+1, Left=-1, обе=0; Down=+1, Up=-1, обе=0. Opposite keys cancel.
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

// updateAxisChangeFrames обновляет last*ChangeFrame и last*Value при изменении оси.
func (i *Input) updateAxisChangeFrames(rawX, rawY int) {
	if rawX != i.prevRawDX {
		i.lastHorizontalChangeFrame = i.frameCounter
	}
	if rawY != i.prevRawDY {
		i.lastVerticalChangeFrame = i.frameCounter
	}
	if rawX != 0 {
		i.lastHorizontalValue = rawX
	}
	if rawY != 0 {
		i.lastVerticalValue = rawY
	}
	i.prevRawDX, i.prevRawDY = rawX, rawY
}

// effectiveDirection: raw + grace. Если ось стала 0 недавно (в пределах DiagonalGraceFrames),
// считаем её ещё «удержанной» — так диагональ не дёргается при кратком отпускании и проще добор второй клавиши.
// Не подставляем ось, которую игрок никогда не нажимал (last*Value был только при реальном нажатии).
func (i *Input) effectiveDirection(rawX, rawY int) (effX, effY int) {
	effX = rawX
	effY = rawY
	if rawX == 0 && i.lastHorizontalValue != 0 && (i.frameCounter-i.lastHorizontalChangeFrame) <= i.DiagonalGraceFrames {
		effX = i.lastHorizontalValue
	}
	if rawY == 0 && i.lastVerticalValue != 0 && (i.frameCounter-i.lastVerticalChangeFrame) <= i.DiagonalGraceFrames {
		effY = i.lastVerticalValue
	}
	return effX, effY
}

// directionChanged — true, если эффективное направление отличается от последнего выданного.
// При смене направления всегда emit сразу, без repeat delay.
func (i *Input) directionChanged(effX, effY int) bool {
	return effX != i.lastEffectiveDX || effY != i.lastEffectiveDY
}

// shouldEmitRepeat — true, если при удержании того же направления пора выдать повтор:
// первый повтор после InitialDelayFrames, далее каждые RepeatIntervalFrames.
func (i *Input) shouldEmitRepeat() bool {
	firstRepeatFrame := i.InitialDelayFrames + 1
	if i.holdFrames < firstRepeatFrame {
		return false
	}
	if i.holdFrames == firstRepeatFrame {
		return true
	}
	return (i.holdFrames-firstRepeatFrame)%i.RepeatIntervalFrames == 0
}

// ReadExploreInput — единственная точка чтения ввода для режима исследования (explore).
// Контракт: (dx, dy int, wait bool). Детали grace/repeat скрыты внутри.
//
// Логика: движение по raw осям; смена направления = немедленный emit; удержание = controlled repeat.
// Wait проверяется только при отсутствии движения в этом кадре (приоритет движения над wait).
func (i *Input) ReadExploreInput() (dx, dy int, wait bool) {
	i.frameCounter++

	rawX, rawY := i.readRawAxes()
	i.debugRawDX, i.debugRawDY = rawX, rawY
	i.updateAxisChangeFrames(rawX, rawY)

	effX, effY := i.effectiveDirection(rawX, rawY)

	// Нет движения: сброс state. Wait только по just-pressed Space.
	if effX == 0 && effY == 0 {
		i.lastEffectiveDX, i.lastEffectiveDY = 0, 0
		i.holdFrames = 0
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			return 0, 0, true
		}
		return 0, 0, false
	}

	// Приоритет движения над wait. Смена направления — срабатывает сразу, repeat сбрасывается.
	if i.directionChanged(effX, effY) {
		i.lastEffectiveDX, i.lastEffectiveDY = effX, effY
		i.holdFrames = 1
		i.debugEmitDX, i.debugEmitDY = effX, effY
		return effX, effY, false
	}

	// То же направление удерживается — повтор только по правилам repeat.
	i.holdFrames++
	if i.shouldEmitRepeat() {
		i.debugEmitDX, i.debugEmitDY = i.lastEffectiveDX, i.lastEffectiveDY
		return i.lastEffectiveDX, i.lastEffectiveDY, false
	}
	return 0, 0, false
}

// DebugRaw возвращает последние считанные raw оси (для отладочного overlay).
func (i *Input) DebugRaw() (dx, dy int) { return i.debugRawDX, i.debugRawDY }

// DebugEmit возвращает последнее реально выданное направление (для отладочного overlay).
func (i *Input) DebugEmit() (dx, dy int) { return i.debugEmitDX, i.debugEmitDY }
