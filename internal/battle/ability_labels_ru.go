package battle

// PlayerAbilityLabelRU — единые короткие русские подписи способностей для inspect, battle HUD и боевого лога.
// Игровые id способностей не меняются; это только presentation layer.
func PlayerAbilityLabelRU(id AbilityID) string {
	switch id {
	case AbilityBasicAttack:
		return "Базовый удар"
	case AbilityRangedAttack:
		return "Дальний удар"
	case AbilityHeal:
		return "Лечение"
	case AbilityGroupHeal:
		return "Масс-лечение"
	case AbilityBuff:
		return "Усиление"
	case AbilityPowerStrike:
		return "Мощный удар"
	default:
		return "Способность"
	}
}
