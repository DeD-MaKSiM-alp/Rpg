package world

import (
	"testing"

	"mygame/world/entity"
	"mygame/world/mapdata"
)

func TestInteractPickupRuinsReturnsChoiceWithoutCollecting(t *testing.T) {
	w := NewWorld(0)
	ch := w.getOrCreateChunk(mapdata.ChunkCoord{X: 0, Y: 0})
	ch.Pickups = append(ch.Pickups, entity.Pickup{X: 1, Y: 1, Kind: entity.PickupKindPOIRuins})
	if r := w.InteractPickupAfterMove(1, 1); r != PickupInteractPOIRequiresChoice {
		t.Fatalf("got %v, want PickupInteractPOIRequiresChoice", r)
	}
	if _, ok := w.PickupPreviewAt(1, 1); !ok {
		t.Fatal("expected pickup still present")
	}
	if !w.MarkPOIPickupCollected(1, 1) {
		t.Fatal("MarkPOIPickupCollected failed")
	}
	if _, ok := w.PickupPreviewAt(1, 1); ok {
		t.Fatal("expected pickup consumed")
	}
}

func TestInteractPickupAltarReturnsChoiceWithoutCollecting(t *testing.T) {
	w := NewWorld(0)
	ch := w.getOrCreateChunk(mapdata.ChunkCoord{X: 0, Y: 0})
	ch.Pickups = append(ch.Pickups, entity.Pickup{X: 2, Y: 2, Kind: entity.PickupKindPOIAltar})
	if r := w.InteractPickupAfterMove(2, 2); r != PickupInteractPOIRequiresChoice {
		t.Fatalf("got %v, want PickupInteractPOIRequiresChoice", r)
	}
	if _, ok := w.PickupPreviewAt(2, 2); !ok {
		t.Fatal("expected pickup still present")
	}
}
