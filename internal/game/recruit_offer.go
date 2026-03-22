package game

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"mygame/internal/hero"
	"mygame/internal/party"
	"mygame/internal/ui"
)

// updateRecruitOfferMode — подтверждение найма с лагеря на карте (лазурный маркер).
func (g *Game) updateRecruitOfferMode() {
	mx, my := ebiten.CursorPosition()
	g.recruitOfferHoverBtn = -1
	switch ui.HitTestRecruitOffer(mx, my, ScreenWidth, ScreenHeight) {
	case ui.RecruitHitAccept:
		g.recruitOfferHoverBtn = 0
	case ui.RecruitHitDecline:
		g.recruitOfferHoverBtn = 1
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		switch ui.HitTestRecruitOffer(mx, my, ScreenWidth, ScreenHeight) {
		case ui.RecruitHitAccept:
			g.completeRecruitOfferAccept()
			return
		case ui.RecruitHitDecline, ui.RecruitHitBackdrop:
			g.cancelRecruitOffer()
			return
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.cancelRecruitOffer()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyY) {
		g.completeRecruitOfferAccept()
	}
}

func (g *Game) cancelRecruitOffer() {
	g.mode = ModeExplore
	g.advanceWorldTurn()
}

func (g *Game) completeRecruitOfferAccept() {
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
