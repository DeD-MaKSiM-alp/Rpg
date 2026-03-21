# Battle ACTION / summary labels — RU polish step report

## 1. Какие англоязычные panel labels оставались

- Заголовки панелей v1: `ABILITIES`, `ACTION`, `COMBAT LOG`, подписи построения `PLAYER` / `ENEMY`, ряды `FRONT` / `BACK`, слоты `EMPTY` / `DEAD` / `HP`.
- Плейсхолдеры: `(waiting)`, `(enemy turn)`.
- Сводка ACTION: `Step:`, `Target:`, `Preview: dmg/heal` (через `pt.PhaseString()` — англ. `ChooseAbility` и т.д.).
- Цели: `self`, `none`, `unit #`.
- Строки состояния: `Раунд N · фаза:` + `PhaseString()` боя; вторая строка с `PlayerTurn.PhaseString()`.
- Верхняя строка v2: мёртвый код с `R%d · PhaseString()` (не показывался, но оставался английский след).
- В v2 цель: `self` для TargetKindSelf.

## 2. Что переведено

| Было | Стало |
|------|--------|
| ABILITIES | СПОСОБНОСТИ |
| ACTION | ХОД |
| COMBAT LOG | ЛОГ БОЯ |
| PLAYER / ENEMY | СОЮЗНИКИ / ВРАГИ |
| FRONT / BACK | ПЕРЕД / ЗАД |
| EMPTY / DEAD / HP | ПУСТО / погиб / ОЗ |
| (waiting) | (ожидание) |
| (enemy turn) | (ход врага) |
| Step: | Шаг: |
| Target: | Цель: |
| Preview: dmg / heal | Вид: урон / Вид: лечение |
| self / none / unit # | себя / нет / юнит # |
| R%d · … (v2) | Раунд %d · … |

Фазы боя и подфазы хода игрока для HUD вынесены в **`BattleContext.PhaseLabelRU()`** и **`PlayerTurnState.PhaseLabelRU()`** (короткие русские строки вместо `PhaseString()` там, где текст виден игроку).

## 3. Стиль

- Короткие слова, без канцелярита; в духе уже существующих подписей способностей.
- `Debug`-строки `PhaseString()` в пакете `battle` **не удалялись** — для отладки/логов разработчика.

## 4. Файлы

| Файл | Изменение |
|------|-----------|
| `internal/battle/battle.go` | `PhaseLabelRU()` |
| `internal/battle/player_turn.go` | `PhaseLabelRU()` для `PlayerTurnState` |
| `internal/battle/phase_labels_ru_test.go` | тесты фаз |
| `internal/ui/battle_panels.go` | переводы v1/v2, `battleActionTargetLabelRU`, top bar без мёртвого R/англ. фазы |
| `internal/ui/battle_panels_labels_test.go` | тесты цели и строки «Шаг» |

## 5. Тесты

- Боевая фаза и подфаза хода дают ожидаемые русские короткие метки.
- `battleActionTargetLabelRU`: себя / нет.
- Строка вида `Шаг: …` не содержит `Step`/`Choose`.

## 6. Сознательно не делалось

- Не трогали `PlayerTurnStatusString`, `FormationSummary`, `ActiveUnitTeamName` и другие англ. строки, ориентированные на debug/диагностику.
- Не трогали `PlayerAbilityLabelRU` и реестр способностей.
- Inspect-карточка не менялась.

## 7. Следующие micro-polish шаги

- При желании перевести оставшиеся **debug-only** строки в `battle` или отдельной сборкой.
- Унифицировать единый стиль «Раунд» в v1 info row (уже `PhaseLabelRU`).
