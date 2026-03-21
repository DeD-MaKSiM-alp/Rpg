package unitdata

// PromotionTargetUnitIDs — нормализованный список целей повышения: UpgradeOptions или один UpgradeToUnitID.
func PromotionTargetUnitIDs(t UnitTemplate) []string {
	if len(t.UpgradeOptions) > 0 {
		out := make([]string, len(t.UpgradeOptions))
		copy(out, t.UpgradeOptions)
		return out
	}
	if t.UpgradeToUnitID != "" {
		return []string{t.UpgradeToUnitID}
	}
	return nil
}
