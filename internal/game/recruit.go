package game

import (
	"fmt"

	"mygame/internal/hero"
	"mygame/internal/party"
)

// tryAddEarlyPoolRecruitToReserve adds one hero from the early rotating pool to the party reserve.
// Serial for pool rotation and labels is len(Reserve)+1 at call time (same rule as historic F9 / camp flows).
// Returns party.ErrPartyFull if the roster has no room.
func tryAddEarlyPoolRecruitToReserve(p *party.Party) (serial int, err error) {
	if p == nil {
		return 0, fmt.Errorf("party: nil")
	}
	if p.TotalMembers() >= party.MaxPartyMembers {
		return 0, party.ErrPartyFull
	}
	serial = len(p.Reserve) + 1
	h := hero.RecruitHeroFromEarlyPool(serial)
	h.RecruitLabel = hero.RecruitDisplayName(serial)
	err = p.AddToReserve(h)
	return serial, err
}
