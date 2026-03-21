package game

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"mygame/internal/hero"
	"mygame/internal/party"
)

// updateRecruitOfferMode — подтверждение найма с лагеря на карте (лазурный маркер).
func (g *Game) updateRecruitOfferMode() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.mode = ModeExplore
		g.advanceWorldTurn()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyY) {
		serial, err := tryAddEarlyPoolRecruitToReserve(&g.party)
		if err != nil {
			if errors.Is(err, party.ErrPartyFull) {
				g.exploreRecruitMsg = fmt.Sprintf("Отряд полон (макс. %d). Лагерь остаётся на месте.", party.MaxPartyMembers)
			} else {
				g.exploreRecruitMsg = err.Error()
			}
		} else {
			g.world.MarkRecruitPickupCollected(g.recruitOfferX, g.recruitOfferY)
			g.exploreRecruitMsg = fmt.Sprintf("В резерв: %s (F5 — состав)", hero.RecruitDisplayName(serial))
		}
		g.exploreRecruitMsgTicks = exploreRestFeedbackDurationTicks
		g.mode = ModeExplore
		g.advanceWorldTurn()
	}
}
