# Отчёт: tier 3 шаблоны Империи

## 1. Какие tier 3 шаблоны добавлены

| `unit_id` | Имя | Линия | Источник в design-doc |
|-----------|-----|-------|------------------------|
| `empire_warrior_dd_1` | Мечник | warrior | ветка ДД, tier 3 |
| `empire_archer_pure_1` | Лучник | archer | ветка «чистого» лучника, tier 3 |
| `empire_healer_single_1` | Целитель | healer | ветка одиночного хила, tier 3 |

У каждого: `Tier: 3`, `UpgradeToUnitID` пустой (пока верх ступени в коде).

## 2. Почему выбраны именно они

- Совпадают с уже зафиксированными в [`empire_roles_and_units.md`](empire_roles_and_units.md) идентификаторами и ролями.
- Одна линия — одна ветка без **UpgradeOptions**: у tier 2 задан **один** следующий шаг (док допускает две ветки; в коде выбрана одна на линию для линейного пути 1→2→3).

## 3. Какие линии теперь доходят до tier 3

- **warrior:** `empire_warrior_squire` → `empire_warrior_dd_1` (милиция/рекрут по-прежнему ведут в оруженосца, затем тот же tier 3).
- **archer:** `empire_archer_marksman_base` → `empire_archer_pure_1`.
- **healer:** `empire_healer_acolyte` → `empire_healer_single_1`.

## 4. Inspect и promotion

- Следующий шаг берётся из `UpgradeToUnitID` — для tier 2 отображается tier 3 id; для tier 3 — доменная проверка даёт «нет следующего шага».
- Цена promotion по-прежнему **tier цели** (`PromotionCostFromTargetTier`): tier 3 → **3** знака; без изменений gameplay-слоя.

## 5. Battle identity bridge

- `hero.CombatUnitSeed()` тянет `TemplateUnitID` и tier из текущего `Hero.UnitID` / шаблона; новые id просто проходят через существующий мост.
- Проверено тестом `TestCombatUnitSeed_AfterTier3Promotion` (хил: novice → acolyte → целитель).

## 6. Тесты

| Файл | Что добавлено/обновлено |
|------|-------------------------|
| `internal/unitdata/unitdata_test.go` | `TestTier3Templates_registeredAndUpgradePath` |
| `internal/hero/promotion_test.go` | `TestTryPromoteHero_ThirdPromotionNoPath`, `TestCombatUnitSeed_AfterTier3Promotion` |
| `internal/game/promotion_gate_test.go` | `noPathEvenAtCamp` — два promote до tier 3 |
| `internal/game/promotion_tier_cost_test.go` | `TestPromotionTrainingMarkCostForHero_tier3Target`, правка «нет цены» после двух promote |

## 7. Сознательно не делалось

- Вторая ветка на tier 2 (танк / яд / массовый хил) — не вводились дополнительные tier 3.
- Новые `AbilityID`, баланс наград за бой, правки `TrainingMarksPerVictory`, изменения recruit pool / лагерей.
- Большой balance pass по статам — только умеренный шаг от tier 2.

## 8. Временное / follow-up

- Темп «+1 знак за бой» vs цена **3** за tier 3 — при необходимости отдельная подстройка награды или формулы цены (не делалось в этом шаге).
- Альтернативные tier 3 ветки из дока — когда появится ветвление promotion.

## 9. Следующие логичные шаги (3–5)

1. Вторые ветки tier 3 (танк / poison archer / group healer) или система выбора ветки.
2. Tier 4 заготовки под теми же id-паттернами.
3. Тонкая настройка статов tier 3 после плейтеста.
4. Синхронизация краткой выжимки `empire_roles_and_units.md` с тем, что в коде выбрана одна ветка на линию.

## Путь к отчёту

`docs/tier3_templates_step_report.md`
