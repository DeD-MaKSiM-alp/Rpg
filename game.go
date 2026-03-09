package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	debugShowChunkOverlay = true

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
	world World
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
}

// Update — главный "тик" логики игры.
// Ebiten вызывает этот метод каждый кадр:
// здесь мы читаем состояние клавиш, обновляем буфер направления
// и при необходимости двигаем игрока по сетке.
func (g *Game) Update() error {
	/*Сначала читаем новый ввод
	Например:
	нажали вправо → {1, 0}
	ничего не нажали → {0, 0}
	*/
	newDirection := g.readDirectionInput()

	/*Если буфер уже активен
	это означает:
	мы уже ждём, не добавится ли вторая клавиша для диагонали
	*/
	if g.hasBufferedInput {
		//Если во время ожидания пришёл ещё ввод
		if newDirection.dx != 0 || newDirection.dy != 0 {
			g.bufferedDirection = mergeDirections(g.bufferedDirection, newDirection)
		}
		//Каждый тик уменьшаем таймер
		g.bufferTicksLeft--

		//Когда время вышло — двигаем игрока
		if g.bufferTicksLeft <= 0 {
			g.TryMovePlayer(g.bufferedDirection.dx, g.bufferedDirection.dy)
			g.hasBufferedInput = false
		}

		g.updateCamera()
		// Заранее создаём чанки вокруг игрока,
		// чтобы соседние области мира уже были готовы,
		// когда игрок приблизится к их границе.
		g.world.PreloadChunksAround(g.player.gridX, g.player.gridY, chunkPreloadRadius)
		// После предзагрузки очищаем слишком дальние чанки,
		// чтобы память не росла бесконечно.
		// Это первый шаг к поддержке очень большого мира.
		g.world.UnloadChunksFarFrom(g.player.gridX, g.player.gridY, chunkUnloadRadius)
		return nil
	}

	/*Если буфера ещё нет
	Это означает:
	увидели первое направление
	сохранили его
	начали короткое ожидание
	*/
	if newDirection.dx != 0 || newDirection.dy != 0 {
		g.bufferedDirection = newDirection
		g.bufferTicksLeft = inputBufferTicks
		g.hasBufferedInput = true
	}
	g.updateCamera()
	// Заранее создаём чанки вокруг игрока,
	// чтобы соседние области мира уже были готовы,
	// когда игрок приблизится к их границе.
	g.world.PreloadChunksAround(g.player.gridX, g.player.gridY, chunkPreloadRadius)
	// После предзагрузки очищаем слишком дальние чанки,
	// чтобы память не росла бесконечно.
	// Это первый шаг к поддержке очень большого мира.
	g.world.UnloadChunksFarFrom(g.player.gridX, g.player.gridY, chunkUnloadRadius)
	return nil
}

// Draw — метод, который отвечает за отрисовку одного кадра игры.
// Ebiten каждый кадр даёт нам поверхность screen,
// на которой мы сначала рисуем фон и сетку, а затем — игрока и остальные объекты.
func (g *Game) Draw(screen *ebiten.Image) {
	// рисуем фон в черном цвете
	screen.Fill(color.Black)

	// Сначала рисуем сам мир.
	g.world.Draw(screen, g.cameraX, g.cameraY)

	// Затем обычную сетку клеток.
	g.drawGrid(screen)

	// После этого, при включённом debug-режиме,
	// рисуем границы чанков поверх мира и сетки.
	if debugShowChunkOverlay {
		g.world.DrawChunkDebugOverlay(screen, g.cameraX, g.cameraY)
	}

	// Игрока рисуем уже поверх мира и всех сеток,
	// чтобы он не терялся за линиями.
	g.player.Draw(screen, g.cameraX, g.cameraY)

	// В самом конце рисуем текстовую debug-информацию.
	if debugShowChunkOverlay {
		g.drawDebugInfo(screen)
	}
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
		playerChunk.x,
		playerChunk.y,
		g.world.LoadedChunkCount(),
		g.world.seed,
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

// функция проверяет правила мира и возвращает значение перемещения
func (g *Game) TryMovePlayer(dx, dy int) {
	//Считает, куда игрок хочет пойти
	nextX := g.player.gridX + dx
	nextY := g.player.gridY + dy

	//Проверяет клетку на ходибельность

	if !g.world.IsWalkable(nextX, nextY) {
		return
	}
	//Если всё нормально — двигает игрока
	g.player.Move(dx, dy)
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
