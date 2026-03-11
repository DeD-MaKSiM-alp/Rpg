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

	panelX := float32(80)
	panelY := float32(80)
	panelW := float32(640)
	panelH := float32(360)

	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelColor, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2, panelBorderColor, false)
}

// drawBattleOverlayText рисует все текстовые блоки боевого overlay.
func drawBattleOverlayText(screen *ebiten.Image, hudFace *text.GoTextFace, battle *battlepkg.BattleContext) {
	panelX := float32(80)
	panelY := float32(80)

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(panelX)+20, float64(panelY)+35)
	titleOp.ColorScale.ScaleWithColor(color.White)

	title := "Battle mode"
	if battle != nil && len(battle.Encounter.Enemies) > 0 {
		title = fmt.Sprintf("Battle mode: enemy #%d", battle.Encounter.Enemies[0].EnemyID)
	}

	text.Draw(screen, title, hudFace, titleOp)

	if battle == nil {
		return
	}

	offsetY := 0.0
	if battle.Result != battlepkg.ResultNone {
		bannerY := float64(panelY) + 70
		bannerOp := &text.DrawOptions{}
		bannerOp.GeoM.Translate(float64(panelX)+20, bannerY)
		bannerOp.ColorScale.ScaleWithColor(color.White)
		var banner string
		switch battle.Result {
		case battlepkg.ResultVictory:
			banner = "VICTORY"
		case battlepkg.ResultDefeat:
			banner = "DEFEAT"
		case battlepkg.ResultEscape:
			banner = "ESCAPE"
		default:
			banner = battle.ResultString()
		}
		text.Draw(screen, banner, hudFace, bannerOp)

		hintOp := &text.DrawOptions{}
		hintOp.GeoM.Translate(float64(panelX)+20, bannerY+25)
		hintOp.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
		text.Draw(screen, "SPACE/ENTER: continue", hudFace, hintOp)
		offsetY = 55
	}

	bodyOp1 := &text.DrawOptions{}
	bodyOp1.GeoM.Translate(float64(panelX)+20, float64(panelY)+80+offsetY)
	bodyOp1.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Player HP: %d", battle.TeamFirstHP(battlepkg.TeamPlayer)), hudFace, bodyOp1)

	bodyOp2 := &text.DrawOptions{}
	bodyOp2.GeoM.Translate(float64(panelX)+20, float64(panelY)+110+offsetY)
	bodyOp2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Enemy HP: %d", battle.TeamFirstHP(battlepkg.TeamEnemy)), hudFace, bodyOp2)

	bodyOp3 := &text.DrawOptions{}
	bodyOp3.GeoM.Translate(float64(panelX)+20, float64(panelY)+125+offsetY)
	bodyOp3.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Раунд: %d | %s", battle.Round, battle.FormationSummary()), hudFace, bodyOp3)

	bodyOp3b := &text.DrawOptions{}
	bodyOp3b.GeoM.Translate(float64(panelX)+20, float64(panelY)+145+offsetY)
	bodyOp3b.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, battle.DisplayPhaseLabel(), hudFace, bodyOp3b)

	// Debug: phase, timer, active, result, last message
	bodyOp3c := &text.DrawOptions{}
	bodyOp3c.GeoM.Translate(float64(panelX)+20, float64(panelY)+165+offsetY)
	bodyOp3c.ColorScale.ScaleWithColor(color.RGBA{R: 180, G: 220, B: 180, A: 255})
	activeHP := 0
	if u := battle.ActiveUnit(); u != nil {
		activeHP = u.State.HP
	}
	lastMsg := battle.LastMessage
	if len(lastMsg) > 40 {
		lastMsg = lastMsg[:37] + "..."
	}
	text.Draw(screen, fmt.Sprintf("phase:%s timer:%d active:%s HP:%d result:%s | %s | %s",
		battle.PhaseString(), battle.PauseFrames, battle.ActiveUnitName(), activeHP, battle.ResultString(), lastMsg, battle.PlayerTurnStatusString()), hudFace, bodyOp3c)

	bodyOp4 := &text.DrawOptions{}
	bodyOp4.GeoM.Translate(float64(panelX)+20, float64(panelY)+195+offsetY)
	bodyOp4.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Arrows: select | SPACE/ENTER: confirm | Backspace: back | Esc: retreat", hudFace, bodyOp4)

	bodyOp5 := &text.DrawOptions{}
	bodyOp5.GeoM.Translate(float64(panelX)+20, float64(panelY)+220+offsetY)
	bodyOp5.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Tip: choose ability -> target -> confirm", hudFace, bodyOp5)

	bodyOp6 := &text.DrawOptions{}
	bodyOp6.GeoM.Translate(float64(panelX)+20, float64(panelY)+250+offsetY)
	bodyOp6.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, battle.LastMessage, hudFace, bodyOp6)
}

// drawHUDText рисует текстовые блоки HUD (счётчик собранных предметов и т.п.).
func drawHUDText(screen *ebiten.Image, pickupCount int, hudFace *text.GoTextFace) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, 20)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fmt.Sprintf("Pickups: %d", pickupCount), hudFace, op)
}
