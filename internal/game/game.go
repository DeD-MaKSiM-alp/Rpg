package game

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
	"mygame/internal/ui"
	"mygame/world"
)

// DefaultWindowTitle — заголовок окна по умолчанию.
const DefaultWindowTitle = "My Game"

// ResolutionPreset задаёт один пресет разрешения окна.
type ResolutionPreset struct {
	Width  int
	Height int
}

// Пресеты разрешения. Переключение: поменять ActivePresetIndex ниже.
var ResolutionPresets = []ResolutionPreset{
	{800, 600},
	{1024, 768},
	{1280, 720},
	{1366, 768},
	{1600, 900},
}

// ActivePresetIndex — единственное место выбора разрешения перед запуском.
// 0=800x600, 1=1024x768, 2=1280x720, 3=1366x768, 4=1600x900.
// Для resizable window в будущем: подставлять сюда индекс по умолчанию, а фактические размеры брать из Layout(w,h).
const ActivePresetIndex = 0

// Viewport задаёт логический размер видимой области мира в тайлах.
// Игровая логика (камера, мир, сетка) опирается на viewport; размер окна в пикселях — отдельно (ScreenWidth/ScreenHeight).
type Viewport struct {
	WidthTiles  int
	HeightTiles int
}

// WorldViewport — текущий viewport мира. Задаётся в applyResolutionPreset() из выбранного пресета.
var WorldViewport Viewport

// Текущие размеры экрана в пикселях. Задаются только в applyResolutionPreset().
// Используются только: Layout(), ebiten.SetWindowSize, UI (HUD, battle overlay).
var (
	ScreenWidth  int
	ScreenHeight int
)

// applyResolutionPreset применяет активный пресет: задаёт ScreenWidth/ScreenHeight, Viewport и размер окна.
func applyResolutionPreset() {
	p := ResolutionPresets[ActivePresetIndex]
	ScreenWidth = p.Width
	ScreenHeight = p.Height
	WorldViewport.WidthTiles = ScreenWidth / tileSize
	WorldViewport.HeightTiles = ScreenHeight / tileSize
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
}

const (
	tileSize                = 48
	debugShowChunkOverlay   = false
	debugShowInputDirection = true // TODO: временный debug для проверки диагонали; удалить после проверки
	chunkPreloadRadius      = 1
	chunkUnloadRadius       = 2
)

// GameMode представляет состояние игры.
type GameMode int

const (
	ModeExplore GameMode = iota
	ModeBattle
)

// Game — основная структура, описывающая состояние всей игры.
type Game struct {
	player      playerpkg.Player
	world       *world.World
	input       *inputpkg.Input
	cameraX     int
	cameraY     int
	pickupCount int
	hudFace     *text.GoTextFace
	mode        GameMode
	battle      *battlepkg.BattleContext

	// Временный debug: последнее направление, возвращённое ReadExploreInput (только для отрисовки).
	debugInputDX, debugInputDY int
}

// NewGame создаёт новый экземпляр игры (мир, игрок, UI-шрифт и т.д.).
func NewGame(worldSeed, playerGridX, playerGridY int) *Game {
	return &Game{
		player:  *playerpkg.NewPlayer(playerGridX, playerGridY),
		world:   world.NewWorld(worldSeed),
		input:   inputpkg.New(),
		hudFace: ui.LoadHUDFace(),
		mode:    ModeExplore,
		battle:  nil,
	}
}

// Run настраивает окно, создаёт игру и запускает главный цикл ebiten.
// Точка входа для запуска из main. Возвращает ошибку от ebiten.RunGame.
func Run(worldSeed, playerGridX, playerGridY int, windowTitle string) error {
	applyResolutionPreset()
	if windowTitle != "" {
		ebiten.SetWindowTitle(windowTitle)
	} else {
		ebiten.SetWindowTitle(DefaultWindowTitle)
	}
	g := NewGame(worldSeed, playerGridX, playerGridY)
	return ebiten.RunGame(g)
}

// Layout сообщает ebiten логический размер экрана.
func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) updateStreamingWorld() {
	g.world.PreloadChunksAround(g.player.GridX, g.player.GridY, chunkPreloadRadius)
	g.world.UnloadChunksFarFrom(g.player.GridX, g.player.GridY, chunkUnloadRadius)
}

func (g *Game) drawDebugInfo(screen *ebiten.Image) {
	playerChunk := g.world.ChunkCoordAt(g.player.GridX, g.player.GridY)
	debugText := fmt.Sprintf(
		"Player: (%d, %d)\nChunk: (%d, %d)\nLoaded chunks: %d\nSeed: %d",
		g.player.GridX, g.player.GridY,
		playerChunk.X, playerChunk.Y,
		g.world.LoadedChunkCount(), g.world.Seed(),
	)
	ebitenutil.DebugPrint(screen, debugText)
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	gridColor := color.RGBA{R: 60, G: 60, B: 60, A: 255}
	wPx := WorldViewport.WidthTiles * tileSize
	hPx := WorldViewport.HeightTiles * tileSize
	for x := 0; x <= WorldViewport.WidthTiles; x++ {
		screenX := float32(x * tileSize)
		vector.StrokeLine(screen, screenX, 0, screenX, float32(hPx), 1, gridColor, false)
	}
	for y := 0; y <= WorldViewport.HeightTiles; y++ {
		screenY := float32(y * tileSize)
		vector.StrokeLine(screen, 0, screenY, float32(wPx), screenY, 1, gridColor, false)
	}
}

func (g *Game) startBattle(enemyID world.EntityID) {
	enc, ok := battlepkg.BuildEncounterFromWorld(g.world, enemyID)
	if !ok {
		return
	}
	g.mode = ModeBattle
	g.battle = battlepkg.BuildBattleContextFromEncounter(enc)
}

func (g *Game) endBattle() {
	g.mode = ModeExplore
	g.battle = nil
}

func (g *Game) updateCamera() {
	g.cameraX = g.player.GridX - WorldViewport.WidthTiles/2
	g.cameraY = g.player.GridY - WorldViewport.HeightTiles/2
}
