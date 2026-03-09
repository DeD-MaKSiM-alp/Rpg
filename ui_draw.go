package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawHUD рисует поверх кадра элементы HUD (например, счётчик собранных предметов).
func (g *Game) drawHUD(screen *ebiten.Image) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, 20)
	op.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		fmt.Sprintf("Pickups: %d", g.pickupCount),
		g.hudFace,
		op,
	)
}

/*
drawBattleOverlay рисует поверх кадра HUD для боевого режима.
- Затемняет фон;
- Рисует центральную панель;
- Показывает ID активного врага;
- Показывает кнопки для победы и отступления.
*/
func (g *Game) drawBattleOverlay(screen *ebiten.Image) {
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	panelColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	panelBorderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	// Затемняем фон.
	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)

	// Центральная панель.
	panelX := float32(120)
	panelY := float32(140)
	panelW := float32(560)
	panelH := float32(260)

	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelColor, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2, panelBorderColor, false)

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(panelX)+20, float64(panelY)+35)
	titleOp.ColorScale.ScaleWithColor(color.White)

	title := "Battle mode"

	// Если контекст боя есть, показываем ID активного врага.
	if g.battle != nil {
		title = fmt.Sprintf("Battle mode: enemy #%d", g.battle.EnemyID)
	}

	text.Draw(
		screen,
		title,
		g.hudFace,
		titleOp,
	)

	// Если контекст отсутствует, дальше ничего не рисуем.
	if g.battle == nil {
		return
	}

	bodyOp1 := &text.DrawOptions{}
	bodyOp1.GeoM.Translate(float64(panelX)+20, float64(panelY)+80)
	bodyOp1.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		fmt.Sprintf("Player HP: %d", g.battle.PlayerHP),
		g.hudFace,
		bodyOp1,
	)

	bodyOp2 := &text.DrawOptions{}
	bodyOp2.GeoM.Translate(float64(panelX)+20, float64(panelY)+110)
	bodyOp2.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		fmt.Sprintf("Enemy HP: %d", g.battle.EnemyHP),
		g.hudFace,
		bodyOp2,
	)

	bodyOp3 := &text.DrawOptions{}
	bodyOp3.GeoM.Translate(float64(panelX)+20, float64(panelY)+145)
	bodyOp3.ColorScale.ScaleWithColor(color.White)

	phaseText := "Фаза: неизвестно"
	switch g.battle.Phase {
	case BattlePhasePlayerTurn:
		phaseText = "Фаза: ход игрока"
	case BattlePhaseEnemyTurn:
		phaseText = "Фаза: ход врага"
	case BattlePhaseFinished:
		phaseText = "Фаза: бой завершён"
	}

	text.Draw(
		screen,
		phaseText,
		g.hudFace,
		bodyOp3,
	)

	bodyOp4 := &text.DrawOptions{}
	bodyOp4.GeoM.Translate(float64(panelX)+20, float64(panelY)+180)
	bodyOp4.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		"Space - attack",
		g.hudFace,
		bodyOp4,
	)

	bodyOp5 := &text.DrawOptions{}
	bodyOp5.GeoM.Translate(float64(panelX)+20, float64(panelY)+210)
	bodyOp5.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		"Esc - retreat",
		g.hudFace,
		bodyOp5,
	)

	bodyOp6 := &text.DrawOptions{}
	bodyOp6.GeoM.Translate(float64(panelX)+20, float64(panelY)+240)
	bodyOp6.ColorScale.ScaleWithColor(color.White)

	text.Draw(
		screen,
		g.battle.LastLog,
		g.hudFace,
		bodyOp6,
	)
}
