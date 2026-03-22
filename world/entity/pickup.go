package entity

// PickupKind — тип подбираемого объекта (world-integrated recruit и т.д.).
type PickupKind int

const (
	// PickupKindResource — обычный пикап (счётчик в HUD).
	PickupKindResource PickupKind = iota
	// PickupKindRecruitCamp — лагерь наёмников: требует подтверждения, затем канонический recruit в резерв.
	PickupKindRecruitCamp
	// --- POI (points of interest): одноразовые точки, эффект в game после шага на клетку ---
	// PickupKindPOIAltar — алтарь: выбор награды (скромная / смелая) в explore.
	PickupKindPOIAltar
	// PickupKindPOISpring — источник: лёгкое восстановление ОЗ.
	PickupKindPOISpring
	// PickupKindPOICache — тайник: +3 к счётчику добычи.
	PickupKindPOICache
	// PickupKindPOIRuins — руины: выбор (осторожно / риск) в explore.
	PickupKindPOIRuins
	// PickupKindPOICampfire — привал: +1 знак и небольшое лечение.
	PickupKindPOICampfire
)

// IsPOIKind — true для типов POI (не ресурс и не лагерь рекрута).
func IsPOIKind(k PickupKind) bool {
	return k >= PickupKindPOIAltar && k <= PickupKindPOICampfire
}

// Pickup — подбираемый объект в мире.
type Pickup struct {
	X         int
	Y         int
	Collected bool
	Kind      PickupKind // нулевое значение = PickupKindResource
}
