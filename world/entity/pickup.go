package entity

// PickupKind — тип подбираемого объекта (world-integrated recruit и т.д.).
type PickupKind int

const (
	// PickupKindResource — обычный пикап (счётчик в HUD).
	PickupKindResource PickupKind = iota
	// PickupKindRecruitCamp — лагерь наёмников: требует подтверждения, затем канонический recruit в резерв.
	PickupKindRecruitCamp
)

// Pickup — подбираемый объект в мире.
type Pickup struct {
	X         int
	Y         int
	Collected bool
	Kind      PickupKind // нулевое значение = PickupKindResource
}
