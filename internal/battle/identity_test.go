package battle

import "testing"

func TestPlayerTemplateIdentitySuffix(t *testing.T) {
	u := &CombatUnit{
		Side: BattleSidePlayer,
		Def: CombatUnitDefinition{
			TemplateUnitID: "empire_warrior_recruit",
			DisplayName:    "Союзник 1",
		},
	}
	if got := PlayerTemplateIdentitySuffix(u); got != " · empire_warrior_recruit" {
		t.Fatalf("got %q", got)
	}
	if PlayerTemplateIdentitySuffix(nil) != "" {
		t.Fatal("nil")
	}
	enemy := &CombatUnit{Side: BattleSideEnemy, Def: CombatUnitDefinition{TemplateUnitID: "x"}}
	if PlayerTemplateIdentitySuffix(enemy) != "" {
		t.Fatal("enemy should not show suffix")
	}
}
