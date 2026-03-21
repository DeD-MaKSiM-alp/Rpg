package hero

import (
	"fmt"

	battlepkg "mygame/internal/battle"
	"mygame/internal/unitdata"
)

// RecruitHeroFromEarlyPool — рекрут из циклического пула ранних шаблонов Империи.
// recruitSerial — порядковый номер найма в сессии (1, 2, 3…), как для RecruitDisplayName.
func RecruitHeroFromEarlyPool(recruitSerial int) Hero {
	ids := unitdata.EarlyRecruitUnitIDs()
	if len(ids) == 0 {
		return recruitHeroFallbackNoTemplate()
	}
	if recruitSerial < 1 {
		recruitSerial = 1
	}
	id := ids[(recruitSerial-1)%len(ids)]
	h, err := NewHeroFromUnitTemplate(id)
	if err != nil {
		return recruitHeroFallbackNoTemplate()
	}
	return h
}

// NewRecruitHero создаёт рекрута по базовому воинскому шаблону (один тип, совместимость со старым API).
func NewRecruitHero() Hero {
	h, err := NewHeroFromUnitTemplate(unitdata.EmpireWarriorRecruit)
	if err != nil {
		return recruitHeroFallbackNoTemplate()
	}
	return h
}

// recruitHeroFallbackNoTemplate — LEGACY: статы как у старого NewRecruitHero без UnitID (если registry недоступен).
func recruitHeroFallbackNoTemplate() Hero {
	h := Hero{
		MaxHP:     9,
		Attack:    2,
		Defense:   0,
		Initiative: 2,
		HealPower: 0,
		Abilities: []battlepkg.AbilityID{battlepkg.AbilityBasicAttack},
	}
	h.CurrentHP = h.MaxHP
	return h
}

// RecruitDisplayName — подпись для UI при получении новобранца (порядковый номер).
func RecruitDisplayName(serial int) string {
	return fmt.Sprintf("Новобранец %d", serial)
}
