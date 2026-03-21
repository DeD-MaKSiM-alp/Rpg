package game

import (
	"errors"
	"testing"

	"mygame/internal/hero"
	"mygame/internal/party"
)

// Регрессия: тело успешного accept в recruit_offer (пул + резерв) без ebiten/input.
func TestRecruitCampOffer_addsHeroToReserve(t *testing.T) {
	p := party.Party{Active: []hero.Hero{hero.DefaultHero()}}
	serial, err := tryAddEarlyPoolRecruitToReserve(&p)
	if err != nil {
		t.Fatal(err)
	}
	if serial != 1 {
		t.Fatalf("serial %d, want 1", serial)
	}
	if p.ReserveCount() != 1 {
		t.Fatalf("reserve count %d", p.ReserveCount())
	}
	if p.Reserve[0].UnitID == "" {
		t.Fatal("expected template UnitID on recruit")
	}
}

func TestTryAddEarlyPoolRecruitToReserve_secondRecruitIncrementsSerial(t *testing.T) {
	p := party.Party{Active: []hero.Hero{hero.DefaultHero()}}
	if _, err := tryAddEarlyPoolRecruitToReserve(&p); err != nil {
		t.Fatal(err)
	}
	s2, err := tryAddEarlyPoolRecruitToReserve(&p)
	if err != nil {
		t.Fatal(err)
	}
	if s2 != 2 {
		t.Fatalf("serial %d, want 2", s2)
	}
	if p.ReserveCount() != 2 {
		t.Fatalf("reserve count %d", p.ReserveCount())
	}
}

func TestTryAddEarlyPoolRecruitToReserve_partyFull(t *testing.T) {
	p := party.Party{Active: []hero.Hero{hero.DefaultHero()}}
	for i := 0; i < party.MaxPartyMembers-1; i++ {
		if _, err := tryAddEarlyPoolRecruitToReserve(&p); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
	if p.TotalMembers() != party.MaxPartyMembers {
		t.Fatalf("total %d, want %d", p.TotalMembers(), party.MaxPartyMembers)
	}
	_, err := tryAddEarlyPoolRecruitToReserve(&p)
	if !errors.Is(err, party.ErrPartyFull) {
		t.Fatalf("got %v, want ErrPartyFull", err)
	}
}
