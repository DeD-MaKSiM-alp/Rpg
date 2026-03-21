package world

import (
	"fmt"
	"testing"

	"mygame/world/entity"
)

// Пресет 800×600 и tileSize=48 как в internal/game: 16×12 тайлов; старт (2,2) как в main.
func TestRecruitCampSingleInStartChunkVisibleFromDefaultCamera(t *testing.T) {
	const (
		playerX, playerY = 2, 2
		widthTiles       = 16
		heightTiles      = 12
	)
	cameraX := playerX - widthTiles/2
	cameraY := playerY - heightTiles/2
	endX := cameraX + widthTiles
	endY := cameraY + heightTiles

	for _, seed := range []int{0, 1, 3, 42, 12345, -7} {
		seed := seed
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			t.Parallel()
			w := NewWorld(seed)
			w.PreloadChunksAround(playerX, playerY, 1)
			if n := w.ActiveRecruitCampCount(); n != 1 {
				t.Fatalf("ActiveRecruitCampCount=%d, want 1", n)
			}
			var campX, campY int
			found := false
			for coord, ch := range w.chunks {
				if coord.X != 0 || coord.Y != 0 {
					continue
				}
				for i := range ch.Pickups {
					p := &ch.Pickups[i]
					if p.Kind != entity.PickupKindRecruitCamp || p.Collected {
						continue
					}
					if found {
						t.Fatalf("second recruit camp at %d,%d", p.X, p.Y)
					}
					found = true
					campX, campY = p.X, p.Y
				}
			}
			if !found {
				t.Fatal("no recruit camp in chunk (0,0)")
			}
			if campX < cameraX || campX >= endX || campY < cameraY || campY >= endY {
				t.Fatalf("camp at %d,%d outside default viewport [%d,%d)x[%d,%d)", campX, campY, cameraX, endX, cameraY, endY)
			}
		})
	}
}

func TestActiveRecruitCampCountZeroWithoutPreload(t *testing.T) {
	w := NewWorld(3)
	if w.ActiveRecruitCampCount() != 0 {
		t.Fatalf("got %d, want 0 before preload", w.ActiveRecruitCampCount())
	}
}

func TestRecruitCampVisibleSeedsWideRange(t *testing.T) {
	const (
		playerX, playerY = 2, 2
		widthTiles       = 16
		heightTiles      = 12
	)
	cameraX := playerX - widthTiles/2
	cameraY := playerY - heightTiles/2
	endX := cameraX + widthTiles
	endY := cameraY + heightTiles

	for seed := 0; seed < 300; seed++ {
		w := NewWorld(seed)
		w.PreloadChunksAround(playerX, playerY, 1)
		if w.ActiveRecruitCampCount() != 1 {
			t.Fatalf("seed %d: ActiveRecruitCampCount=%d", seed, w.ActiveRecruitCampCount())
		}
		var campX, campY int
		for coord, ch := range w.chunks {
			if coord.X != 0 || coord.Y != 0 {
				continue
			}
			for i := range ch.Pickups {
				p := &ch.Pickups[i]
				if p.Kind == entity.PickupKindRecruitCamp && !p.Collected {
					campX, campY = p.X, p.Y
					goto found
				}
			}
		}
		t.Fatalf("seed %d: no camp", seed)
	found:
		if campX < cameraX || campX >= endX || campY < cameraY || campY >= endY {
			t.Fatalf("seed %d: camp %d,%d outside viewport", seed, campX, campY)
		}
	}
}
