# Inspect active target highlight — UX step report

## 1. Визуальное различие hover и active inspect target

- **Только hover** (карточка закрыта или курсор не на «открытом» юните): прежняя мягкая заливка и рамка; при открытой карточке на **другом** юните сила hover по-прежнему снижается (`InspectHoverStrength` → 0.52).
- **Только active-open** (карточка открыта, курсор не на этом юните): более плотная заливка, рамка толще (~2.35 бою), тонкая внешняя обводка `AccentStrip`, золотистая левая полоса 4px (бой) / аналог в составе.
- **Hover + active на одном target**: отдельный «комбинированный» слой — ещё плотнее заливка, рамка сильнее (`ValidTarget` / `SelectedKill` для врага в бою), двойная кромка, полоса 5px.

## 2. Бой

- `internal/game/draw.go` вызывает `DrawBattleInspectHighlights` после `DrawBattleOverlay` (вместо только hover).
- `DrawBattleInspectHighlights` (`internal/ui/inspect_hover_draw.go`) строит план через `BuildInspectBattleHighlightPlan` и рисует:
  - **combined** — один проход по ростеру и токену поля;
  - иначе **active** на `battleInspectUnitID`, затем **hover** на `inspectHoverBattleUnitID` с учётом силы.
- Подсветка дублируется на HUD-rect юнита и на battlefield token, как раньше для hover.

## 3. Состав (formation overlay)

- Перед отрисовкой строк вызывается `BuildFormationInspectHighlightPlan(hoverGlobalIdx, selected, inspectOpen)`.
- На строку: **combined** (курсор на строке с открытой карточкой), **active-open** (карточка открыта, курсор в другом месте), **nav-selected** (выбор без карточки), **hover** на других строках.
- Сохранено правило: при **закрытой** карточке и курсоре на **выбранной** строке отдельная hover-полоса не рисуется (как раньше `globalIdx != selected` для hover).

## 4. Helpers / стили

- `BuildInspectBattleHighlightPlan`, `InspectBattleHighlightPlan` — бой.
- `BuildFormationInspectHighlightPlan`, `FormationInspectHighlightPlan`, `FormationInspectHoverStrength` — состав (индексы строк, −1 = нет).
- Существующий `InspectHoverStrength` не менялся по смыслу; используется в плане боя для hover при открытой карточке на другом юните.

## 5. Изменённые файлы

| Файл | Изменение |
|------|-----------|
| `internal/ui/inspect_hover_style.go` | Планы подсветки battle + formation, `FormationInspectHoverStrength` |
| `internal/ui/inspect_hover_draw.go` | `DrawBattleInspectHighlights`, слои active/combined/hover |
| `internal/ui/formation_overlay.go` | Ветвление по плану, заливки `formationInspectActiveOpenFill` / `formationInspectCombinedFill` |
| `internal/game/draw.go` | Вызов нового draw API |
| `internal/ui/inspect_hover_style_test.go` | Тесты планов и согласованности formation/battle |

## 6. Тесты

- Только hover / только active / combined / active + hover на другом юните — для `BuildInspectBattleHighlightPlan`.
- Combined для formation; сравнение `FormationInspectHoverStrength` с `InspectHoverStrength` для одинаковых числовых id.
- Существующие тесты `InspectHoverStrength` сохранены.

## 7. Ограничения и временное

- Нет анимации pulse; статичные альфа и обводки.
- Логика индексов состава и `UnitID` в бою разделена намеренно (0 = «нет юнита» в бою vs валидный индекс строки 0 в составе).
- Документация `docs/unit_hover_highlight_step_report.md` по-прежнему упоминает старое имя `DrawBattleInspectHoverHighlight` — при желании обновить отдельным шагом.

## 8. Следующие UX-polish шаги

- Лёгкий pulse alpha только для active-open (если понадобится), без общей анимационной системы.
- Подсветка «inspect доступен» при наведении на мёртвого/неинтерактивного юнита — если появится отдельное правило геймдизайна.
- Обновить устаревшие упоминания API в старых отчётах.
