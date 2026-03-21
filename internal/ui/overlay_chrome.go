// overlay_chrome.go — единый визуальный язык модальных оверлеев (postbattle, recruit, formation, explore bar),
// согласованный с inspect-card / battle scene (тёмная подложка, рамки, AccentStrip).

package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawUnifiedModalPanelChrome — фон и двойной контур карточки; левая акцентная полоса как у inspect.
func drawUnifiedModalPanelChrome(screen *ebiten.Image, x, y, w, h float32) {
	if w <= 0 || h <= 0 {
		return
	}
	vector.FillRect(screen, x, y, w, h, Theme.PostBattlePanelBG, false)
	vector.FillRect(screen, x, y, 4, h, Theme.AccentStrip, false)
	vector.StrokeRect(screen, x, y, w, h, 2, Theme.PostBattleBorder, false)
	if w > 12 && h > 12 {
		vector.StrokeRect(screen, x+5, y+4, w-10, h-8, 1, Theme.RosterCardInnerStroke, false)
	}
}

// drawUnifiedBottomBarChrome — нижняя полоса подсказок explore: та же тональность, что и модальные панели.
func drawUnifiedBottomBarChrome(screen *ebiten.Image, x, y, w, h float32) {
	if w <= 0 || h <= 0 {
		return
	}
	vector.FillRect(screen, x, y, w, h, Theme.PanelBGDeep, false)
	vector.FillRect(screen, x, y, 4, h, Theme.AccentStrip, false)
	vector.StrokeRect(screen, x, y, w, h, 1, Theme.PostBattleBorder, false)
	DrawThinAccentLine(screen, x+6, y+4, w-12)
}
