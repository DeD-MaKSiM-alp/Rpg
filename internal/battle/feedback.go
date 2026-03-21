// feedback.go — краткоживущий визуальный слой (juice), не доменное состояние.

package battle

// Длительности в кадрах (~60 FPS).
const (
	FeedbackDamageFlashTicks = 14
	FeedbackHealFlashTicks   = 16
	FeedbackDeathFlashTicks  = 28
	FeedbackFloatTicks       = 52
	feedbackMaxFloats        = 36
)

// FeedbackFlashKind — тип вспышки на токене/карточке.
const (
	FeedbackKindDamage = 0
	FeedbackKindHeal   = 1
	FeedbackKindDeath  = 2
)

// FeedbackUnitFlash — короткая вспышка по юниту.
type FeedbackUnitFlash struct {
	Kind  int
	Ticks int
}

// FeedbackFloat — всплывающее число урона/лечения.
type FeedbackFloat struct {
	UnitID     UnitID
	Value      int
	Heal       bool
	TicksLeft  int
	TotalTicks int
}

// BattleFeedbackState — только для отрисовки; не участвует в расчёте боя.
type BattleFeedbackState struct {
	UnitFlash map[UnitID]FeedbackUnitFlash
	Floats    []FeedbackFloat
	FrameTick int // для pulse acting / общего времени
}

func maxTicksForKind(kind int) int {
	switch kind {
	case FeedbackKindDamage:
		return FeedbackDamageFlashTicks
	case FeedbackKindHeal:
		return FeedbackHealFlashTicks
	case FeedbackKindDeath:
		return FeedbackDeathFlashTicks
	default:
		return 1
	}
}

func (b *BattleContext) tickFeedback() {
	if b == nil {
		return
	}
	b.Feedback.FrameTick++
	if b.Feedback.UnitFlash != nil {
		for id := range b.Feedback.UnitFlash {
			e := b.Feedback.UnitFlash[id]
			e.Ticks--
			if e.Ticks <= 0 {
				delete(b.Feedback.UnitFlash, id)
			} else {
				b.Feedback.UnitFlash[id] = e
			}
		}
	}
	if len(b.Feedback.Floats) == 0 {
		return
	}
	dst := b.Feedback.Floats[:0]
	for _, f := range b.Feedback.Floats {
		f.TicksLeft--
		if f.TicksLeft > 0 {
			dst = append(dst, f)
		}
	}
	b.Feedback.Floats = dst
}

func (b *BattleContext) pushDamageFeedback(target UnitID, dmg int, killed bool) {
	if b == nil || target == 0 || dmg <= 0 {
		return
	}
	if b.Feedback.UnitFlash == nil {
		b.Feedback.UnitFlash = make(map[UnitID]FeedbackUnitFlash)
	}
	kind := FeedbackKindDamage
	ticks := FeedbackDamageFlashTicks
	if killed {
		kind = FeedbackKindDeath
		ticks = FeedbackDeathFlashTicks
	}
	b.Feedback.UnitFlash[target] = FeedbackUnitFlash{Kind: kind, Ticks: ticks}
	b.pushFloat(target, dmg, false)
}

func (b *BattleContext) pushHealFeedback(target UnitID, amount int) {
	if b == nil || target == 0 || amount <= 0 {
		return
	}
	if b.Feedback.UnitFlash == nil {
		b.Feedback.UnitFlash = make(map[UnitID]FeedbackUnitFlash)
	}
	b.Feedback.UnitFlash[target] = FeedbackUnitFlash{Kind: FeedbackKindHeal, Ticks: FeedbackHealFlashTicks}
	b.pushFloat(target, amount, true)
}

func (b *BattleContext) pushFloat(target UnitID, val int, heal bool) {
	if len(b.Feedback.Floats) >= feedbackMaxFloats {
		b.Feedback.Floats = b.Feedback.Floats[1:]
	}
	b.Feedback.Floats = append(b.Feedback.Floats, FeedbackFloat{
		UnitID:     target,
		Value:      val,
		Heal:       heal,
		TicksLeft:  FeedbackFloatTicks,
		TotalTicks: FeedbackFloatTicks,
	})
}

// FeedbackFlashIntensity возвращает kind и интенсивность 0..1 для оверлея; kind < 0 — нет эффекта.
func (b *BattleContext) FeedbackFlashIntensity(id UnitID) (kind int, intensity float32) {
	if b == nil || b.Feedback.UnitFlash == nil {
		return -1, 0
	}
	e, ok := b.Feedback.UnitFlash[id]
	if !ok {
		return -1, 0
	}
	mt := maxTicksForKind(e.Kind)
	if mt <= 0 {
		return e.Kind, 0
	}
	return e.Kind, float32(e.Ticks) / float32(mt)
}
