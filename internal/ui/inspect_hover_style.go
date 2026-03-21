package ui

import battlepkg "mygame/internal/battle"

// InspectBattleHighlightPlan — что рисовать в бою: объединённый слой, только active, только hover.
type InspectBattleHighlightPlan struct {
	CombinedUnitID battlepkg.UnitID // 0 = нет объединённого прохода (hover+open на одном юните)
	ActiveUnitID   battlepkg.UnitID // 0 = не рисовать persistent active-open
	HoverUnitID    battlepkg.UnitID // 0 = не рисовать hover-слой
	HoverStrength  float32
}

// BuildInspectBattleHighlightPlan раскладывает три случая: только hover, только active-open, оба на одном юните, active+hover на разных.
func BuildInspectBattleHighlightPlan(hoverID, openID battlepkg.UnitID, inspectOpen bool) InspectBattleHighlightPlan {
	if !inspectOpen || openID == 0 {
		st := float32(1.0)
		if hoverID != 0 {
			st = InspectHoverStrength(false, 0, hoverID)
		}
		return InspectBattleHighlightPlan{HoverUnitID: hoverID, HoverStrength: st}
	}
	if hoverID != 0 && hoverID == openID {
		return InspectBattleHighlightPlan{CombinedUnitID: openID}
	}
	hs := float32(0)
	if hoverID != 0 {
		hs = InspectHoverStrength(true, openID, hoverID)
	}
	return InspectBattleHighlightPlan{
		ActiveUnitID:  openID,
		HoverUnitID:   hoverID,
		HoverStrength: hs,
	}
}

// InspectHoverStrength — множитель насыщенности подсветки при наведении на другого юнита,
// пока открыта карточка inspect по другому (мягче, чтобы не спорить с «открытым»).
func InspectHoverStrength(inspectOpen bool, openInspectID, hoverID battlepkg.UnitID) float32 {
	if hoverID == 0 {
		return 0
	}
	if !inspectOpen || openInspectID == 0 || hoverID == openInspectID {
		return 1.0
	}
	return 0.52
}

// FormationInspectHighlightPlan — подсветка строки состава (индексы глобальные; -1 = нет).
type FormationInspectHighlightPlan struct {
	CombinedGlobalIdx int
	ActiveGlobalIdx   int
	HoverGlobalIdx    int
	HoverStrength     float32
}

// BuildFormationInspectHighlightPlan — та же логика слоёв, что и в бою, для строки героя.
func BuildFormationInspectHighlightPlan(hoverGlobalIdx, openRowGlobalIdx int, inspectOpen bool) FormationInspectHighlightPlan {
	none := -1
	if !inspectOpen || openRowGlobalIdx < 0 {
		st := float32(0)
		if hoverGlobalIdx >= 0 {
			st = 1.0
		}
		return FormationInspectHighlightPlan{
			CombinedGlobalIdx: none,
			ActiveGlobalIdx:   none,
			HoverGlobalIdx:    hoverGlobalIdx,
			HoverStrength:     st,
		}
	}
	if hoverGlobalIdx >= 0 && hoverGlobalIdx == openRowGlobalIdx {
		return FormationInspectHighlightPlan{
			CombinedGlobalIdx: openRowGlobalIdx,
			ActiveGlobalIdx:   none,
			HoverGlobalIdx:    none,
			HoverStrength:     0,
		}
	}
	hs := float32(0)
	if hoverGlobalIdx >= 0 {
		hs = FormationInspectHoverStrength(inspectOpen, openRowGlobalIdx, hoverGlobalIdx)
	}
	return FormationInspectHighlightPlan{
		CombinedGlobalIdx: none,
		ActiveGlobalIdx:   openRowGlobalIdx,
		HoverGlobalIdx:    hoverGlobalIdx,
		HoverStrength:     hs,
	}
}

// FormationInspectHoverStrength — как InspectHoverStrength, но для глобального индекса строки (-1 = нет наведения).
func FormationInspectHoverStrength(inspectOpen bool, openRowGlobalIdx, hoverGlobalIdx int) float32 {
	if hoverGlobalIdx < 0 {
		return 0
	}
	if !inspectOpen || openRowGlobalIdx < 0 || hoverGlobalIdx == openRowGlobalIdx {
		return 1.0
	}
	return 0.52
}
