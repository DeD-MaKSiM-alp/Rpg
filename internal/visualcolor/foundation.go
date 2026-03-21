// Package visualcolor — общие RGBA-токены для internal/ui и world/render без циклов импорта.
// Единственный канонический набор значений для полей, которые используются и в UI, и в мире.
package visualcolor

import "image/color"

// Foundation — shared foundation colors (UI Theme ∩ world render ∩ player cues).
// При смене палитры править здесь; ui.Theme для этих полей ссылается на те же значения.
var Foundation = struct {
	PanelBGDeep          color.RGBA
	SceneTint            color.RGBA
	PanelBorder          color.RGBA
	PanelTitleSep        color.RGBA
	BattlefieldTokenAlly color.RGBA
	ValidTarget          color.RGBA
	HoverTarget          color.RGBA
	AccentStrip          color.RGBA
	PostBattleBorder     color.RGBA
	AbilityHoverBG       color.RGBA
	HPEnemyFill          color.RGBA
	EnemyAccent          color.RGBA
	SelectedKill         color.RGBA
	TextPrimary          color.RGBA
	ActiveTurn           color.RGBA
}{
	PanelBGDeep:          color.RGBA{R: 18, G: 20, B: 26, A: 255},
	SceneTint:            color.RGBA{R: 20, G: 22, B: 30, A: 255},
	PanelBorder:          color.RGBA{R: 72, G: 78, B: 92, A: 255},
	PanelTitleSep:        color.RGBA{R: 48, G: 52, B: 62, A: 255},
	BattlefieldTokenAlly: color.RGBA{R: 55, G: 95, B: 75, A: 255},
	ValidTarget:          color.RGBA{R: 80, G: 145, B: 255, A: 255},
	HoverTarget:          color.RGBA{R: 120, G: 185, B: 255, A: 255},
	AccentStrip:          color.RGBA{R: 180, G: 145, B: 70, A: 255},
	PostBattleBorder:     color.RGBA{R: 95, G: 100, B: 125, A: 255},
	AbilityHoverBG:       color.RGBA{R: 42, G: 55, B: 72, A: 255},
	HPEnemyFill:          color.RGBA{R: 200, G: 95, B: 95, A: 255},
	EnemyAccent:          color.RGBA{R: 115, G: 75, B: 85, A: 255},
	SelectedKill:         color.RGBA{R: 235, G: 75, B: 75, A: 255},
	TextPrimary:          color.RGBA{R: 235, G: 236, B: 240, A: 255},
	ActiveTurn:           color.RGBA{R: 255, G: 210, B: 75, A: 255},
}
