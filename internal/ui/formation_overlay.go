package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mygame/internal/party"
)

// DrawFormationOverlay — минимальный экран порядка отряда в explore (канонический порядок = party.Active).
func DrawFormationOverlay(screen *ebiten.Image, hudFace *text.GoTextFace, p *party.Party, selected int, screenW, screenH int) {
	if hudFace == nil || p == nil {
		return
	}
	sw, sh := float32(screenW), float32(screenH)
	vector.FillRect(screen, 0, 0, sw, sh, color.RGBA{0, 0, 0, 200}, false)

	pad := float32(24)
	lineH := uiLineH
	x := pad
	y := pad

	title := "Порядок отряда → построение в бою"
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	text.Draw(screen, title, hudFace, op)
	y += lineH * 1.4

	sub := "Сверху вниз: передний ряд (3), затем задний — как при входе в бой (PlayerSlotForPartyIndex)."
	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	op2.ColorScale.ScaleWithColor(color.RGBA{R: 190, G: 195, B: 200, A: 255})
	text.Draw(screen, sub, hudFace, op2)
	y += lineH * 2.2

	n := len(p.Active)
	if n == 0 {
		op3 := &text.DrawOptions{}
		op3.GeoM.Translate(float64(x), float64(y))
		op3.ColorScale.ScaleWithColor(color.RGBA{R: 200, G: 100, B: 100, A: 255})
		text.Draw(screen, "Нет участников.", hudFace, op3)
		return
	}

	for i := 0; i < n; i++ {
		h := p.Active[i]
		slot := party.FormationSlotCaption(i)
		role := party.MemberRoleCaption(i)
		mark := "  "
		if i == selected {
			mark = "▶ "
		}
		line := fmt.Sprintf("%s%d. %s · %s · HP %d/%d", mark, i+1, role, slot, h.CurrentHP, h.MaxHP)
		col := color.RGBA{R: 220, G: 225, B: 230, A: 255}
		if i == selected {
			col = color.RGBA{R: 255, G: 230, B: 120, A: 255}
		}
		opL := &text.DrawOptions{}
		opL.GeoM.Translate(float64(x), float64(y))
		opL.ColorScale.ScaleWithColor(col)
		text.Draw(screen, line, hudFace, opL)
		y += lineH * 1.25
	}

	y += lineH * 0.5
	help := "↑↓ выбор   ←→ сдвиг в списке (меняет слот в бою)   Esc — выход"
	if n < 2 {
		help = "Нужно минимум 2 участника, чтобы менять порядок.   Esc — выход"
	}
	opH := &text.DrawOptions{}
	opH.GeoM.Translate(float64(x), float64(y))
	opH.ColorScale.ScaleWithColor(color.RGBA{R: 160, G: 170, B: 185, A: 255})
	text.Draw(screen, help, hudFace, opH)
}

// DrawExploreFormationHint — подсказки в explore: F5 (порядок), R (отдых), опционально баннер после отдыха.
// Снизу вверх: F5, R, при наличии — зелёный баннер после отдыха.
func DrawExploreFormationHint(screen *ebiten.Image, hudFace *text.GoTextFace, screenW, screenH int, restFeedback string) {
	if hudFace == nil || screenH < 40 {
		return
	}
	yF5 := float64(screenH) - 26
	yR := yF5 - 22
	yFeed := yR - 22
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, yF5)
	op.ColorScale.ScaleWithColor(color.RGBA{R: 170, G: 175, B: 185, A: 255})
	text.Draw(screen, "F5: порядок отряда (слоты в бою как в списке)", hudFace, op)
	opR := &text.DrawOptions{}
	opR.GeoM.Translate(10, yR)
	opR.ColorScale.ScaleWithColor(color.RGBA{R: 165, G: 170, B: 180, A: 255})
	text.Draw(screen, "R: отдых — +HP живым (доля MaxHP), 0 HP не поднимает; затем ход мира", hudFace, opR)
	if restFeedback != "" {
		opF := &text.DrawOptions{}
		opF.GeoM.Translate(10, yFeed)
		opF.ColorScale.ScaleWithColor(color.RGBA{R: 120, G: 220, B: 160, A: 255})
		text.Draw(screen, restFeedback, hudFace, opF)
	}
}
