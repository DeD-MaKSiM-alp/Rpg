package postbattle

import (
	"slices"
	"testing"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
	"mygame/internal/progression"
)

func TestFlow_Begin_IsActive_and_Reset(t *testing.T) {
	var f Flow
	if f.IsActive() {
		t.Fatal("new Flow should not be active")
	}
	f.Begin(battlepkg.BattleOutcomeVictory, []string{"line1"})
	if !f.IsActive() || f.Step != StepResult {
		t.Fatalf("after Begin victory: active=%v step=%v", f.IsActive(), f.Step)
	}
	if f.Outcome != battlepkg.BattleOutcomeVictory {
		t.Fatalf("outcome %v", f.Outcome)
	}
	if len(f.VictorySummaryLines) != 1 || f.VictorySummaryLines[0] != "line1" {
		t.Fatalf("summary %+v", f.VictorySummaryLines)
	}
	f.Reset()
	if f.IsActive() || f.Step != StepNone {
		t.Fatalf("after Reset: active=%v step=%v", f.IsActive(), f.Step)
	}
}

func TestFlow_Begin_defeat_noVictorySummary(t *testing.T) {
	var f Flow
	f.Begin(battlepkg.BattleOutcomeDefeat, []string{"ignored"})
	if f.Step != StepResult {
		t.Fatalf("step %v", f.Step)
	}
	if len(f.VictorySummaryLines) != 0 {
		t.Fatalf("defeat should not store victory lines, got %+v", f.VictorySummaryLines)
	}
}

func TestFlow_confirmResultStep_wrongStep_returnsFalse(t *testing.T) {
	var f Flow
	p := party.DefaultParty()
	if f.confirmResultStep(&p) {
		t.Fatal("confirmResultStep on StepNone should not end")
	}
}

func TestFlow_confirmResultStep_nilLeader_ends(t *testing.T) {
	var f Flow
	f.Begin(battlepkg.BattleOutcomeVictory, nil)
	empty := party.Party{Active: []hero.Hero{}}
	if !f.confirmResultStep(&empty) {
		t.Fatal("expected end when leader missing")
	}
	if f.Step != StepResult {
		t.Fatalf("step should stay StepResult (no transition), got %v", f.Step)
	}
}

func TestFlow_confirmResultStep_defeat_endsWithoutReward(t *testing.T) {
	var f Flow
	f.Begin(battlepkg.BattleOutcomeDefeat, nil)
	p := party.DefaultParty()
	if !f.confirmResultStep(&p) {
		t.Fatal("defeat should end post-battle immediately")
	}
	if f.Step != StepResult {
		t.Fatalf("step %v (caller ends battle; step not advanced to reward)", f.Step)
	}
	if len(f.RewardOffer) != 0 {
		t.Fatalf("unexpected reward offer %+v", f.RewardOffer)
	}
}

func TestFlow_confirmResultStep_retreat_endsWithoutReward(t *testing.T) {
	var f Flow
	f.Begin(battlepkg.BattleOutcomeRetreat, nil)
	p := party.DefaultParty()
	if !f.confirmResultStep(&p) {
		t.Fatal("retreat should end post-battle immediately")
	}
}

func TestFlow_confirmResultStep_victory_transitionsToRewardWithOffer(t *testing.T) {
	var f Flow
	f.Begin(battlepkg.BattleOutcomeVictory, []string{"xp"})
	p := party.DefaultParty()
	if f.confirmResultStep(&p) {
		t.Fatal("victory with offer should not end before reward pick")
	}
	if f.Step != StepReward {
		t.Fatalf("step %v want StepReward", f.Step)
	}
	n := len(f.RewardOffer)
	if n == 0 {
		t.Fatal("expected non-empty reward offer for default leader")
	}
	if n > progression.RewardOfferCount {
		t.Fatalf("offer len %d > RewardOfferCount", n)
	}
	if f.SelectedIndex != 0 {
		t.Fatalf("SelectedIndex %d", f.SelectedIndex)
	}
}

func TestFlow_confirmRewardSelection_invalidIndex(t *testing.T) {
	var f Flow
	f.Begin(battlepkg.BattleOutcomeVictory, nil)
	p := party.DefaultParty()
	if f.confirmResultStep(&p) {
		t.Fatal("need reward step")
	}
	if f.confirmRewardSelection(&p, -1) {
		t.Fatal("invalid index should not end")
	}
	if f.confirmRewardSelection(&p, len(f.RewardOffer)) {
		t.Fatal("out of range should not end")
	}
}

func TestFlow_confirmRewardSelection_appliesRewardAndEnds(t *testing.T) {
	var f Flow
	f.Begin(battlepkg.BattleOutcomeVictory, nil)
	p := party.DefaultParty()
	if f.confirmResultStep(&p) {
		t.Fatal("need reward step")
	}
	leader := p.Leader()
	if leader == nil {
		t.Fatal("leader")
	}
	beforeHP, beforeAtk, beforeDef, beforeIni := leader.MaxHP, leader.Attack, leader.Defense, leader.Initiative
	beforeHeal, beforeBonus := leader.HealPower, leader.BasicAttackBonus
	beforeAbils := slices.Clone(leader.Abilities)
	idx := 0
	k := f.RewardOffer[idx]
	if !f.confirmRewardSelection(&p, idx) {
		t.Fatal("expected end after valid pick")
	}
	switch k {
	case progression.RewardMaxHP:
		if leader.MaxHP != beforeHP+2 || leader.CurrentHP != beforeHP+2 {
			t.Fatalf("MaxHP/CurrentHP %+v", leader)
		}
	case progression.RewardAttack:
		if leader.Attack != beforeAtk+1 {
			t.Fatalf("Attack %d want %d", leader.Attack, beforeAtk+1)
		}
	case progression.RewardDefense:
		if leader.Defense != beforeDef+1 {
			t.Fatalf("Defense %d want %d", leader.Defense, beforeDef+1)
		}
	case progression.RewardInitiative:
		if leader.Initiative != beforeIni+1 {
			t.Fatalf("Initiative %d want %d", leader.Initiative, beforeIni+1)
		}
	case progression.RewardAbilityHeal:
		if slices.Equal(leader.Abilities, beforeAbils) {
			t.Fatal("expected Heal ability added")
		}
		if !slices.Contains(leader.Abilities, battlepkg.AbilityHeal) {
			t.Fatal("missing AbilityHeal")
		}
	case progression.RewardAbilityRanged:
		if !slices.Contains(leader.Abilities, battlepkg.AbilityRangedAttack) {
			t.Fatal("missing AbilityRangedAttack")
		}
	case progression.RewardHealUpgrade:
		if leader.HealPower != beforeHeal+2 {
			t.Fatalf("HealPower %d want %d", leader.HealPower, beforeHeal+2)
		}
	case progression.RewardBasicAttackUpgrade:
		if leader.BasicAttackBonus != beforeBonus+1 {
			t.Fatalf("BasicAttackBonus %d want %d", leader.BasicAttackBonus, beforeBonus+1)
		}
	default:
		t.Fatalf("unexpected reward kind in offer[0]: %v", k)
	}
}

func TestFlow_wrapRewardIndex(t *testing.T) {
	tests := []struct {
		idx, n, want int
	}{
		{0, 3, 0},
		{2, 3, 2},
		{-1, 3, 2},
		{3, 3, 0},
		{0, 0, 0},
		{5, 3, 2},
	}
	for _, tc := range tests {
		if got := wrapRewardIndex(tc.idx, tc.n); got != tc.want {
			t.Fatalf("wrapRewardIndex(%d,%d)=%d want %d", tc.idx, tc.n, got, tc.want)
		}
	}
}
