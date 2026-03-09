package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	inputpkg "mygame/internal/input"
	playerpkg "mygame/internal/player"
	"mygame/world"
)

// Константы, описывающие конфигурацию игры:
// размеры окна, размер клетки сетки и параметры буфера ввода.
const (
	// screenWidth — логическая ширина игрового экрана в пикселях.
	// Через неё рассчитывается ширина сетки и пределы, в которых можно рисовать и двигать объекты.
	screenWidth = 800

	// screenHeight — логическая высота игрового экрана в пикселях.
	// Аналогично ширине, используется для сетки, ограничений движения и размеров окна.
	screenHeight = 600

	// tileSize — длина стороны одной "клетки" карты в пикселях.
	// По нему мы переводим координаты в сетке (gridX, gridY) в экранные координаты.
	tileSize = 48

	// Сколько клеток помещается на экране
	visibleTilesX = screenWidth / tileSize
	visibleTilesY = screenHeight / tileSize

	// inputBufferTicks — сколько кадров (тиков) мы ждём,
	// чтобы "добрать" вторую клавишу и превратить одиночное нажатие в диагональное движение.
	// При 60 FPS значение 6 даёт примерно 100 мс "окно" для набора диагонали.
	inputBufferTicks = 6

	// debugShowChunkOverlay — включает отладочную отрисовку чанков:
	// границы чанков и текстовую информацию поверх игрового кадра.
	debugShowChunkOverlay = false

	// chunkPreloadRadius — радиус предзагрузки чанков вокруг игрока.
	// Эти чанки создаются заранее, чтобы при движении по миру
	// соседние области уже были готовы к отрисовке и логике.
	chunkPreloadRadius = 1

	// chunkUnloadRadius — радиус, внутри которого чанки сохраняются в памяти.
	// Всё, что дальше этого радиуса от игрока, будет выгружаться.
	//
	// Важно:
	// радиус выгрузки лучше делать больше, чем радиус предзагрузки,
	// чтобы чанки не создавались и не удалялись слишком агрессивно.
	chunkUnloadRadius = 2
)

/*
GameMode представляет состояние игры:
- ModeExplore — исследовательский режим, где игрок может двигаться и собирать предметы;
- ModeBattle — боевой режим, где игрок может сражаться с врагом.
*/
type GameMode int

/*
Константы для GameMode:
- ModeExplore — исследовательский режим, где игрок может двигаться и собирать предметы;
- ModeBattle — боевой режим, где игрок может сражаться с врагом.
*/
const (
	ModeExplore GameMode = iota
	ModeBattle
)

// Game — основная структура, описывающая состояние всей игры.
// В ней мы храним:
//   - игрока;
//   - временный буфер ввода направления (для более плавного диагонального движения);
//   - счётчик "оставшегося" времени буфера;
//   - флаг активности буфера.
//
// Экземпляр Game передаётся в ebiten, который вызывает её методы Update/Draw/Layout каждый кадр.
type Game struct {
	// player — объект, который отвечает за положение и отрисовку игрока на сетке.
	player playerpkg.Player

	// world — объект мира.
	// Game не знает, как именно мир устроен внутри:
	// одной картой, чанками или позже более сложной генерацией.
	// Он просто обращается к миру за данными:
	// можно ли пройти, как рисовать мир и как работать с чанками.
	world *world.World
	// bufferedDirection — текущее "собранное" направление движения игрока.
	bufferedDirection inputpkg.Direction

	// bufferTicksLeft — сколько тиков (кадров) ещё остаётся ждать вторую клавишу,
	// прежде чем применить накопленное направление к игроку.
	bufferTicksLeft int

	// hasBufferedInput — флаг, показывающий, активен ли сейчас буфер направления.
	// Если false, значит мы ждём первое нажатие и ещё не начали "окно ожидания".
	hasBufferedInput bool

	// cameraX и cameraY — координаты верхнего левого угла видимой области мира,
	// выраженные в клетках, а не в пикселях.
	// То есть камера показывает прямоугольную область мира
	// размером visibleTilesX на visibleTilesY клеток.
	cameraX int
	cameraY int

	pickupCount int

	hudFace *text.GoTextFace

	mode GameMode

	// battle хранит состояние текущего боя.
	// nil означает, что активного боя сейчас нет.
	battle *battlepkg.BattleContext
}

// updateStreamingWorld поддерживает "ленивый" мир вокруг игрока:
// подгружает ближайшие чанки и выгружает слишком дальние.
func (g *Game) updateStreamingWorld() {
	g.world.PreloadChunksAround(g.player.GridX, g.player.GridY, chunkPreloadRadius)
	g.world.UnloadChunksFarFrom(g.player.GridX, g.player.GridY, chunkUnloadRadius)
}

// drawDebugInfo рисует поверх кадра текстовую отладочную информацию.
//
// Здесь мы показываем:
//   - координаты игрока в мире;
//   - координаты текущего чанка игрока;
//   - количество чанков, которые уже загружены в память;
//   - seed мира.
//
// Это помогает быстро проверять,
// как работает перемещение по миру и система чанков.
func (g *Game) drawDebugInfo(screen *ebiten.Image) {
	playerChunk := g.world.ChunkCoordAt(g.player.GridX, g.player.GridY)

	debugText := fmt.Sprintf(
		"Player: (%d, %d)\nChunk: (%d, %d)\nLoaded chunks: %d\nSeed: %d",
		g.player.GridX,
		g.player.GridY,
		playerChunk.X,
		playerChunk.Y,
		g.world.LoadedChunkCount(),
		g.world.Seed(),
	)

	ebitenutil.DebugPrint(screen, debugText)
}

// Layout сообщает ebiten, какой логический размер экрана мы хотим использовать.
// Параметры w и h — это текущий физический размер окна,
// а возвращаемые screenWidth/screenHeight — "виртуальные" размеры,
// в пределах которых мы рисуем игру.
func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

// drawGrid рисует фоновые линии сетки по всему экрану.
// Сетка нужна, чтобы визуально обозначить клетки, по которым перемещается игрок.
func (g *Game) drawGrid(screen *ebiten.Image) {
	// цвет линий сетки
	gridColor := color.RGBA{R: 60, G: 60, B: 60, A: 255}

	// Вертикальные линии: рисуем их через каждый tileSize пикселей
	// и при необходимости отдельно добавляем правую границу окна.
	for x := 0; x <= screenWidth; x += tileSize {
		screenX := float32(x)
		vector.StrokeLine(screen, screenX, 0, screenX, float32(screenHeight), 1, gridColor, false)
	}
	if screenWidth%tileSize != 0 {
		screenX := float32(screenWidth)
		vector.StrokeLine(screen, screenX, 0, screenX, float32(screenHeight), 1, gridColor, false)
	}

	// Горизонтальные линии: аналогично вертикальным —
	// линии через каждый tileSize и дополнительная нижняя граница окна.
	for y := 0; y <= screenHeight; y += tileSize {
		screenY := float32(y)
		vector.StrokeLine(screen, 0, screenY, float32(screenWidth), screenY, 1, gridColor, false)
	}
	if screenHeight%tileSize != 0 {
		screenY := float32(screenHeight)
		vector.StrokeLine(screen, 0, screenY, float32(screenWidth), screenY, 1, gridColor, false)
	}
}

/*
startBattle начинает новый боевой режим с указанным врагом.
- Устанавливает режим игры в ModeBattle;
- Создаёт новый контекст боя;
- Сбрасывает буфер ввода движения, чтобы после выхода из боя старый ввод не сработал неожиданно.
*/
func (g *Game) startBattle(enemyID world.EntityID) {
	enc, ok := battlepkg.BuildEncounterFromWorld(g.world, enemyID)
	if !ok {
		return
	}
	g.mode = ModeBattle
	g.battle = battlepkg.BuildBattleContextFromEncounter(enc)

	g.hasBufferedInput = false
	g.bufferTicksLeft = 0
	g.bufferedDirection = inputpkg.Direction{}
}

/*
endBattle завершает боевой режим и возвращает игру в режим исследования мира.
*/
func (g *Game) endBattle() {
	g.mode = ModeExplore

	// Полностью очищаем контекст боя,
	// потому что после завершения старое состояние нам больше не нужно.
	g.battle = nil
}

// updateCamera обновляет положение камеры.
//
// Теперь мир бесконечный, поэтому камера больше не ограничивается
// размерами карты справа, снизу, слева или сверху.
// Она просто старается держать игрока примерно в центре экрана.
//
// cameraX и cameraY по-прежнему выражены в клетках мира,
// а не в пикселях.
func (g *Game) updateCamera() {
	g.cameraX = g.player.GridX - visibleTilesX/2
	g.cameraY = g.player.GridY - visibleTilesY/2
}

// TryMovePlayer пытается переместить игрока на одну клетку в заданном направлении.
func (g *Game) TryMovePlayer(dx, dy int) {
	moved, enemyID, pickedUp := playerpkg.TryMovePlayer(&g.player, g.world, dx, dy)
	if pickedUp {
		g.pickupCount++
	}
	if !moved {
		return
	}
	if enemyID != 0 {
		g.startBattle(enemyID)
	}
}
