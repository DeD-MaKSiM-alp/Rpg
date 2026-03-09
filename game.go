package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

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

// Direction представляет намерение игрока двигаться по сетке.
// dx — смещение по горизонтали (‑1, 0, 1),
// dy — смещение по вертикали (‑1, 0, 1).
// Эта структура нужна, чтобы аккуратно собирать направление
// из нескольких нажатий клавиш (например, диагонали).
type Direction struct {
	dx int // изменение координаты игрока по X в клетках
	dy int // изменение координаты игрока по Y в клетках
}

type GameMode int

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
	player Player

	// world — объект мира.
	// Game не знает, как именно мир устроен внутри:
	// одной картой, чанками или позже более сложной генерацией.
	// Он просто обращается к миру за данными:
	// можно ли пройти, как рисовать мир и как работать с чанками.
	world *world.World
	// bufferedDirection — текущее "собранное" направление движения игрока.
	// Например: сначала нажали вправо → {1, 0}, затем успели нажать вниз → станет {1, 1}.
	bufferedDirection Direction

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

	hudFace *text.GoXFace

	mode GameMode

	// battle хранит состояние текущего боя.
	// nil означает, что активного боя сейчас нет.
	battle *BattleContext
}

func (g *Game) Update() error {
	if g.mode == ModeBattle {
		g.updateBattleMode()
		return nil
	}

	return g.updateExploreMode()
}

// updateExploreMode обрабатывает один кадр в режиме исследования мира:
// читает ввод, обновляет буфер направления, двигает игрока и поддерживает мир вокруг.
func (g *Game) updateExploreMode() error {
	newDirection := g.readDirectionInput()

	if g.hasBufferedInput {
		g.updateBufferedInput(newDirection)
	} else {
		g.startInputBufferIfNeeded(newDirection)
	}

	g.updateCamera()
	g.updateStreamingWorld()

	return nil
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

// updateStreamingWorld поддерживает "ленивый" мир вокруг игрока:
// подгружает ближайшие чанки и выгружает слишком дальние.
func (g *Game) updateStreamingWorld() {
	g.world.PreloadChunksAround(g.player.gridX, g.player.gridY, chunkPreloadRadius)
	g.world.UnloadChunksFarFrom(g.player.gridX, g.player.gridY, chunkUnloadRadius)
}

// Draw — метод, который отвечает за отрисовку одного кадра игры.
// Ebiten каждый кадр даёт нам поверхность screen,
// на которой мы последовательно рисуем:
//
//	фон, мир, сетку, (опционально) overlay чанков, игрока, debug-текст и HUD.
func (g *Game) Draw(screen *ebiten.Image) {
	// рисуем фон в черном цвете
	screen.Fill(color.Black)

	// Сначала рисуем сам мир.
	g.world.Draw(screen, g.cameraX, g.cameraY, visibleTilesX, visibleTilesY, tileSize)

	// Затем обычную сетку клеток.
	g.drawGrid(screen)

	// После этого, при включённом debug-режиме,
	// рисуем границы чанков поверх мира и сетки.
	if debugShowChunkOverlay {
		g.world.DrawChunkDebugOverlay(screen, g.cameraX, g.cameraY, visibleTilesX, visibleTilesY, tileSize, screenWidth, screenHeight)
	}

	// Игрока рисуем уже поверх мира и всех сеток,
	// чтобы он не терялся за линиями.
	g.player.Draw(screen, g.cameraX, g.cameraY)

	// При включённом debug-режиме поверх всего рисуем текстовую debug-информацию.
	if debugShowChunkOverlay {
		g.drawDebugInfo(screen)
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(10, 20)
	op.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		fmt.Sprintf("Pickups: %d", g.pickupCount),
		g.hudFace,
		op,
	)

	if g.mode == ModeBattle {
		g.drawBattleOverlay(screen)
	}
}

func (g *Game) drawBattleOverlay(screen *ebiten.Image) {
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	panelColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	panelBorderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	// Затемняем фон.
	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)

	// Центральная панель.
	panelX := float32(120)
	panelY := float32(140)
	panelW := float32(560)
	panelH := float32(220)

	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelColor, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2, panelBorderColor, false)

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(panelX)+20, float64(panelY)+35)
	titleOp.ColorScale.ScaleWithColor(color.White)

	title := "Battle mode"

	// Если контекст боя есть, показываем ID активного врага.
	// Это простой, но полезный шаг:
	// UI боя начинает читать данные из BattleContext.
	if g.battle != nil {
		title = fmt.Sprintf("Battle mode: enemy #%d", g.battle.EnemyID)
	}

	text.Draw(
		screen,
		title,
		g.hudFace,
		titleOp,
	)

	bodyOp := &text.DrawOptions{}
	bodyOp.GeoM.Translate(float64(panelX)+20, float64(panelY)+80)
	bodyOp.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		"B - win test battle and remove enemy",
		g.hudFace,
		bodyOp,
	)

	bodyOp2 := &text.DrawOptions{}
	bodyOp2.GeoM.Translate(float64(panelX)+20, float64(panelY)+115)
	bodyOp2.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		"Esc - leave battle mode without removing enemy",
		g.hudFace,
		bodyOp2,
	)
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
	playerChunk := g.world.ChunkCoordAt(g.player.gridX, g.player.gridY)

	debugText := fmt.Sprintf(
		"Player: (%d, %d)\nChunk: (%d, %d)\nLoaded chunks: %d\nSeed: %d",
		g.player.gridX,
		g.player.gridY,
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

// mergeDirections объединяет два направления движения.
// Используется, чтобы аккуратно "достроить" диагональное движение:
// если по одной оси уже есть значение, а по другой приходит новое — берём новое.
func mergeDirections(a, b Direction) Direction {
	result := a

	if b.dx != 0 {
		result.dx = b.dx
	}

	if b.dy != 0 {
		result.dy = b.dy
	}

	return result
}

func (g *Game) startBattle(enemyID world.EntityID) {
	g.mode = ModeBattle

	// Создаём новый контекст боя.
	// Пока он знает только ID врага,
	// но позже сюда добавятся все остальные боевые данные.
	g.battle = &BattleContext{
		EnemyID: enemyID,
	}

	// Сбрасываем буфер ввода движения,
	// чтобы после выхода из боя старый ввод не сработал неожиданно.
	g.hasBufferedInput = false
	g.bufferTicksLeft = 0
	g.bufferedDirection = Direction{}
}

func (g *Game) endBattle() {
	g.mode = ModeExplore

	// Полностью очищаем контекст боя,
	// потому что после завершения старое состояние нам больше не нужно.
	g.battle = nil
}

func (g *Game) updateBattleMode() {
	// Страховка:
	// если по какой-то причине игра находится в ModeBattle,
	// но контекст боя отсутствует, выходим обратно в исследование.
	if g.battle == nil {
		g.endBattle()
		return
	}

	// B = тестовая победа над врагом.
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.world.RemoveEnemy(g.battle.EnemyID)
		g.endBattle()
		return
	}

	// Escape = выйти из battle mode без победы.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.endBattle()
		return
	}
}

func (g *Game) TryMovePlayer(dx, dy int) {
	nextX := g.player.gridX + dx
	nextY := g.player.gridY + dy

	// Если в целевой клетке враг — не двигаемся,
	// а входим в режим боя.
	enemy := g.world.GetEnemyAt(nextX, nextY)
	if enemy != nil {
		g.startBattle(enemy.ID)
		return
	}

	if !g.world.IsWalkable(nextX, nextY) {
		return
	}

	g.player.Move(dx, dy)

	if g.world.CollectPickupAt(g.player.gridX, g.player.gridY) {
		g.pickupCount++
	}

	enemyID, startedBattle := g.world.AdvanceTurn(g.player.gridX, g.player.gridY)
	if startedBattle {
		g.startBattle(enemyID)
	}
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
	g.cameraX = g.player.gridX - visibleTilesX/2
	g.cameraY = g.player.gridY - visibleTilesY/2
}
