# Отчёт: минимальный слой данных юнита (unit templates + `UnitID`)

## 1. Что изменено

- В доменную модель **`hero.Hero`** добавлено поле **`UnitID string`** — стабильный идентификатор шаблона из реестра (`internal/unitdata`), не отображаемое имя и не UI-ярлык.
- Добавлен компактный пакет **`internal/unitdata`**: структура **`UnitTemplate`**, in-code **`map[string]UnitTemplate`**, функции **`GetUnitTemplate`**, **`MustGetUnitTemplate`**, **`EarlyRecruitUnitIDs`**, вспомогательные подписи для inspect (**`FactionDisplayRU`**, **`LineDisplayRU`**, **`AttackKindDisplayRU`**).
- Стартовый лидер и рекрут создаются через шаблоны: **`hero.NewHeroFromUnitTemplate`**, **`hero.RecruitHeroFromEarlyPool`** (циклический пул из 3 ранних шаблонов Империи).
- **Inspect** (карточка бойца) читает канонические поля шаблона по `UnitID` и показывает fallback, если id пустой или неизвестен.
- **Боевой контур** не переведён на шаблоны: **`CombatUnitSeed`** по-прежнему берёт статы и способности из runtime-героя. Дополнительно выставляются **`Def.Role`** и **`Def.IsRanged`** по **набору способностей** героя (лучник/целитель/боец), чтобы токены и дальность не расходились с новыми рекрутами.
- Добавлены **unit-тесты** в `internal/unitdata` и `internal/hero`.

## 2. Новые структуры / файлы / точки входа

| Файл | Назначение |
|------|------------|
| `internal/unitdata/unitdata.go` | Реестр шаблонов, константы `UnitID`, `EarlyRecruitUnitIDs()`, lookup |
| `internal/unitdata/unitdata_test.go` | Тесты реестра и пула рекрута |
| `internal/hero/hero.go` | `UnitID`, `NewHeroFromUnitTemplate`, `DefaultHero` через шаблон, доработка `CombatUnitSeed` |
| `internal/hero/recruit.go` | `RecruitHeroFromEarlyPool`, `NewRecruitHero` (совместимость), LEGACY fallback без `UnitID` |
| `internal/hero/hero_test.go` | Тесты фабрики, пула, сида |
| `internal/game/update.go` | F9: `RecruitHeroFromEarlyPool` + `RecruitLabel` |
| `internal/game/recruit_offer.go` | Лагерь: тот же пул |
| `internal/ui/character_inspect.go` | Заголовок из шаблона, блок «Шаблон», fallback |

### Стартовый набор `unit_id` (Империя, ранний этап)

- `empire_militia_spearman_t1` — стартовый лидер (милития · копейщик).
- `empire_warrior_recruit` — пехотный новобранец.
- `empire_archer_recruit` — рекрут-лучник.
- `empire_healer_novice` — послушник (хил).

## 3. Путь данных: template → hero → party → inspect / recruit

1. **Шаблон** задаётся в `unitdata` (identity + стартовые статы + способности).
2. **`hero.NewHeroFromUnitTemplate`** копирует в **`Hero`** поля runtime-состояния и **`UnitID`**.
3. **`party.Party`** хранит героев как раньше; **`PlayerCombatSeeds`** не менялся по сигнатуре.
4. **Recruit (F9 / лагерь)** выбирает шаблон из **`EarlyRecruitUnitIDs()`** по формуле `(recruitSerial-1) % len(pool)`; **`RecruitLabel`** остаётся «Новобранец N» для подписи.
5. **Inspect** по **`hero.UnitID`** подтягивает **`UnitTemplate`**; при отсутствии — текст про legacy.

## 4. Сознательно не делалось на этом шаге

- Внешние JSON/YAML, save/load, региональные таблицы рекрута, weighted random.
- Глубокая интеграция шаблонов в **`battle`** (кроме вывода **роли/IsRanged** из способностей героя).
- Полный перенос design-schema из markdown в код.
- Миграция **`ArchetypeID`** в боевом сиде (по-прежнему `"player"` в **`BuildPlayerCombatSeed`**).

## 5. Legacy / временные пути

- **`DefaultHero`**: при невозможности прочитать шаблон — статический fallback **без** `UnitID` (комментарий LEGACY в коде).
- **`recruitHeroFallbackNoTemplate`**: тот же запасной путь без `UnitID`, если registry недоступен.
- **Inspect**: неизвестный или пустой `UnitID` — явное сообщение, дальше только поля героя.
- **`RecruitDisplayName`**: сохранён для баннеров и подзаголовка; заголовок карточки берётся из **DisplayName шаблона**, если шаблон найден.

## 6. Логичные следующие 3–5 шагов

1. Пробросить **`UnitID`** (или `ArchetypeID`) в **`CombatUnitSeed.Def`** для отладки/логов без дублирования статов.
2. Расширить пул рекрута по биому/источнику (всё ещё in-code).
3. Связать progression tier с **`UnitTemplate.Tier`** и эволюцией `unit_id`.
4. Убрать дублирование «роль из способностей» vs «роль из шаблона», когда появится единый источник правды в hero.
5. Внешние данные (опционально) — выгрузка того же registry в файл позже.

## 7. Риски и спорные места

- **Две правды о роли**: inspect использует шаблон; в бою **`Def.Role`** выводится из **способностей**. При ручной правке абилок без смены шаблона возможен рассинхрон (пока маловероятен).
- **`effectiveRange` / `IsRanged`**: для лучника выставляется **`Def.IsRanged`**, что влияет на общую логику дальности; целитель остаётся с **`IsRanged == false`** (как и раньше для игрока).
- Пул рекрута **циклический**, не случайный — предсказуемо для демо.
