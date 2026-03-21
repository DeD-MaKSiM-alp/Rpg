# Отчёт: hover-подсветка под курсором (бой и состав)

## 1. Где работает hover / highlight

- **Бой (до post-battle):** каждый кадр по позиции мыши вызывается `HitTestUnitUnderCursor`; для попавшего `UnitID` поверх HUD рисуется полупрозрачная подложка и контур на **ростерной карточке** и **токене поля** (если оба есть в layout).
- **Состав (F5):** каждый кадр `FormationHitTestGlobalIndex`; подсвечивается **строка** под курсором (если это не выбранная строка клавиатурой).

## 2. Стиль подсветки

- **Бой:** мягкая заливка (ally — сине-стальная тон, enemy — красноватая) + тонкая рамка `HoverTarget` / `EnemyAccent`; сила `InspectHoverStrength` уменьшается (~0.52), если открыта карточка inspect **по другому** юниту.
- **Состав:** фон строки `formationHoverFill`, рамка `HoverTarget`, тонкая **золотая вертикальная полоса** слева (`AccentStrip`); при открытой карточке и наведении на **другую** строку — такое же ослабление (~0.52), как в бою.

## 3. Hover vs открытый inspect

- **Hover** — текущий юнит/строка под курсором, обновляется каждый кадр из hit-test.
- **Открытая карточка** — `battleInspectUnitID` / выбор в составе; не смешивается с hover-логикой.
- Если карточка открыта по юниту A, а курсор над юнитом B — подсветка B **мягче**, чтобы не конкурировать с фокусом карточки на A.

## 4. Состояние (Game)

| Поле | Значение |
|------|----------|
| `inspectHoverBattleUnitID` | `battle.UnitID`, 0 = нет |
| `inspectHoverFormationGlobalIdx` | индекс строки, -1 = нет |

Очистка: в начале `Update` сброс formation-hover при `mode != Formation`, battle-hover при `mode != ModeBattle`; при post-battle battle-hover = 0; в `endBattle` и `startBattle` battle-hover = 0; `NewGame`: `inspectHoverFormationGlobalIdx = -1`.

## 5. Изменённые файлы

- `internal/game/game.go` — поля, сброс в `endBattle` / `startBattle`
- `internal/game/update.go` — очистка по режиму, hit-test в `updateBattleMode` / `updateFormationMode`
- `internal/game/draw.go` — `DrawBattleInspectHoverHighlight` после `DrawBattleOverlay`; formation с hover
- `internal/ui/inspect_hover_style.go` — `InspectHoverStrength`
- `internal/ui/inspect_hover_draw.go` — `DrawBattleInspectHoverHighlight`
- `internal/ui/formation_overlay.go` — параметр `hoverGlobalIdx`, стиль hover-строки
- `internal/ui/inspect_hover_style_test.go` — тесты силы подсветки
- `docs/unit_hover_highlight_step_report.md` — этот отчёт

Подсказки в HUD **не** дублировались: подсветка считается достаточной.

## 6. Тесты

- `InspectHoverStrength`: нулевой hover; ослабление при открытой карточке на другом юните; полная сила при совпадении с открытым или при закрытой карточке.

## 7. Ограничения

- Подсветка не рисуется в post-battle flow.
- v1 и v2 battle layout оба используют `ComputeBattleHUDLayout` — hit-test и draw совпадают.

## 8. Логичные следующие UX-шаги

1. Лёгкая анимация пульсации альфы (не обязательно).
2. Тот же hover для других кликабельных зон боя (кнопка «Назад») — отдельный hit-test.
3. Звук hover (опционально).
4. Подсветка «inspect target» чуть сильнее при открытой карточке на том же юните (тонкая вторая обводка).
5. Синхронизация цветов с темой при смене палитры.
