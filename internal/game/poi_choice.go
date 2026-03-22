package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"mygame/internal/party"
	"mygame/world/entity"
)

// altarBoldHPLossPreview — ожидаемый урон лидеру для подсказки на алтаре (см. applyAltarPOIChoiceBold).
func altarBoldHPLossPreview(p *party.Party) int {
	if p == nil {
		return 0
	}
	lh := p.Leader()
	if lh == nil || lh.CurrentHP <= 0 {
		return 0
	}
	loss := lh.MaxHP / 5
	if loss < 1 {
		loss = 1
	}
	return loss
}

// updatePOIChoiceMode — risk/reward для руин и алтаря (аналогично recruit: Esc без сбора, Enter — выбранный вариант).
func (g *Game) updatePOIChoiceMode() {
	switch g.poiChoiceKind {
	case entity.PickupKindPOIRuins, entity.PickupKindPOIAltar:
	default:
		g.mode = ModeExplore
		g.advanceWorldTurn()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mode = ModeExplore
		g.advanceWorldTurn()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) ||
		inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.poiChoiceSel = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad2) ||
		inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.poiChoiceSel = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.poiChoiceSel ^= 1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.resolvePOIChoice()
	}
}

func (g *Game) resolvePOIChoice() {
	switch g.poiChoiceKind {
	case entity.PickupKindPOIRuins:
		if g.poiChoiceSel == 0 {
			g.applyRuinsPOIChoiceSafe()
		} else {
			g.applyRuinsPOIChoiceRisky()
		}
	case entity.PickupKindPOIAltar:
		if g.poiChoiceSel == 0 {
			g.applyAltarPOIChoiceModest()
		} else {
			g.applyAltarPOIChoiceBold()
		}
	default:
		g.mode = ModeExplore
		g.advanceWorldTurn()
		return
	}

	g.world.MarkPOIPickupCollected(g.poiChoiceX, g.poiChoiceY)
	g.mode = ModeExplore
	g.advanceWorldTurn()
}
