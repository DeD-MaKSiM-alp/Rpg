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
		if g.party.TotalMembers() >= party.MaxPartyMembers {
			g.exploreRecruitMsg = fmt.Sprintf("Отряд полон (макс. %d). Лагерь остаётся на месте.", party.MaxPartyMembers)
			g.exploreRecruitMsgTicks = exploreRestFeedbackDurationTicks
			g.mode = ModeExplore
			g.advanceWorldTurn()
			return
		}
		idx := len(g.party.Reserve) + 1
		h := hero.RecruitHeroFromEarlyPool(idx)
		h.RecruitLabel = hero.RecruitDisplayName(idx)
		if err := g.party.AddToReserve(h); err != nil {
			if errors.Is(err, party.ErrPartyFull) {
				g.exploreRecruitMsg = fmt.Sprintf("Отряд полон (макс. %d)", party.MaxPartyMembers)
			} else {
				g.exploreRecruitMsg = err.Error()
			}
		} else {
			g.world.MarkRecruitPickupCollected(g.recruitOfferX, g.recruitOfferY)
			g.exploreRecruitMsg = fmt.Sprintf("В резерв: %s (F5 — состав)", hero.RecruitDisplayName(idx))
		}
		g.exploreRecruitMsgTicks = exploreRestFeedbackDurationTicks
		g.mode = ModeExplore
		g.advanceWorldTurn()
	}
}
