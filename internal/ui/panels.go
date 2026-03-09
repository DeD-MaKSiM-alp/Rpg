package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	battlepkg "mygame/internal/battle"
)

// drawBattleOverlayPanel рисует затемнённый фон и центральную панель боевого overlay.
func drawBattleOverlayPanel(screen *ebiten.Image, screenWidth, screenHeight int) {
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	panelColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	panelBorderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)

	panelX := float32(120)
	panelY := float32(140)
	panelW := float32(560)
	panelH := float32(260)

	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelColor, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2, panelBorderColor, false)
}

// drawBattleOverlayText рисует все текстовые блоки боевого overlay.
func drawBattleOverlayText(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext) {
	panelX := float32(120)
	panelY := float32(140)

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(panelX)+20, float64(panelY)+35)
	titleOp.ColorScale.ScaleWithColor(color.White)

	title := "Battle mode"
	if battle != nil {
		title = fmt.Sprintf("Battle mode: enemy #%d", battle.EnemyID)
	}

	text.Draw(screen, title, hudFace, titleOp)

	if battle == nil {
		return
	}

	bodyOp1 := &text.DrawOptions{}
	bodyOp1.GeoM.Translate(float64(panelX)+20, float64(panelY)+80)
	bodyOp1.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Player HP: %d", battle.PlayerHP), hudFace, bodyOp1)

	bodyOp2 := &text.DrawOptions{}
	bodyOp2.GeoM.Translate(float64(panelX)+20, float64(panelY)+110)
	bodyOp2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Enemy HP: %d", battle.EnemyHP), hudFace, bodyOp2)

	bodyOp3 := &text.DrawOptions{}
	bodyOp3.GeoM.Translate(float64(panelX)+20, float64(panelY)+145)
	bodyOp3.ColorScale.ScaleWithColor(color.White)
	phaseText := "Фаза: неизвестно"
	switch battle.Phase {
	case battlepkg.BattlePhasePlayerTurn:
		phaseText = "Фаза: ход игрока"
	case battlepkg.BattlePhaseEnemyTurn:
		phaseText = "Фаза: ход врага"
	case battlepkg.BattlePhaseFinished:
		phaseText = "Фаза: бой завершён"
	}
	text.Draw(screen, phaseText, hudFace, bodyOp3)

	bodyOp4 := &text.DrawOptions{}
	bodyOp4.GeoM.Translate(float64(panelX)+20, float64(panelY)+180)
	bodyOp4.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Space - attack", hudFace, bodyOp4)

	bodyOp5 := &text.DrawOptions{}
	bodyOp5.GeoM.Translate(float64(panelX)+20, float64(panelY)+210)
	bodyOp5.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Esc - retreat", hudFace, bodyOp5)

	bodyOp6 := &text.DrawOptions{}
	bodyOp6.GeoM.Translate(float64(panelX)+20, float64(panelY)+240)
	bodyOp6.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, battle.LastLog, hudFace, bodyOp6)
}

// drawHUDText рисует текстовые блоки HUD (счётчик собранных предметов и т.п.).
func drawHUDText(screen *ebiten.Image, pickupCount int, hudFace *text.GoTextFace) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, 20)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Pickups: %d", pickupCount), hudFace, op)
}
