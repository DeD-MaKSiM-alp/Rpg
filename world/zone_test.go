package world

import (
	"testing"

	"mygame/world/entity"
)

func TestZoneKindForChunk_deterministic(t *testing.T) {
	a := ZoneKindForChunk(2, -3, 42)
	b := ZoneKindForChunk(2, -3, 42)
	if a != b {
		t.Fatalf("same coords+seed: %v vs %v", a, b)
	}
}

func TestZoneKindForChunk_chunk00_isFrontier(t *testing.T) {
	if ZoneKindForChunk(0, 0, 999) != ZoneFrontier {
		t.Fatal("want frontier at 0,0")
	}
}

func TestZoneKindForChunk_neighbor_notAlwaysFrontier(t *testing.T) {
	z := ZoneKindForChunk(1, 0, 7)
	if z == ZoneFrontier {
		t.Fatal("chunk (1,0) should not be frontier")
	}
}

func TestWorld_ZoneKindAtWorld_matchesChunk(t *testing.T) {
	w := NewWorld(123)
	x, y := 3, 3
	coord := w.ChunkCoordAt(x, y)
	want := ZoneKindForChunk(coord.X, coord.Y, w.seed)
	if g := w.ZoneKindAtWorld(x, y); g != want {
		t.Fatalf("world cell: got %v want %v", g, want)
	}
}

func TestZoneSpawnRules_monotonicResourceVsProspect(t *testing.T) {
	wild := spawnRulesForZone(ZoneWild)
	pro := spawnRulesForZone(ZoneProspect)
	if pro.ResThreshold <= wild.ResThreshold {
		t.Fatalf("prospect res %d should exceed wild %d", pro.ResThreshold, wild.ResThreshold)
	}
}

func TestPickEnemyKindForZone_danger_notAlwaysSlime(t *testing.T) {
	sawNonSlime := false
	for attempt := 0; attempt < 200; attempt++ {
		k := pickEnemyKindForZone(ZoneDanger, 5, 5, 99, attempt)
		if k != entity.EnemyKindSlime {
			sawNonSlime = true
			break
		}
	}
	if !sawNonSlime {
		t.Fatal("expected non-slime enemy in danger zone over attempts")
	}
}
