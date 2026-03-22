package ui

// ExploreHUDTextPolicy — централизованные отступы и лимиты explore HUD (player-facing).
// Числа согласованы с прежней отрисовкой: верхний текст начинался с TopHUD+8, низ — inset 14 от края.
type ExploreHUDTextPolicy struct {
	Tier ResolutionTier

	// Верхний HUD: внешний отступ контента от TopHUD (раньше фиксированные +8).
	TopContentPadX float32
	TopContentPadY float32
	// TopBackgroundPad — внутренний отступ подложки от края TopHUD (заливка от TopHUD.Y).
	TopBackgroundPad float32
	// TopLineGapExtra — добавка к Preset.LineH для шага строки (раньше +2 к LineH).
	TopLineGapExtra float32
	// TopStatusHeaderH — полоса заголовка «статусной» панели (player-facing).
	TopStatusHeaderH float32

	// Нижняя панель: горизонтальный inset строк (раньше translate 14).
	BottomTextInsetX float32
	BottomChromePad  float32

	// TransientBannerPadX — горизонтальный отступ текста в полосе над низом.
	TransientBannerPadX float32

	// Party strip (левая колонка): плотность и зазоры.
	PartyStripPad            float32
	PartyStripPadSmall       float32
	PartyStripBaseLineH      float32
	PartyRowGap              float32
	PartyReserveExtraGap     float32
	PartyTitleToBodyGap      float32
	PartyLeaderGap           float32
	PartyMinLineScale        float32
	PartyMinLineH            float32

	// Лимит строк нижней полосы по tier (дублирует preset.BottomMaxLines для явной ссылки в коде).
	MaxBottomLines int
}

// ExploreHUDTextPolicyForTier возвращает политику для tier (привязка к ScreenLayoutPreset).
func ExploreHUDTextPolicyForTier(tier ResolutionTier) ExploreHUDTextPolicy {
	p := presetForTier(tier)
	padX := float32(8)
	padY := float32(8)
	if tier == TierSmall {
		padX = 8
		padY = 8
	}
	return ExploreHUDTextPolicy{
		Tier:               tier,
		TopContentPadX:     padX,
		TopContentPadY:     padY,
		TopBackgroundPad:   8,
		TopLineGapExtra:    2,
		TopStatusHeaderH:   20,
		BottomTextInsetX:   14,
		BottomChromePad:    p.BottomChromePad,
		TransientBannerPadX: 12,
		PartyStripPad:            8,
		PartyStripPadSmall:       6,
		PartyStripBaseLineH:      p.LineH,
		PartyRowGap:              6,
		PartyReserveExtraGap:     4,
		PartyTitleToBodyGap:      4,
		PartyLeaderGap:           2,
		PartyMinLineScale:        0.5,
		PartyMinLineH:            11,
		MaxBottomLines:           p.BottomMaxLines,
	}
}

// --- Категории строк нижней панели и порядок вытеснения (overflow) ---

// ExploreBottomLineKind — семантическая категория строки нижней панели explore.
// Порядок констант не задаёт порядок отрисовки (см. PlanExploreBottomLines).
type ExploreBottomLineKind int

const (
	BottomKindZone ExploreBottomLineKind = iota
	BottomKindInteraction
	BottomKindHotkeys // R · F5 · F9 (одна строка)
	BottomKindBannerRest
	BottomKindBannerRecruit
	BottomKindBannerPOI
)

// ExploreBottomEvictionPriority — порядок вытеснения при нехватке MaxBottomLines (сначала выкидывается менее приоритетное).
// Должен оставаться синхронизирован с ApplyExploreHintOverflow (text_policy.go).
func ExploreBottomEvictionPriority() []ExploreBottomLineKind {
	return []ExploreBottomLineKind{
		BottomKindBannerPOI,
		BottomKindBannerRecruit,
		BottomKindBannerRest,
		BottomKindInteraction,
		BottomKindZone,
	}
}

// ExploreBottomBaselineSlotCount — сколько строк «базовых» подсказок внизу (сейчас одна объединённая строка хоткеев).
func ExploreBottomBaselineSlotCount() int { return 1 }
