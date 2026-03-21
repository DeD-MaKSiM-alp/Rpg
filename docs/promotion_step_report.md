# Отчёт: минимальный promotion (UpgradeToUnitID)

## 1. Что изменено

- Добавлены **шаблоны tier 2** в `internal/unitdata` (`empire_warrior_squire`, `empire_archer_marksman_base`, `empire_healer_acolyte`), чтобы `UpgradeToUnitID` у tier 1 вёл на **существующую** запись реестра.
- Центральная операция **`hero.TryPromoteHero(h *Hero) error`** в `internal/hero/promotion.go`: валидация пути, пересборка героя из целевого шаблона, сохранение progression-полей, перенос HP по доле.
- Вспомогательные: **`PromotionStatusLine`**, **`PromotionErrUserMessage`**, **`preserveHPRatioOnPromotion`** (пакетный уровень, тестируется).
- **Точка входа:** экран состава (F5) → карточка (I) → клавиша **P** (вне боя). Баннер успеха/ошибки на карточке (`formationMsg`), тики как у explore-баннеров.
- **Inspect:** строка статуса повышения, подсказка в футере карточки, убран дублирующий «Следующий шаг (данные)» из блока шаблона (информация в `PromotionStatusLine`).
- **Battle identity:** без изменений пакета `battle` — после смены `Hero.UnitID` существующий `CombatUnitSeed` подхватывает новый шаблон автоматически.

## 2. Где entry point

- **Режим:** `ModeFormation`, открыта карточка (`formationInspectOpen`).
- **Клавиша:** `P`.
- **Не доступно** в explore/battle/recruit offer без экрана состава с карточкой.

## 3. Правила promotion

| Аспект | Правило |
|--------|---------|
| **UnitID** | Заменяется на `next.UnitID` из реестра по `current.UpgradeToUnitID`. |
| **Статы базовые** | `MaxHP`, `Attack`, `Defense`, `Initiative`, `HealPower` — из **нового** шаблона. |
| **Способности** | Полностью из **нового** шаблона (каноническая согласованность с template). |
| **CombatExperience** | **Сохраняется** (party-wide / progression). |
| **BasicAttackBonus** | **Сохраняется** (награды лидера). |
| **RecruitLabel** | **Сохраняется** (runtime-подпись). |
| **CurrentHP / MaxHP** | `MaxHP` из шаблона; **CurrentHP** = округлённая доля `oldHP/oldMax` на новом `MaxHP`. **0 HP** остаётся **0** (павший не поднимается promotion’ом). |
| **Валидация** | Пустой `UnitID` → отказ. Нет `UpgradeToUnitID` → отказ. Цель не в реестре → `ErrPromotionTargetMissing`. `UpgradeToUnitID == UnitID` → `ErrPromotionSelfLoop`. |

## 4. Связь с battle bridge

- Меняется только **`hero.Hero`**; следующий бой строит сиды из обновлённого героя — **`TemplateUnitID`**, tier, role и т.д. совпадают с новым шаблоном без доработок в `battle`.

## 5. Legacy

- Герой без **`UnitID`** (fallback-рекрут): **`ErrPromotionNoUnitID`**, текст в баннере и в `PromotionStatusLine`.

## 6. Ограничения текущей реализации

- Один линейный шаг по **`UpgradeToUnitID`** (нет ветвлений).
- Tier 2 шаблоны без **`UpgradeToUnitID`** — второе повышение для них недоступно.
- Нет стоимости, лагеря, квестов — только мгновенный шаг по запросу игрока (debug-friendly flow).

## 7. Следующие логичные шаги (3–5)

1. Условия promotion (ресурс, лагерь, уровень tier).
2. Ветвление (`UpgradeOptions`) и выбор в UI.
3. Заполнение tier 3+ в `unitdata` и цепочки для squire/marksman/acolyte.
4. Синхронизация отображаемого имени в бою с шаблоном (опционально).
5. Анимация/подтверждение перед сменой шаблона.
