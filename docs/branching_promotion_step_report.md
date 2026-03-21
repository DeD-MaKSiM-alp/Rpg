# Отчёт: минимальный шаг branching promotion (tier 2 → два tier 3)

## 1. Какие линии получили ветвление

- **Воинская:** с `empire_warrior_squire` доступны две цели — **танк** (`empire_warrior_tank_1`) и **ДД** (`empire_warrior_dd_1`). Соответствует design-doc (`empire_roles_and_units.md`: `upgrade_to` у оруженосца).
- **Линия хилов:** с `empire_healer_acolyte` — **одиночный хил** (`empire_healer_single_1`) и **групповая поддержка** (`empire_healer_group_1`).
- **Стрелковая:** остаётся **линейной** (tier 2 → один tier 3), без изменения UX для лучника.

## 2. Какие новые шаблоны добавлены

- `empire_warrior_tank_1` — танковый профиль (выше HP/defense, ниже attack относительно ДД-ветки), те же способности, что позволяет текущий движок без новых боевых правил.
- `empire_healer_group_1` — отдельная identity/stats; набор способностов на базе уже существующих heal-паттернов (без новых боевых систем).

## 3. Как хранится promotion path в данных

- Поле **`UpgradeToUnitID`** — прежний линейный путь (одна цель).
- Новое поле **`UpgradeOptions []string`** — явный список из **двух** целей для ветвления.
- Контракт нормализации: **`unitdata.PromotionTargetUnitIDs(t)`** — если `UpgradeOptions` непустой, берётся он; иначе, если задан `UpgradeToUnitID`, возвращается один элемент; иначе пусто.
- Старые шаблоны только с `UpgradeToUnitID` **не трогались** по смыслу.

## 4. Выбор ветки в UI

- В режиме состава / inspect (**I**): при **двух** целях индекс ветки `formationPromoteBranchIdx` изначально **−1** (ветка не выбрана).
- **← / →** задают **0** или **1** по порядку из `PromotionTargetUnitIDs`.
- **P** вызывает gate с выбранной целью; пока ветка не выбрана, gate отклоняет с сообщением про выбор ветки.
- Карточка (`DrawCharacterInspectOverlay`) показывает обе ветки, стоимость по каждой (tier цели), краткие подписи.

## 5. Домен: допустимость target

- **`TryPromoteHero`** — только если ровно **одна** цель; при двух и более — `ErrPromotionBranchChoiceRequired`.
- **`TryPromoteHeroTo(h, targetUnitID)`** — применяет promotion только если `targetUnitID` входит в нормализованный список текущего шаблона; иначе `ErrPromotionTargetNotAllowed`.
- **`PromotionTargetUnitIDs(h)`** / **`ValidatePromotionPathsExist`** — единая точка для списка и проверки «есть ли шаг».

## 6. Policy / gate и выбранная ветка

- **`EvaluatePromotionGate(h, atCamp, trainingMarks, selectedTargetUnitID)`**:
  - при **одной** цели `selectedTargetUnitID` может быть `""` — используется единственная цель;
  - при **двух** целях пустой `selectedTargetUnitID` → отказ («сначала выберите ветку»);
  - стоимость считается от **tier выбранной** цели (`promotionCostForTargetUnitID` / `PromotionTrainingMarkCostForHeroTarget`).

## 7. Тесты

- `internal/unitdata`: регистрация tier 3, ветки squire/acolyte, `PromotionTargetUnitIDs`.
- `internal/hero`: ветвление, `TryPromoteHeroTo` (танк), отказ для чужого `UnitID`, `CombatUnitSeed`, сценарий «нет пути после tier 3».
- `internal/game`: gate при двух ветках без выбора, исправленные сценарии стоимости после смены warrior-пути на ветвление.

## 8. Сознательно не делалось

- Tier 4+, универсальные деревья, новые боевые механики, перепись battle-слоя.
- Разная стоимость веток при одном tier (обе tier 3 — одна цена в знаках).
- Полноэкранный UI выбора класса; тяжёлые E2E-тесты UI.

## 9. Следующие логичные шаги (3–5)

1. Ветвление **лучника** (если появится второй tier-3 шаблон без новых механик).
2. Реальные **различия heal-паттернов** между `single_1` и `group_1`, когда в battle будут готовы массовые хилы.
3. **Tier 4** для танка/ДД по уже намеченным `upgrade_to` в design-doc.
4. Локализация/копирайт имён веток в карточке.
5. Лёгкий баланс-пасс по веткам после игрового теста.
