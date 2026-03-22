package ui

import (
	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/unitdata"
)

func heroResourceProfileInspectLine(h *hero.Hero) string {
	if h == nil {
		return ""
	}
	if tpl, ok := unitdata.GetUnitTemplate(h.UnitID); ok {
		return battlepkg.ResourceProfileInspectLineRU(tpl.Role)
	}
	return ""
}
