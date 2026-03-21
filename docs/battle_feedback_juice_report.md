# Отчёт: battle feedback / combat juice (минимальный слой)

## 1. Executive summary

**Цель:** дать бою кратковременный визуальный отклик на урон, лечение, смерть и усилить ощущение активного юнита, без VFX-пайплайна и без изменения доменной логики.

**Реализовано:**
- **`BattleFeedbackState`** в `BattleContext`: вспышки по юнитам (`UnitFlash`), всплывающие числа (`Floats`), счётчик кадров `FrameTick` для pulse.
- **Запуск** из **`ApplyActionResult`** после `ResolveAbility` (единая точка после факта урона/лечения).
- **Тик** `tickFeedback()` в начале **`Battle.Update`** (включая паузу после действия).
- **Отрисовка:** вспышки на **токенах** и **карточках v2**, оверлей на **слотах v1**, **floating numbers** (v1 и v2), **пульсирующее кольцо** acting на токене.

**Совместимость:** HP и исход боя не менялись; feedback только визуальный.

---

## 2. Состояние до изменений

- Урон/лечение отражались в логе и числах HP без «сочувствующей» анимации.
- Токены и карточки не реагировали на событие попадания.
- Активный юнит был обозначен статичным кольцом.

---

## 3. Целевая модель этапа

| Эффект | Реализация |
|--------|------------|
| Урон | Вспышка `FeedbackDamageOverlay` + число `-%d` красным |
| Лечение | Вспышка `FeedbackHealOverlay` + `+%d` зелёным |
| Смерть | Вспышка `FeedbackDeathOverlay` (заменяет damage flash при kill) |
| Acting | `sin(FrameTick)` масштабирует радиус внешнего кольца на токене |
| Floats | Привязка к центру **BattlefieldTokens** или **UnitRects** |

**Не входит:** miss/invalid target, shake, частицы, отдельный таймлайн анимаций.

---

## 4. Изменённые файлы

| Файл | Содержание |
|------|------------|
| `internal/battle/feedback.go` | Типы, константы тиков, `tickFeedback`, `pushDamageFeedback`, `pushHealFeedback`, `FeedbackFlashIntensity`. |
| `internal/battle/battle.go` | Поле `Feedback`, вызовы push в `ApplyActionResult`. |
| `internal/battle/update.go` | `b.tickFeedback()` в начале `Update`. |
| `internal/ui/theme.go` | Цвета оверлеев `FeedbackDamage/Heal/DeathOverlay`. |
| `internal/ui/battle_feedback_draw.go` | `drawFeedbackOverlayRect`, `DrawBattleFeedbackFloats`. |
| `internal/ui/battlefield_tokens.go` | Вспышки на круге, pulse кольца, вызов floats. |
| `internal/ui/battle_panels.go` | Оверлей на v2 карточках и v1 слотах; floats в конце `drawBattleOverlayText`. |

---

## 5. Инварианты

- Источник событий — **`ApplyActionResult`** (после `ResolveAbility`).
- `Feedback` не участвует в расчёте урона/победы.
- `tickFeedback` не меняет `TurnOrder`, `Phase`, `Units` HP.

---

## 6. Упрощения / follow-up

- Нет отдельной очереди «событий для UI» — только push из `ApplyActionResult`.
- Текст float использует общий HUD face; альфа через цвет.
- **Follow-up:** impact VFX, звук, отдельный пул pop-up, polish смерти.

---

## 7. Ручная проверка

- [ ] Удар по врагу: вспышка + число на токене и карточке.
- [ ] Heal: зелёная вспышка + `+N`.
- [ ] Смерть: тёмная вспышка, крест на токене как раньше.
- [ ] Ход союзника: пульс кольца на токене.
- [ ] v1 (F8): вспышка на слоте + floats в overlay.
- [ ] v2: floats только из `DrawBattlefieldV2Scene` + v2 roster overlay.
- [ ] Нет дублирования floats в v2 (только из сцены — v1 из `drawBattleOverlayText`).

---

## 8. Follow-up

- Лёгкие impact-блики, звуки.
- Синхронизация с длительностью `PhaseActionPause`.
- Отдельные пресеты для врагов/архетипов.
