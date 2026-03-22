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

// drawUnifiedBottomBarChrome — нижний модуль explore: контекст и действия (отличается от верха/слева цветом и акцентом).
func drawUnifiedBottomBarChrome(screen *ebiten.Image, x, y, w, h float32) {
	if w <= 0 || h <= 0 {
		return
	}
	vector.FillRect(screen, x, y, w, h, Theme.ExploreContextBG, false)
	vector.FillRect(screen, x, y, 5, h, Theme.ExploreContextLeftStrip, false)
	vector.FillRect(screen, x+5, y, w-5, 2, Theme.PanelTitleSep, false)
	vector.FillRect(screen, x, y+h-1, w, 1, Theme.ExploreModuleEdge, false)
}

// drawPostBattleEventChrome — модалка результата/награды: одна читаемая карточка без двойного контура inspect.
func drawPostBattleEventChrome(screen *ebiten.Image, x, y, w, h float32) {
	if w <= 0 || h <= 0 {
		return
	}
	vector.FillRect(screen, x, y, w, h, Theme.PostBattleEventCardBG, false)
	vector.FillRect(screen, x, y, 5, h, Theme.AccentStrip, false)
	vector.StrokeRect(screen, x, y, w, h, 1, Theme.PostBattleBorder, false)
}
