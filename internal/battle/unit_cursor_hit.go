package battle

import "sort"

// HitTestUnitUnderCursor возвращает UnitID юнита под курсором (ростер или токен поля), иначе 0.
// Использует тот же layout, что и отрисовка HUD.
func HitTestUnitUnderCursor(b *BattleContext, screenW, screenH int, mx, my int) UnitID {
	if b == nil || len(b.Units) == 0 {
		return 0
	}
	layout := b.ComputeBattleHUDLayout(screenW, screenH)
	mxf := float32(mx)
	myf := float32(my)
	var ids []UnitID
	for id := range b.Units {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		if layout.pointHitsUnit(id, mxf, myf) {
			return id
		}
	}
	return 0
}
