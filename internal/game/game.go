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

// Экспортируемые размеры экрана для настройки окна из main.
const (
	ScreenWidth  = 800
	ScreenHeight = 600
)

const (
	tileSize             = 48
	visibleTilesX        = ScreenWidth / tileSize
	visibleTilesY        = ScreenHeight / tileSize
	debugShowChunkOverlay = false
	chunkPreloadRadius   = 1
	chunkUnloadRadius    = 2
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
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
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
	for x := 0; x <= ScreenWidth; x += tileSize {
		screenX := float32(x)
		vector.StrokeLine(screen, screenX, 0, screenX, float32(ScreenHeight), 1, gridColor, false)
	}
	if ScreenWidth%tileSize != 0 {
		vector.StrokeLine(screen, float32(ScreenWidth), 0, float32(ScreenWidth), float32(ScreenHeight), 1, gridColor, false)
	}
	for y := 0; y <= ScreenHeight; y += tileSize {
		screenY := float32(y)
		vector.StrokeLine(screen, 0, screenY, float32(ScreenWidth), screenY, 1, gridColor, false)
	}
	if ScreenHeight%tileSize != 0 {
		vector.StrokeLine(screen, 0, float32(ScreenHeight), float32(ScreenWidth), float32(ScreenHeight), 1, gridColor, false)
	}
}

func (g *Game) startBattle(enemyID world.EntityID) {
	g.mode = ModeBattle
	g.battle = battlepkg.NewBattleContext(enemyID)
}

func (g *Game) endBattle() {
	g.mode = ModeExplore
	g.battle = nil
}

func (g *Game) updateCamera() {
	g.cameraX = g.player.GridX - visibleTilesX/2
	g.cameraY = g.player.GridY - visibleTilesY/2
}
