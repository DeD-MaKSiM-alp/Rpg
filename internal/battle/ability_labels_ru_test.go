package battle

import "testing"

func TestPlayerAbilityLabelRU_knownIDs(t *testing.T) {
	cases := map[AbilityID]string{
		AbilityBasicAttack:  "Базовый удар",
		AbilityRangedAttack: "Дальний удар",
		AbilityHeal:         "Лечение",
		AbilityGroupHeal:    "Масс-лечение",
		AbilityBuff:         "Усиление",
		AbilityPowerStrike:  "Мощный удар",
	}
	for id, want := range cases {
		if got := PlayerAbilityLabelRU(id); got != want {
			t.Fatalf("id %d: want %q, got %q", id, want, got)
		}
	}
}

func TestPlayerAbilityLabelRU_unknownFallback(t *testing.T) {
	if got := PlayerAbilityLabelRU(AbilityID(999)); got != "Способность" || got == "" {
		t.Fatalf("unexpected fallback: %q", got)
	}
}
