package ui

import (
	"strings"
	"testing"

	battlepkg "mygame/internal/battle"
)

func TestBattleCompactUnavailableHintRU_shortOnSmallOnly(t *testing.T) {
	cases := []struct {
		full   string
		expect string
	}{
		{"КД: ещё 3 р.", "КД 3р"},
		{"Мана: нужно 10, сейчас 4", "мана 4/10"},
		{"Энергия: нужно 5, сейчас 2", "энерг. 2/5"},
	}
	for _, tc := range cases {
		got := battleCompactUnavailableHintRU(TierSmall, tc.full)
		if got != tc.expect {
			t.Fatalf("small: %q -> want %q, got %q", tc.full, tc.expect, got)
		}
		if g2 := battleCompactUnavailableHintRU(TierMedium, tc.full); g2 != tc.full {
			t.Fatalf("medium must pass through: %q -> got %q", tc.full, g2)
		}
	}
}

func TestBattleV1FooterControlsHintRU_shorterOnSmall(t *testing.T) {
	b := &battlepkg.BattleContext{}
	short := battleV1FooterControlsHintRU(b, TierSmall)
	if len(short) > 80 {
		t.Fatalf("small footer hint unexpectedly long: %q", short)
	}
	full := battleV1FooterControlsHintRU(b, TierLarge)
	if len(full) <= len(short)+10 {
		t.Fatalf("large hint should be longer than small: short=%d full=%d", len(short), len(full))
	}
}

func TestBattleV2BottomHintRU_shorterOnSmall(t *testing.T) {
	b := &battlepkg.BattleContext{}
	s := battleV2BottomHintRU(b, TierSmall)
	if !strings.Contains(s, "ПКМ") && !strings.Contains(s, "Esc") {
		t.Fatalf("unexpected small hint: %q", s)
	}
	f := battleV2BottomHintRU(b, TierLarge)
	if len(f) <= len(s)+15 {
		t.Fatalf("large hint should exceed small: s=%d f=%d", len(s), len(f))
	}
}

func TestApplyExploreHintOverflow_dropsOptionalByPriority(t *testing.T) {
	z, rest, rec, poi, inter := ApplyExploreHintOverflow(
		TierSmall,
		"zone",
		"rest",
		"recruit",
		"poi",
		"interaction",
	)
	if poi != "" {
		t.Fatalf("expected poi dropped first on small tier, got %q", poi)
	}
	if z == "" || inter == "" {
		t.Fatalf("expected zone and interaction to survive after one drop: z=%q inter=%q", z, inter)
	}
	_ = rest
	_ = rec
}

func TestPresetForTier_bottomMaxLinesOrdered(t *testing.T) {
	s := presetForTier(TierSmall).BottomMaxLines
	m := presetForTier(TierMedium).BottomMaxLines
	l := presetForTier(TierLarge).BottomMaxLines
	if !(s <= m && m <= l) {
		t.Fatalf("BottomMaxLines should not decrease with tier: small=%d medium=%d large=%d", s, m, l)
	}
}
