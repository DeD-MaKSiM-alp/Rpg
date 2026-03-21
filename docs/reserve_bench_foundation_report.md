# Reserve / Bench foundation — инженерный отчёт

## 1. Executive summary

- **Цель:** канонически разделить отряд на **Active** (в бою) и **Reserve/Bench** (вне боя), не ломая цепочку `Party → PlayerCombatSeeds → Battle` и существующий battle vertical slice.
- **Результат:** в `party.Party` добавлено поле **`Reserve []hero.Hero`**; **`PlayerCombatSeeds()`** по-прежнему строится **только из `Active`**; лидер — **`Active[0]`**; добавлены **`MoveActiveToReserve` / `MoveReserveToActive`** с ограничениями; **отдых** (`ApplyWorldRest`) действует на **Active + Reserve**; **formation overlay (F5)** показывает оба списка, **Enter** переносит между строем и резервом; explore strip показывает счётчик резерва.
- **Совместимость:** бой, post-battle sync по `PartyActiveIndex`, награды лидеру, reorder внутри Active — сохранены; battle не знает о Reserve.

## 2. Состояние до изменений

- `Party` = только **`Active`**; весь ростер считался «в строю».
- `PlayerCombatSeeds`, formation, sync, `Leader()` — завязаны на **`Active`**.
- Ограничение: нельзя было моделировать скамейку без дублирования данных вне `Party`.

## 3. Целевая модель этапа

| Элемент | Смысл |
|--------|--------|
| **Active** | Участники боя; порядок = formation / `PlayerSlotForPartyIndex`; в бою не больше **`MaxActiveBattleSlots` (6)**. |
| **Reserve** | Не попадают в сиды; порядок — список скамейки (без боевых слотов). |
| **Лидер** | Всегда **`Active[0]`**; нельзя опустошить Active (минимум 1 в строю). |
| **Перенос** | Active→Reserve: если в строю ≥2. Reserve→Active: если в строю < 6. |
| **HP 0** | Перенос разрешён (игрок может убрать «труп» в резерв или наоборот); в бой по-прежнему только `CanEnterBattle()`. |

Источник истины — **один**: структура `Party` в `Game`.

## 4. Изменённые файлы

| Файл | Изменения |
|------|-----------|
| `internal/party/party.go` | Поле `Reserve`; `MoveActiveToReserve`, `MoveReserveToActive`, `ActiveCount`/`ReserveCount`, `MaxActiveBattleSlots`; `ApplyWorldRest` на оба списка; `TwoMemberDemoParty` = 2 active + 1 reserve. |
| `internal/ui/formation_overlay.go` | Два блока UI, глобальный индекс выбора, подсказка про Enter; предупреждение при полном строю. |
| `internal/ui/explore_hud.go` | Заголовок с числом резерва; строка «Резерв не в бою»; подсказка F5. |
| `internal/game/update.go` | `updateFormationMode`: навигация по `na+nr`, reorder только по active; **Enter** для обмена. |
| `internal/game/game.go` | Комментарии к `ModeFormation` / `formationSel`. |

**Не менялись:** `sync_party.go`, `BuildBattleContext`, `postbattle`, `PlayerCombatSeeds` логика индексации (индекс только по `Active`).

## 5. Инварианты

- В бой попадают **только** герои из **`Active`** с `CanEnterBattle()`.
- **`PartyActiveIndex`** — индекс в **`party.Active`** после боя.
- **`len(Active) >= 1`** при нормальном использовании API переносов.
- Reserve **не** участвует в симуляции боя.

## 6. Упрощения и ограничения этапа

- Нет найма, удаления героя, отдельного лагеря, save/load.
- Нет отдельных правил recovery **только** для резерва — отдых мировой одинаков для обоих списков.
- Нет drag-and-drop; только клавиатура в formation overlay.
- Демо с 2+1: **`TwoMemberDemoParty()`** — вручную в коде при отладке; **`NewGame`** по-прежнему **`DefaultParty()`**.

## 7. Ручная проверка (checklist)

- [ ] Старт игры, explore, полоска отряда с резервом = 0.
- [ ] Подставить `TwoMemberDemoParty` в `NewGame`: F5 — видны строй и резерв; Enter переносит туда-обратно.
- [ ] Reorder ←→ только на строках строя (≥2).
- [ ] Заполнить строй до 6 из резерва — дальше перенос блокируется.
- [ ] Бой после смены состава — только active в бою; post-battle HP синхронизируется.
- [ ] R в explore — HP растёт у active и reserve.
- [ ] Один юнит в строю — Enter в резерв не срабатывает.

## 8. Follow-up

- Найм / увольнение, лимит размера партии, отдельные правила отдыха для резерва.
- Лидер: обязательность в строю / смена лидера без reorder.
- Save/load состава; UI мышью для formation.
- Прогрессия не-лидеров, мета-интеграция мира.
