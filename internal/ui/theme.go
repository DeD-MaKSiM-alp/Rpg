// Package ui — theme: единая палитра и микро-графика для battle / formation / explore.
// Цель: согласованность без финального арт-пайплайна; vector-only, без внешних ассетов.
package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mygame/internal/visualcolor"
)

// Theme — базовые цвета UI foundation (RGBA).
// Семантика: ally/enemy/active/dead/hover/target не смешиваются с gameplay-логикой — только отрисовка.
var Theme = struct {
	// Панели и фоны
	PanelBG               color.RGBA // основной фон панели
	PanelBGDeep           color.RGBA // чуть темнее (вкладки, подложки)
	PanelBorder           color.RGBA
	PanelTitleSep         color.RGBA // линия под заголовком
	OverlayDim            color.RGBA // затемнение полноэкранного overlay
	SceneTint             color.RGBA // «поле боя» placeholder
	BattlefieldTokenAlly  color.RGBA // заливка токена союзника на поле
	BattlefieldTokenEnemy color.RGBA // заливка токена врага на поле
	// Полосы рядов (v2): передний — ближе к линии столкновения; задний — к краю стороны.
	BattlefieldBackRowBand     color.RGBA
	BattlefieldFrontRowBand    color.RGBA
	BattlefieldFrontRowBorder  color.RGBA // лёгкая акцентная кромка переднего ряда со стороны центра
	BattlefieldEmptyCellBorder color.RGBA // пустые ячейки слабее PanelTitleSep
	// Сцена v2: лёгкое разделение зон и ось боя (vector, без ассетов).
	BattlefieldAllyZoneTint   color.RGBA // полупрозрачная подложка зоны союзников (слева от центра)
	BattlefieldEnemyZoneTint  color.RGBA // зона врагов (справа)
	BattlefieldCenterGutter   color.RGBA // узкий «канал» по линии столкновения
	BattlefieldSceneVignette  color.RGBA // лёгкое затемнение по краям сцены (оверлей)
	// Краткие вспышки feedback (juice) — альфа в базе умножается на интенсивность кадра
	FeedbackDamageOverlay color.RGBA
	FeedbackHealOverlay   color.RGBA
	FeedbackDeathOverlay  color.RGBA

	// Текст
	TextPrimary   color.RGBA
	TextSecondary color.RGBA
	TextMuted     color.RGBA
	TextDanger    color.RGBA
	TextSuccess   color.RGBA

	// Бой: стороны и состояния слотов/карт
	AllyAccent   color.RGBA // рамка союзника по умолчанию
	EnemyAccent  color.RGBA // враг (если нужен отдельный оттенок)
	ActiveTurn   color.RGBA // чей ход (acting)
	WaitAlly     color.RGBA // союзник ждёт очереди
	HoverTarget  color.RGBA
	SelectedKill color.RGBA // выбранная цель атаки
	ValidTarget  color.RGBA // допустимая цель
	EmptySlot    color.RGBA
	DeadFill     color.RGBA
	DeadText     color.RGBA

	// Способности / кнопки
	AbilityBG          color.RGBA
	AbilityBorder      color.RGBA
	AbilitySelectedBG  color.RGBA
	AbilitySelectedBrd color.RGBA
	AbilityHoverBG     color.RGBA
	ButtonBG           color.RGBA
	ButtonBorder       color.RGBA
	ButtonHoverBG      color.RGBA
	ButtonHoverBorder  color.RGBA
	DisabledFG         color.RGBA

	// HP micro-bars
	HPBarTrack  color.RGBA
	HPAllyFill  color.RGBA
	HPEnemyFill color.RGBA
	HPHealTint  color.RGBA // подсказка «heal» в превью (не обязательно на баре)

	// Explore / recovery
	ExploreBarBG     color.RGBA // подложка полоски подсказок
	ExploreBarBorder color.RGBA
	RecoveryBanner   color.RGBA
	HintLine         color.RGBA

	// Post-battle
	PostBattlePanelBG   color.RGBA
	PostBattleBorder    color.RGBA
	PostBattleRowSelect color.RGBA
	PostBattleRowBrd    color.RGBA

	// Ростер боя (v2): «карточка» — внутренняя подложка и рамка мини-портрета (как у inspect-карточки).
	RosterCardContentWell color.RGBA
	RosterCardInnerStroke color.RGBA
	RosterCardPortraitBG  color.RGBA

	// Акцентная полоса (header strip)
	AccentStrip color.RGBA
}{
	PanelBG:                    color.RGBA{R: 26, G: 28, B: 34, A: 255},
	PanelBGDeep:                visualcolor.Foundation.PanelBGDeep,
	PanelBorder:                visualcolor.Foundation.PanelBorder,
	PanelTitleSep:              visualcolor.Foundation.PanelTitleSep,
	OverlayDim:                 color.RGBA{R: 0, G: 0, B: 0, A: 200},
	SceneTint:                  visualcolor.Foundation.SceneTint,
	BattlefieldTokenAlly:       visualcolor.Foundation.BattlefieldTokenAlly,
	BattlefieldTokenEnemy:      color.RGBA{R: 110, G: 55, B: 65, A: 255},
	BattlefieldBackRowBand:     color.RGBA{R: 22, G: 26, B: 34, A: 200},
	BattlefieldFrontRowBand:    color.RGBA{R: 28, G: 34, B: 44, A: 210},
	BattlefieldFrontRowBorder:  color.RGBA{R: 72, G: 88, B: 108, A: 200},
	BattlefieldEmptyCellBorder: color.RGBA{R: 38, G: 42, B: 50, A: 130},
	BattlefieldAllyZoneTint:    color.RGBA{R: 45, G: 95, B: 85, A: 38},
	BattlefieldEnemyZoneTint:   color.RGBA{R: 105, G: 55, B: 72, A: 40},
	BattlefieldCenterGutter:    color.RGBA{R: 18, G: 22, B: 32, A: 235},
	BattlefieldSceneVignette:   color.RGBA{R: 0, G: 0, B: 0, A: 45},
	FeedbackDamageOverlay:      color.RGBA{R: 255, G: 90, B: 90, A: 125},
	FeedbackHealOverlay:        color.RGBA{R: 100, G: 220, B: 150, A: 115},
	FeedbackDeathOverlay:       color.RGBA{R: 60, G: 60, B: 70, A: 140},

	TextPrimary:   visualcolor.Foundation.TextPrimary,
	TextSecondary: color.RGBA{R: 190, G: 194, B: 204, A: 255},
	TextMuted:     color.RGBA{R: 130, G: 136, B: 150, A: 255},
	TextDanger:    color.RGBA{R: 230, G: 120, B: 120, A: 255},
	TextSuccess:   color.RGBA{R: 130, G: 210, B: 160, A: 255},

	AllyAccent:   color.RGBA{R: 85, G: 95, B: 115, A: 255},
	EnemyAccent:  visualcolor.Foundation.EnemyAccent,
	ActiveTurn:   visualcolor.Foundation.ActiveTurn,
	WaitAlly:     color.RGBA{R: 75, G: 95, B: 118, A: 255},
	HoverTarget:  visualcolor.Foundation.HoverTarget,
	SelectedKill: visualcolor.Foundation.SelectedKill,
	ValidTarget:  visualcolor.Foundation.ValidTarget,
	EmptySlot:    color.RGBA{R: 38, G: 40, B: 46, A: 255},
	DeadFill:     color.RGBA{R: 22, G: 22, B: 24, A: 255},
	DeadText:     color.RGBA{R: 105, G: 105, B: 110, A: 255},

	AbilityBG:          color.RGBA{R: 36, G: 38, B: 46, A: 255},
	AbilityBorder:      color.RGBA{R: 72, G: 76, B: 88, A: 255},
	AbilitySelectedBG:  color.RGBA{R: 52, G: 50, B: 32, A: 255},
	AbilitySelectedBrd: color.RGBA{R: 175, G: 165, B: 90, A: 255},
	AbilityHoverBG:     visualcolor.Foundation.AbilityHoverBG,
	ButtonBG:           color.RGBA{R: 38, G: 40, B: 48, A: 255},
	ButtonBorder:       color.RGBA{R: 125, G: 128, B: 142, A: 255},
	ButtonHoverBG:      color.RGBA{R: 55, G: 72, B: 95, A: 255},
	ButtonHoverBorder:  color.RGBA{R: 190, G: 210, B: 255, A: 255},
	DisabledFG:         color.RGBA{R: 165, G: 165, B: 170, A: 255},

	HPBarTrack:  visualcolor.Foundation.PanelBGDeep,
	HPAllyFill:  color.RGBA{R: 90, G: 175, B: 130, A: 255},
	HPEnemyFill: visualcolor.Foundation.HPEnemyFill,
	HPHealTint:  color.RGBA{R: 120, G: 200, B: 160, A: 255},

	ExploreBarBG:     color.RGBA{R: 12, G: 14, B: 20, A: 210},
	ExploreBarBorder: color.RGBA{R: 55, G: 62, B: 78, A: 255},
	RecoveryBanner:   color.RGBA{R: 110, G: 215, B: 155, A: 255},
	HintLine:         color.RGBA{R: 165, G: 172, B: 188, A: 255},

	PostBattlePanelBG:   color.RGBA{R: 26, G: 28, B: 36, A: 255},
	PostBattleBorder:    visualcolor.Foundation.PostBattleBorder,
	PostBattleRowSelect: color.RGBA{R: 52, G: 62, B: 88, A: 255},
	PostBattleRowBrd:    color.RGBA{R: 115, G: 135, B: 195, A: 255},

	RosterCardContentWell: color.RGBA{R: 20, G: 22, B: 28, A: 255},
	RosterCardInnerStroke: color.RGBA{R: 58, G: 64, B: 78, A: 220},
	RosterCardPortraitBG:  color.RGBA{R: 14, G: 16, B: 22, A: 255},

	AccentStrip: visualcolor.Foundation.AccentStrip,
}

// DrawHPBarMicro рисует компактный HP-бар (трек + заливка). cur/max в юнитах; side: "ally" | "enemy".
func DrawHPBarMicro(screen *ebiten.Image, x, y, w, h float32, cur, max int, alive bool, enemy bool) {
	if w <= 0 || h <= 0 || max <= 0 {
		return
	}
	vector.FillRect(screen, x, y, w, h, Theme.HPBarTrack, false)
	if !alive {
		return
	}
	ratio := float32(cur) / float32(max)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	fill := Theme.HPAllyFill
	if enemy {
		fill = Theme.HPEnemyFill
	}
	fw := w * ratio
	if fw < 1 && ratio > 0 {
		fw = 1
	}
	vector.FillRect(screen, x, y, fw, h, fill, false)
}

// DrawThinAccentLine горизонтальная акцентная линия (заголовок секции).
func DrawThinAccentLine(screen *ebiten.Image, x, y, w float32) {
	if w <= 0 {
		return
	}
	vector.FillRect(screen, x, y, w, 2, Theme.AccentStrip, false)
}
