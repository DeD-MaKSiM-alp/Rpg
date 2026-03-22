package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	inputpkg "mygame/internal/input"
	"mygame/internal/party"
	playerpkg "mygame/internal/player"
	"mygame/internal/postbattle"
	"mygame/internal/ui"
	"mygame/world"
	"mygame/world/entity"
)

// DefaultWindowTitle — заголовок окна по умолчанию.
const DefaultWindowTitle = "My Game"

// --- Разрешение и viewport: одна точка правды ---
//
// Активное разрешение задаётся только через applyResolutionPreset(), которая вызывается из Run().
// ScreenWidth/ScreenHeight — текущее logical/game resolution; задаются только из пресета.
// Layout() возвращает именно их. Никаких ручных присваиваний ScreenWidth/ScreenHeight вне applyResolutionPreset().
//
// Viewport мира отделён от размера окна: WorldViewport (WidthTiles, HeightTiles) задаёт видимую область мира в тайлах.
// Камера, drawGrid и world.Draw опираются на WorldViewport; размер окна нужен только для Layout и UI (HUD).
// Так проще тестировать HUD на разных разрешениях и в будущем добавить resize/scale.

// ResolutionPreset задаёт один пресет разрешения окна.
type ResolutionPreset struct {
	Width  int
	Height int
}

// ResolutionPresets — список пресетов. Активный выбирается через ActivePresetIndex.
var ResolutionPresets = []ResolutionPreset{
	{800, 600},
	{1024, 768},
	{1280, 720},
	{1366, 768},
	{1600, 900},
	{1920, 1080},
	{2560, 1440},
}

// ActivePresetIndex — индекс активного пресета разрешения. Меняется в runtime по F6/F7.
// 0=800x600, 1=1024x768, 2=1280x720, 3=1366x768, 4=1600x900.
var ActivePresetIndex = 2

// Viewport задаёт логический размер видимой области мира в тайлах (не пиксели окна).
type Viewport struct {
	WidthTiles  int
	HeightTiles int
}

// WorldViewport — текущая видимая область мира. Задаётся только в applyResolutionPreset().
var WorldViewport Viewport

// ScreenWidth, ScreenHeight — текущее logical resolution. Задаются только в applyResolutionPreset().
// Используются: Layout(), ebiten.SetWindowSize, UI (HUD, battle overlay). World/camera используют WorldViewport.
var (
	ScreenWidth  int
	ScreenHeight int
)

// applyResolutionPreset — единственное место применения размеров окна и viewport.
// Берёт пресет по ActivePresetIndex, выставляет ScreenWidth/ScreenHeight, WorldViewport и ebiten.SetWindowSize.
// Можно вызывать в runtime (переключение по F6/F7).
func applyResolutionPreset() {
	idx := ActivePresetIndex
	if idx < 0 || idx >= len(ResolutionPresets) {
		idx = 0
		ActivePresetIndex = idx
	}
	preset := ResolutionPresets[idx]
	ScreenWidth = preset.Width
	ScreenHeight = preset.Height
	WorldViewport.WidthTiles = preset.Width / tileSize
	WorldViewport.HeightTiles = preset.Height / tileSize
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
	// ModeFormation — экран состава: Active (строй) + Reserve, порядок Active; PlayerCombatSeeds только из Active.
	ModeFormation
	// ModeRecruitOffer — подтверждение найма с лагеря на карте (world PickupKindRecruitCamp).
	ModeRecruitOffer
	// ModePOIChoice — выбор risk/reward у POI (руины / алтарь) до сбора пикапа.
	ModePOIChoice
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

	// Party — канонический отряд (ростер); лидер Active[0] получает награды, сиды для боя — PlayerCombatSeeds().
	party party.Party

	// BattlesWon — число выигранных боёв за сессию; используется для эскалации сложности врагов и генерации оффера наград.
	BattlesWon int

	// TrainingMarks — сессионный счётчик «знаков обучения»: начисляется за победы, тратится на promotion в лагере (см. promotion_gate.go).
	TrainingMarks int

	// Post-battle: экран результата и выбор награды (оркестрация вынесена в postbattle.Flow).
	postBattle postbattle.Flow

	// BattleHUDStyle: 0 = v1 table (fallback/debug), 1 = v2 Disciples-like. Used to set battle.LayoutStyle each frame.
	BattleHUDStyle int

	// Временный debug: последнее направление, возвращённое ReadExploreInput (только для отрисовки).
	debugInputDX, debugInputDY int

	// formationSel — индекс строки в overlay: [0, len(Active)) строй, [len(Active), ...) резерв.
	formationSel int
	// formationInspectOpen — карточка бойца (I) поверх состава.
	formationInspectOpen bool
	// formationPromoteBranchIdx — выбор ветки при двух UpgradeOptions: -1 = не выбрано, 0/1 = ветка.
	formationPromoteBranchIdx int

	// exploreRestMsg / exploreRestMsgTicks — краткая обратная связь после отдыха (R) в explore).
	exploreRestMsg      string
	exploreRestMsgTicks int

	// exploreRecruitMsg — баннер после попытки рекрута (F9 в explore); этапный acquisition без мира.
	exploreRecruitMsg      string
	exploreRecruitMsgTicks int

	// explorePOIMsg — краткий баннер после взаимодействия с POI в мире.
	explorePOIMsg      string
	explorePOIMsgTicks int

	// recruitOfferX/Y — клетка лагеря наёмников при ModeRecruitOffer (после подтверждения — MarkRecruitPickupCollected).
	recruitOfferX int
	recruitOfferY int

	// poiChoice* — клетка и вид POI при ModePOIChoice; poiChoiceSel 0/1 — выбранный вариант.
	poiChoiceX, poiChoiceY int
	poiChoiceKind          entity.PickupKind
	poiChoiceSel           int

	// formationMsg — баннер после promotion (P) на экране состава с открытой карточкой.
	formationMsg      string
	formationMsgTicks int

	// battleInspectOpen — карточка по ПКМ в бою (не пост-бой).
	battleInspectOpen   bool
	battleInspectUnitID battlepkg.UnitID

	// inspectHoverBattleUnitID — юнит под курсором для подсказки ПКМ-inspect (0 = нет).
	inspectHoverBattleUnitID battlepkg.UnitID
	// inspectHoverFormationGlobalIdx — строка состава под курсором (-1 = нет).
	inspectHoverFormationGlobalIdx int
}

// NewGame создаёт новый экземпляр игры (мир, игрок, UI-шрифт и т.д.).
func NewGame(worldSeed, playerGridX, playerGridY int) *Game {
	return &Game{
		player:                         *playerpkg.NewPlayer(playerGridX, playerGridY),
		world:                          world.NewWorld(worldSeed),
		input:                          inputpkg.New(),
		hudFace:                        ui.LoadHUDFace(),
		mode:                           ModeExplore,
		battle:                         nil,
		party:                          party.DefaultParty(),
		BattleHUDStyle:                 1, // 1 = v2 Disciples-like (default), 0 = v1 table (debug fallback)
		inspectHoverFormationGlobalIdx: -1,
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

func (g *Game) syncPromotionBranchForBattleInspect() {
	if g.battle == nil {
		g.formationPromoteBranchIdx = 0
		return
	}
	u := g.battle.Units[g.battleInspectUnitID]
	if u == nil || u.Origin.PartyActiveIndex < 0 {
		g.formationPromoteBranchIdx = 0
		return
	}
	h := g.party.HeroAtGlobalIndex(u.Origin.PartyActiveIndex)
	if h == nil {
		g.formationPromoteBranchIdx = 0
		return
	}
	targets, err := hero.PromotionTargetUnitIDs(h)
	if err != nil || len(targets) < 2 {
		g.formationPromoteBranchIdx = 0
	} else {
		g.formationPromoteBranchIdx = -1
	}
}

func (g *Game) syncPromotionBranchForInspectHero() {
	h := g.party.HeroAtGlobalIndex(g.formationSel)
	if h == nil {
		g.formationPromoteBranchIdx = 0
		return
	}
	targets, err := hero.PromotionTargetUnitIDs(h)
	if err != nil || len(targets) < 2 {
		g.formationPromoteBranchIdx = 0
	} else {
		g.formationPromoteBranchIdx = -1
	}
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
	g.postBattle.Reset()
	g.battleInspectOpen = false
	g.battleInspectUnitID = 0
	g.inspectHoverBattleUnitID = 0
	g.exploreRestMsg = ""
	g.exploreRestMsgTicks = 0
	seeds := g.party.PlayerCombatSeeds()
	if len(seeds) == 0 {
		g.mode = ModeExplore
		return
	}
	g.battle = battlepkg.BuildBattleContextFromEncounter(enc, seeds, g.BattlesWon)
}

func (g *Game) endBattle() {
	g.mode = ModeExplore
	g.battle = nil
	g.postBattle.Reset()
	g.battleInspectOpen = false
	g.battleInspectUnitID = 0
	g.inspectHoverBattleUnitID = 0
}

func (g *Game) updateCamera() {
	g.cameraX = g.player.GridX - WorldViewport.WidthTiles/2
	g.cameraY = g.player.GridY - WorldViewport.HeightTiles/2
}
