package world

// PickupInteractionResult — результат взаимодействия с пикапом после шага на клетку.
type PickupInteractionResult int

const (
	// PickupInteractNone — нет пикапа или уже собран.
	PickupInteractNone PickupInteractionResult = iota
	// PickupInteractResource — обычный пикап собран автоматически.
	PickupInteractResource
	// PickupInteractRecruitOffer — лагерь рекрута: игра должна показать подтверждение; пикап ещё не помечен собранным.
	PickupInteractRecruitOffer
	// PickupInteractPOIRequiresChoice — POI с выбором (руины/алтарь): пикап ещё не собран; эффект после подтверждения в game.
	PickupInteractPOIRequiresChoice
	// PickupInteractPOI* — POI: пикап помечен собранным; эффект в internal/game.
	PickupInteractPOIAltar
	PickupInteractPOISpring
	PickupInteractPOICache
	PickupInteractPOIRuins
	PickupInteractPOICampfire
)
