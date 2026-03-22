package world

import (
	"mygame/world/entity"
	"mygame/world/generation"
	"mygame/world/mapdata"
)

// ZoneKind — грубый тип области (чанк); детерминирован от (chunkX, chunkY, worldSeed).
type ZoneKind int

const (
	// ZoneFrontier — только чанк (0,0): старт, без врагов/ресурсных пикапов по текущим правилам.
	ZoneFrontier ZoneKind = iota
	ZoneWild              // реже опасность и добыча
	ZoneAncient           // чаще руины/алтарь, выше шанс врага
	ZoneProspect          // выше шанс ресурсов
	ZoneDanger            // чаще враги, тяжелее типы
)

const zoneHashSalt = 0x5A7E

func maxAbsChunk(a, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	if a > b {
		return a
	}
	return b
}

// ZoneKindForChunk возвращает тип зоны для чанка. Одинаковые координаты и seed дают один и тот же результат.
func ZoneKindForChunk(chunkX, chunkY, worldSeed int) ZoneKind {
	d := maxAbsChunk(chunkX, chunkY)
	if d == 0 {
		return ZoneFrontier
	}
	h := generation.Hash2D(chunkX, chunkY, worldSeed+zoneHashSalt) % 4
	switch h {
	case 0:
		return ZoneWild
	case 1:
		return ZoneAncient
	case 2:
		return ZoneProspect
	default:
		return ZoneDanger
	}
}

// ZoneKindAtWorld — зона по мировой клетке (одна на весь чанк).
func (w *World) ZoneKindAtWorld(worldX, worldY int) ZoneKind {
	coord, _, _ := mapdata.WorldToChunkLocal(worldX, worldY)
	return ZoneKindForChunk(coord.X, coord.Y, w.seed)
}

// ZoneTitleRU — короткое название для HUD.
func (z ZoneKind) ZoneTitleRU() string {
	switch z {
	case ZoneFrontier:
		return "Рубеж"
	case ZoneWild:
		return "Дикие земли"
	case ZoneAncient:
		return "Древние руины"
	case ZoneProspect:
		return "Рудные земли"
	case ZoneDanger:
		return "Опасные земли"
	default:
		return "Неизвестно"
	}
}

// ZoneHUDLine — строка для нижней панели explore.
func (w *World) ZoneHUDLine(worldX, worldY int) string {
	z := w.ZoneKindAtWorld(worldX, worldY)
	return "Район: «" + z.ZoneTitleRU() + "»"
}

// zoneSpawnRules — пороги в процентах: событие при hash%100 < threshold.
type zoneSpawnRules struct {
	ResThreshold   int // шанс ресурсного пикапа
	POIThreshold   int // шанс попытки POI в чанке
	POI2Chance     int // второй POI в чанке
	EnemyThreshold int
}

func spawnRulesForZone(z ZoneKind) zoneSpawnRules {
	switch z {
	case ZoneFrontier:
		return zoneSpawnRules{ResThreshold: 0, POIThreshold: 35, POI2Chance: 18, EnemyThreshold: 0}
	case ZoneWild:
		return zoneSpawnRules{ResThreshold: 18, POIThreshold: 42, POI2Chance: 22, EnemyThreshold: 14}
	case ZoneAncient:
		return zoneSpawnRules{ResThreshold: 22, POIThreshold: 52, POI2Chance: 35, EnemyThreshold: 28}
	case ZoneProspect:
		return zoneSpawnRules{ResThreshold: 38, POIThreshold: 44, POI2Chance: 28, EnemyThreshold: 12}
	case ZoneDanger:
		return zoneSpawnRules{ResThreshold: 16, POIThreshold: 40, POI2Chance: 25, EnemyThreshold: 38}
	default:
		return zoneSpawnRules{ResThreshold: 28, POIThreshold: 48, POI2Chance: 32, EnemyThreshold: 22}
	}
}

// pickEnemyKindForZone — тип врага в зависимости от зоны (детерминированно).
func pickEnemyKindForZone(z ZoneKind, chunkX, chunkY, seed, attempt int) entity.EnemyKind {
	h := generation.Hash2D(chunkX, chunkY, seed+12000+attempt*11) % 100
	switch z {
	case ZoneDanger:
		if h < 35 {
			return entity.EnemyKindBandit
		}
		return entity.EnemyKindWolf
	case ZoneAncient:
		if h < 40 {
			return entity.EnemyKindWolf
		}
		return entity.EnemyKindSlime
	case ZoneProspect:
		if h < 15 {
			return entity.EnemyKindWolf
		}
		return entity.EnemyKindSlime
	default:
		return entity.EnemyKindSlime
	}
}
