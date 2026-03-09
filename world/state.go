package world

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"mygame/world/entity"
	"mygame/world/generation"
	"mygame/world/mapdata"
	"mygame/world/render"
)

// Реэкспорт типов и констант для внешних пакетов (internal/game, internal/player, internal/battle).
type (
	Entity        = entity.Entity
	EntityID      = entity.EntityID
	EntityType    = entity.EntityType
	EntitySpawnKey = entity.EntitySpawnKey
	ChunkCoord    = mapdata.ChunkCoord
)

const EntityEnemy = entity.EntityEnemy

// World хранит бесконечный мир: чанки, пикапы, сущности.
type World struct {
	seed                int
	chunks              map[mapdata.ChunkCoord]*mapdata.Chunk
	collectedPickups    map[mapdata.PickupKey]bool
	entities            map[entity.EntityID]*entity.Entity
	nextEntityID        entity.EntityID
	defeatedEnemySpawns map[entity.EntitySpawnKey]bool
}

// NewWorld создаёт новый мир с заданным seed.
func NewWorld(seed int) *World {
	return &World{
		seed:                seed,
		chunks:              make(map[mapdata.ChunkCoord]*mapdata.Chunk),
		collectedPickups:    make(map[mapdata.PickupKey]bool),
		entities:            make(map[entity.EntityID]*entity.Entity),
		defeatedEnemySpawns: make(map[entity.EntitySpawnKey]bool),
	}
}

// Seed возвращает зерно мира.
func (w *World) Seed() int {
	return w.seed
}

// ——— DrawSource для render ———

func (w *World) GetTileAt(worldX, worldY int) mapdata.TileType {
	coord, localX, localY := mapdata.WorldToChunkLocal(worldX, worldY)
	chunk := w.getOrCreateChunk(coord)
	return chunk.Tiles[localY][localX]
}

func (w *World) Chunks() map[mapdata.ChunkCoord]*mapdata.Chunk {
	return w.chunks
}

func (w *World) Entities() map[entity.EntityID]*entity.Entity {
	return w.entities
}

// ——— Чанки ———

func (w *World) newChunk(chunkX, chunkY, seed int) *mapdata.Chunk {
	chunk := &mapdata.Chunk{
		ChunkX: chunkX,
		ChunkY: chunkY,
		Tiles:  make([][]mapdata.TileType, mapdata.ChunkSize),
	}
	for y := 0; y < mapdata.ChunkSize; y++ {
		chunk.Tiles[y] = make([]mapdata.TileType, mapdata.ChunkSize)
	}
	for localY := 0; localY < mapdata.ChunkSize; localY++ {
		for localX := 0; localX < mapdata.ChunkSize; localX++ {
			worldX := chunkX*mapdata.ChunkSize + localX
			worldY := chunkY*mapdata.ChunkSize + localY
			chunk.Tiles[localY][localX] = generation.GenerateTile(worldX, worldY, seed)
		}
	}
	chunk.Pickups = w.generatePickupsForChunk(chunkX, chunkY, seed, chunk.Tiles)
	w.generateEnemiesForChunk(chunkX, chunkY, seed, chunk.Tiles)
	return chunk
}

func (w *World) getOrCreateChunk(coord mapdata.ChunkCoord) *mapdata.Chunk {
	if c, ok := w.chunks[coord]; ok {
		return c
	}
	chunk := w.newChunk(coord.X, coord.Y, w.seed)
	w.chunks[coord] = chunk
	return chunk
}

// PreloadChunksAround заранее создаёт чанки вокруг клетки (radius в чанках).
func (w *World) PreloadChunksAround(worldX, worldY, radius int) {
	center, _, _ := mapdata.WorldToChunkLocal(worldX, worldY)
	for chunkY := center.Y - radius; chunkY <= center.Y+radius; chunkY++ {
		for chunkX := center.X - radius; chunkX <= center.X+radius; chunkX++ {
			w.getOrCreateChunk(mapdata.ChunkCoord{X: chunkX, Y: chunkY})
		}
	}
}

// UnloadChunksFarFrom удаляет чанки дальше radius от клетки.
func (w *World) UnloadChunksFarFrom(worldX, worldY, radius int) {
	center, _, _ := mapdata.WorldToChunkLocal(worldX, worldY)
	for coord := range w.chunks {
		dx := coord.X - center.X
		if dx < 0 {
			dx = -dx
		}
		dy := coord.Y - center.Y
		if dy < 0 {
			dy = -dy
		}
		if dx > radius || dy > radius {
			delete(w.chunks, coord)
		}
	}
}

// ChunkCoordAt возвращает координаты чанка для мировой клетки.
func (w *World) ChunkCoordAt(worldX, worldY int) ChunkCoord {
	coord, _, _ := mapdata.WorldToChunkLocal(worldX, worldY)
	return ChunkCoord(coord)
}

// LoadedChunkCount возвращает число загруженных чанков.
func (w *World) LoadedChunkCount() int {
	return len(w.chunks)
}

// ——— Проходимость и пикапы ———

// IsWalkable возвращает, можно ли пройти по клетке.
func (w *World) IsWalkable(x, y int) bool {
	coord, localX, localY := mapdata.WorldToChunkLocal(x, y)
	chunk := w.getOrCreateChunk(coord)
	return generation.IsTileWalkable(chunk.Tiles[localY][localX])
}

// CollectPickupAt собирает пикап в клетке; возвращает true, если что-то собрано.
func (w *World) CollectPickupAt(worldX, worldY int) bool {
	coord, _, _ := mapdata.WorldToChunkLocal(worldX, worldY)
	chunk := w.getOrCreateChunk(coord)
	key := mapdata.PickupKey{X: worldX, Y: worldY}
	for i := range chunk.Pickups {
		p := &chunk.Pickups[i]
		if p.Collected {
			continue
		}
		if p.X == worldX && p.Y == worldY {
			p.Collected = true
			w.collectedPickups[key] = true
			return true
		}
	}
	return false
}

func (w *World) generatePickupsForChunk(chunkX, chunkY, seed int, tiles [][]mapdata.TileType) []entity.Pickup {
	if chunkX == 0 && chunkY == 0 {
		return nil
	}
	if generation.Hash2D(chunkX, chunkY, seed+5000)%100 >= 28 {
		return nil
	}
	for attempt := 0; attempt < 8; attempt++ {
		localX := mapdata.PositiveMod(generation.Hash2D(chunkX, chunkY, seed+6000+attempt*17), mapdata.ChunkSize)
		localY := mapdata.PositiveMod(generation.Hash2D(chunkY, chunkX, seed+7000+attempt*23), mapdata.ChunkSize)
		tile := tiles[localY][localX]
		if !generation.IsTileWalkable(tile) {
			continue
		}
		worldX := chunkX*mapdata.ChunkSize + localX
		worldY := chunkY*mapdata.ChunkSize + localY
		if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
			continue
		}
		key := mapdata.PickupKey{X: worldX, Y: worldY}
		if w.collectedPickups[key] {
			return nil
		}
		return []entity.Pickup{{X: worldX, Y: worldY, Collected: false}}
	}
	return nil
}

// ——— Враги и сущности ———

// GetEnemyAt возвращает врага в клетке или nil.
func (w *World) GetEnemyAt(worldX, worldY int) *Entity {
	return entity.GetEnemyAt(w.entities, worldX, worldY)
}

func (w *World) isEnemyBlockingTile(worldX, worldY int, ignoreID entity.EntityID) bool {
	return entity.IsEnemyBlockingTile(w.entities, worldX, worldY, ignoreID)
}

// RemoveEnemy убирает врага после боя.
func (w *World) RemoveEnemy(id EntityID) {
	entity.RemoveEnemy(w.entities, w.defeatedEnemySpawns, id)
}

func (w *World) addEntity(entityType entity.EntityType, worldX, worldY int) *entity.Entity {
	return entity.AddEntity(w.entities, &w.nextEntityID, entityType, worldX, worldY)
}

func (w *World) generateEnemiesForChunk(chunkX, chunkY, seed int, tiles [][]mapdata.TileType) {
	if chunkX == 0 && chunkY == 0 {
		return
	}
	if generation.Hash2D(chunkX, chunkY, seed+9000)%100 >= 22 {
		return
	}
	for attempt := 0; attempt < 8; attempt++ {
		localX := mapdata.PositiveMod(generation.Hash2D(chunkX, chunkY, seed+10000+attempt*19), mapdata.ChunkSize)
		localY := mapdata.PositiveMod(generation.Hash2D(chunkY, chunkX, seed+11000+attempt*29), mapdata.ChunkSize)
		tile := tiles[localY][localX]
		if !generation.IsTileWalkable(tile) {
			continue
		}
		worldX := chunkX*mapdata.ChunkSize + localX
		worldY := chunkY*mapdata.ChunkSize + localY
		if worldX >= 0 && worldX <= 6 && worldY >= 0 && worldY <= 6 {
			continue
		}
		spawnKey := entity.SpawnKey(worldX, worldY)
		if w.defeatedEnemySpawns[spawnKey] {
			return
		}
		if entity.GetEnemyAt(w.entities, worldX, worldY) != nil {
			return
		}
		w.addEntity(entity.EntityEnemy, worldX, worldY)
		return
	}
}

// AdvanceTurn выполняет ход врагов в 3 фазы: collect intents → resolve → apply.
// Возвращает ID врага и true, если есть валидный attack intent по игроку (старт боя).
func (w *World) AdvanceTurn(playerX, playerY int) (EntityID, bool) {
	intents := w.collectEnemyIntents(playerX, playerY)
	attackID, validMoves := w.resolveEnemyIntents(intents, playerX, playerY)
	if attackID != 0 {
		return EntityID(attackID), true
	}
	w.applyMoveIntents(validMoves)
	return 0, false
}

// collectEnemyIntents возвращает срез намерений всех живых врагов в стабильном порядке (по EntityID).
func (w *World) collectEnemyIntents(playerX, playerY int) []entity.Intent {
	ids := w.aliveEnemyIDsStable()
	out := make([]entity.Intent, 0, len(ids))
	for _, id := range ids {
		e := w.entities[id]
		if e == nil {
			continue
		}
		out = append(out, entity.BuildEnemyIntent(e, playerX, playerY))
	}
	return out
}

func (w *World) aliveEnemyIDsStable() []entity.EntityID {
	ids := make([]entity.EntityID, 0, len(w.entities))
	for id, e := range w.entities {
		if !e.Alive || e.Type != entity.EntityEnemy {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// resolveEnemyIntents возвращает enemyID для старта боя (первый валидный attack в стабильном порядке)
// и срез валидных move intents для применения.
func (w *World) resolveEnemyIntents(intents []entity.Intent, playerX, playerY int) (entity.EntityID, []entity.Intent) {
	moveTargetCount := make(map[struct{ x, y int }]int)
	for _, in := range intents {
		if in.Type != entity.IntentMove {
			continue
		}
		key := struct{ x, y int }{in.TargetX, in.TargetY}
		moveTargetCount[key]++
	}
	conflictCells := make(map[struct{ x, y int }]bool)
	for k, n := range moveTargetCount {
		if n > 1 {
			conflictCells[k] = true
		}
	}

	var attackID entity.EntityID
	var validMoves []entity.Intent
	for _, in := range intents {
		switch in.Type {
		case entity.IntentAttack:
			if attackID != 0 {
				continue
			}
			e := w.entities[in.EntityID]
			if e == nil || !e.Alive {
				continue
			}
			if in.TargetX != playerX || in.TargetY != playerY {
				continue
			}
			if !entity.IsAdjacent8(e.X, e.Y, playerX, playerY) {
				continue
			}
			attackID = in.EntityID
		case entity.IntentMove:
			if in.TargetX == playerX && in.TargetY == playerY {
				continue
			}
			if conflictCells[struct{ x, y int }{in.TargetX, in.TargetY}] {
				continue
			}
			if !w.IsWalkable(in.TargetX, in.TargetY) {
				continue
			}
			if w.isEnemyBlockingTile(in.TargetX, in.TargetY, in.EntityID) {
				continue
			}
			validMoves = append(validMoves, in)
		}
	}
	return attackID, validMoves
}

func (w *World) applyMoveIntents(moves []entity.Intent) {
	for _, in := range moves {
		e := w.entities[in.EntityID]
		if e == nil || !e.Alive {
			continue
		}
		if w.isEnemyBlockingTile(in.TargetX, in.TargetY, in.EntityID) {
			continue
		}
		if !w.IsWalkable(in.TargetX, in.TargetY) {
			continue
		}
		e.X = in.TargetX
		e.Y = in.TargetY
	}
}

// ——— Рендер ———

// Draw рисует видимую часть мира.
func (w *World) Draw(screen *ebiten.Image, cameraX, cameraY, visibleTilesX, visibleTilesY, tileSize int) {
	render.Draw(screen, w, cameraX, cameraY, visibleTilesX, visibleTilesY, tileSize)
}

// DrawChunkDebugOverlay рисует отладочную сетку чанков.
func (w *World) DrawChunkDebugOverlay(
	screen *ebiten.Image,
	cameraX, cameraY int,
	visibleTilesX, visibleTilesY int,
	tileSize int,
	screenWidth, screenHeight int,
) {
	render.DrawChunkDebugOverlay(screen, cameraX, cameraY, visibleTilesX, visibleTilesY, tileSize, screenWidth, screenHeight)
}
