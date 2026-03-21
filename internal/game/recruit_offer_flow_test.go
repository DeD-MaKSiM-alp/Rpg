package game

import (
	"testing"

	"mygame/internal/hero"
	"mygame/internal/party"
)

// Регрессия: тело успешного accept в recruit_offer (пул + резерв) без ebiten/input.
func TestRecruitCampOffer_addsHeroToReserve(t *testing.T) {
	p := party.Party{Active: []hero.Hero{hero.DefaultHero()}}
	idx := len(p.Reserve) + 1
	h := hero.RecruitHeroFromEarlyPool(idx)
	h.RecruitLabel = hero.RecruitDisplayName(idx)
	if err := p.AddToReserve(h); err != nil {
		t.Fatal(err)
	}
	if p.ReserveCount() != 1 {
		t.Fatalf("reserve count %d", p.ReserveCount())
	}
	if h.UnitID == "" {
		t.Fatal("expected template UnitID on recruit")
	}
}
