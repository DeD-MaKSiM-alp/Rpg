package battle

import (
	"testing"

	"mygame/world/entity"
)

func TestHitTestUnitUnderCursor_missesFarOutside(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	ctx := BuildBattleContextFromEncounter(enc, nil, 0)
	if ctx == nil {
		t.Fatal("nil ctx")
	}
	if got := HitTestUnitUnderCursor(ctx, 800, 600, -1000, -1000); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestHitTestUnitUnderCursor_hitsUnitRectCenter(t *testing.T) {
	enc := Encounter{Enemies: []EncounterEnemy{{EnemyID: 1, Kind: entity.EnemyKindSlime}}}
	ctx := BuildBattleContextFromEncounter(enc, nil, 0)
	if ctx == nil {
		t.Fatal("nil ctx")
	}
	layout := ctx.ComputeBattleHUDLayout(800, 600)
	if len(layout.UnitRects) == 0 {
		t.Fatal("no unit rects")
	}
	for id, hr := range layout.UnitRects {
		cx := int(hr.X + hr.W/2)
		cy := int(hr.Y + hr.H/2)
		got := HitTestUnitUnderCursor(ctx, 800, 600, cx, cy)
		if got != id {
			t.Fatalf("center of unit %d: want %d, got %d", id, id, got)
		}
		return
	}
}
