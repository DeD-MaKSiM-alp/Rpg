package world

import (
	"mygame/world/entity"
	"mygame/world/generation"
	"mygame/world/mapdata"
)

// Порядок ротации типов POI при размещении (детерминированно от чанка).
var poiKindsOrder = [...]entity.PickupKind{
	entity.PickupKindPOIAltar,
	entity.PickupKindPOISpring,
	entity.PickupKindPOICache,
	entity.PickupKindPOIRuins,
	entity.PickupKindPOICampfire,
}

// generatePOIsForChunk — 0–2 одноразовых POI в чанке; не пересекаются с пикапами; учёт collectedPickups.
func (w *World) generatePOIsForChunk(chunkX, chunkY, seed int, tiles [][]mapdata.TileType, existing []entity.Pickup) []entity.Pickup {
	zone := ZoneKindForChunk(chunkX, chunkY, seed)
	prof := spawnRulesForZone(zone)
	// Не каждый чанк: иначе слишком плотная сетка.
	if generation.Hash2D(chunkX, chunkY, seed+9100)%100 >= prof.POIThreshold {
		return nil
	}
	n := 1
	if generation.Hash2D(chunkY, chunkX, seed+9200)%100 < prof.POI2Chance {
		n = 2
	}
	var out []entity.Pickup
	placed := 0
	for attempt := 0; attempt < 24 && placed < n; attempt++ {
		localX := mapdata.PositiveMod(generation.Hash2D(chunkX, chunkY, seed+9300+attempt*41), mapdata.ChunkSize)
		localY := mapdata.PositiveMod(generation.Hash2D(chunkY, chunkX, seed+9400+attempt*43), mapdata.ChunkSize)
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
			continue
		}
		if pickupCellOccupied(existing, worldX, worldY) {
			continue
		}
		kind := pickPOIKindForZone(zone, worldX, worldY, seed, attempt)
		out = append(out, entity.Pickup{
			X: worldX, Y: worldY, Collected: false, Kind: kind,
		})
		existing = append(existing, out[len(out)-1])
		placed++
	}
	return out
}

// pickPOIKindForZone — веса POI по зоне (детерминированно от клетки).
func pickPOIKindForZone(zone ZoneKind, worldX, worldY, seed, attempt int) entity.PickupKind {
	h := generation.Hash2D(worldX, worldY, seed+9500+attempt*41)
	switch zone {
	case ZoneAncient:
		switch h % 13 {
		case 0, 1, 2, 3, 4:
			return entity.PickupKindPOIRuins
		case 5, 6, 7:
			return entity.PickupKindPOIAltar
		case 8, 9:
			return entity.PickupKindPOISpring
		case 10, 11:
			return entity.PickupKindPOICampfire
		default:
			return entity.PickupKindPOICache
		}
	case ZoneProspect:
		switch h % 12 {
		case 0, 1, 2, 3:
			return entity.PickupKindPOICache
		case 4, 5, 6:
			return entity.PickupKindPOISpring
		case 7, 8:
			return entity.PickupKindPOICampfire
		case 9, 10:
			return entity.PickupKindPOIRuins
		default:
			return entity.PickupKindPOIAltar
		}
	case ZoneDanger:
		switch h % 10 {
		case 0, 1, 2, 3, 4:
			return entity.PickupKindPOIRuins
		case 5, 6, 7:
			return entity.PickupKindPOIAltar
		default:
			return entity.PickupKindPOICampfire
		}
	case ZoneWild:
		switch h % 12 {
		case 0, 1, 2:
			return entity.PickupKindPOISpring
		case 3, 4, 5:
			return entity.PickupKindPOICampfire
		case 6, 7, 8:
			return entity.PickupKindPOICache
		default:
			return entity.PickupKindPOIRuins
		}
	default:
		return poiKindsOrder[(generation.Hash2D(worldX, worldY, seed+9500)+attempt)%len(poiKindsOrder)]
	}
}

func poiResultFromKind(k entity.PickupKind) PickupInteractionResult {
	switch k {
	case entity.PickupKindPOIAltar:
		return PickupInteractPOIAltar
	case entity.PickupKindPOISpring:
		return PickupInteractPOISpring
	case entity.PickupKindPOICache:
		return PickupInteractPOICache
	case entity.PickupKindPOIRuins:
		return PickupInteractPOIRuins
	case entity.PickupKindPOICampfire:
		return PickupInteractPOICampfire
	default:
		return PickupInteractNone
	}
}
