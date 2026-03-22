package ui

import (
	"strings"
	"testing"

	"mygame/internal/hero"
)

func TestFormatLeaderHUDProgressionLine_includesLevel(t *testing.T) {
	h := hero.DefaultHero()
	h.CombatExperience = 2
	s := FormatLeaderHUDProgressionLine(&h)
	if s == "" {
		t.Fatal("empty")
	}
	if !strings.Contains(s, "ур.1") || !strings.Contains(s, "опыт 2") {
		t.Fatalf("unexpected: %q", s)
	}
}
