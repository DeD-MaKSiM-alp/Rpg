package hero

import "testing"

func TestCombatLevelFromTotalXP(t *testing.T) {
	if CombatLevelFromTotalXP(-1) != 1 {
		t.Fatal("negative xp")
	}
	if CombatLevelFromTotalXP(0) != 1 {
		t.Fatal("lvl 1 at 0")
	}
	if CombatLevelFromTotalXP(3) != 1 {
		t.Fatal("lvl 1 at 3")
	}
	if CombatLevelFromTotalXP(4) != 2 {
		t.Fatal("lvl 2 at 4")
	}
}

func TestHero_CombatXPToNextLevel(t *testing.T) {
	h := Hero{CombatExperience: 0}
	if h.CombatXPToNextLevel() != CombatXPPerLevel {
		t.Fatalf("at 0 want %d, got %d", CombatXPPerLevel, h.CombatXPToNextLevel())
	}
	h.CombatExperience = 3
	if h.CombatXPToNextLevel() != 1 {
		t.Fatalf("at 3 want 1, got %d", h.CombatXPToNextLevel())
	}
	h.CombatExperience = 4
	if h.CombatXPToNextLevel() != CombatXPPerLevel {
		t.Fatalf("at 4 want %d, got %d", CombatXPPerLevel, h.CombatXPToNextLevel())
	}
}

func TestHero_EffectiveBasicAttackBonusMatchesLevelChunks(t *testing.T) {
	h := Hero{CombatExperience: 7, BasicAttackBonus: 1}
	// 7/4 = 1 from level, +1 reward = 2
	if h.EffectiveBasicAttackBonusForCombat() != 2 {
		t.Fatalf("want 2, got %d", h.EffectiveBasicAttackBonusForCombat())
	}
}
