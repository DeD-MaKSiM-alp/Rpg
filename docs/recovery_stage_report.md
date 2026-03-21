# Отчёт: минимальный recovery loop для `CurrentHP` (этап)

## 1. Executive summary

**Цель этапа:** дать игроку канонический способ восстанавливать persistent `Hero.CurrentHP` между боями, без отдельной camp/metagame-системы, согласованно с уже существующим пайплайном `Party → Battle → sync → Party` и пошаговым миром.

**Результат:** добавлен **отдых в explore по клавише `R`**: вызывается `party.ApplyWorldRest()`, затем **`advanceWorldTurn()`** (тот же контракт, что у «ожидания» — ход мира, враги, возможный бой). Источник истины остаётся в `party.Party.Active[].CurrentHP`.

**Где восстанавливать HP:** только в режиме **explore**, не в formation overlay и не в бою.

**Совместимость:** post-battle sync, `PlayerCombatSeeds`, старт боя, formation (F5) не менялись по контракту; добавлена одна точка мутации HP вне боя.

---

## 2. Состояние до изменений

- **`CurrentHP`** жил в `hero.Hero`, партия — в `party.Party.Active`.
- Урон и изменения HP в бою → **sync** после исхода боя обратно в партию.
- Участники с `CurrentHP == 0` не попадали в сиды боя.
- **Разрыв:** между боями не было системного действия, повышающего `CurrentHP`, кроме наград вроде +Max HP (и косвенно CurrentHP). Отряд мог уйти в спираль истощения без recovery loop.

---

## 3. Целевая модель этапа

| Вопрос | Решение |
|--------|---------|
| Где доступен recovery | Только **explore**, клавиша **`R`** |
| Кого затрагивает | Все участники с **`CurrentHP > 0`**: прибавка **`max(1, MaxHP / RestRecoveryDivisor)`**, `RestRecoveryDivisor = 4`, clamp к **MaxHP** |
| `CurrentHP == 0` | **Не воскрешаются** отдыхом; если после лечения **никто** не может сражаться, **лидер `Active[0]` получает 1 HP** (анти soft-lock) |
| World turn | **Да** — после отдыха вызывается **`advanceWorldTurn()`** (как у wait) |
| Философия | Минимальное правило, расширяемое: позже camp/cost/safe zones/consumables |

---

## 4. Изменённые файлы

| Файл | Зачем |
|------|--------|
| `internal/party/party.go` | `HasFightableMember()`, `ApplyWorldRest()`, константы `RestRecoveryDivisor`, `RestExploreBanner` |
| `internal/game/game.go` | Поля `exploreRestMsg` / `exploreRestMsgTicks`; сброс баннера при **`startBattle`** |
| `internal/game/update.go` | Тик баннера; **`R`** после F5-блока: `ApplyWorldRest` + `advanceWorldTurn()` |
| `internal/game/draw.go` | Передача текста баннера в UI explore |
| `internal/ui/formation_overlay.go` | `DrawExploreFormationHint(..., restFeedback)` — подсказки F5, R, зелёный баннер |

---

## 5. Ключевые инварианты после изменений

- Recovery меняет **канонический** `CurrentHP` в `party.Party`.
- Следующий бой берёт HP через **`PlayerCombatSeeds()`** из обновлённой партии.
- Правило **одинаково** для single-unit и multi-unit (цикл по `Active`).
- Отдых **встроен в turn-based world loop** через `advanceWorldTurn()`.
- UI отображает правило (0 HP не поднимает, затем ход мира) и краткий баннер после отдыха.

---

## 6. Что намеренно упрощено / отсутствует

- Нет стоимости ресурса, инвентаря, зелий, лагеря, банка отряда, save/load.
- Нет полноценных **revive**-правил кроме аварийного **1 HP лидеру** при полном выбытии.
- Баланс доли MaxHP (делитель 4) — рабочий placeholder, не финальный баланс.

---

## 7. Ручная проверка (чеклист)

- [ ] Старт игры → explore, подсказки F5 / R видны.
- [ ] Урон в бою → sync → explore → HP как в партии.
- [ ] **R** → HP у живых растёт (не выше MaxHP), ход мира отрабатывает (враги двигаются / возможен бой).
- [ ] Повторный вход в бой — стартовое HP совпадает с партией.
- [ ] Партия из одного героя — отдых работает.
- [ ] Несколько участников — все живые получают прибавку.
- [ ] Участник с 0 HP — отдых **не** поднимает; если все 0 — лидер получает **1 HP**.
- [ ] Несколько **R** подряд — мир двигается каждый раз.
- [ ] Formation (F5) / reorder — без регрессий; отдых не в formation.
- [ ] Retreat → explore → **R** → следующий бой — HP каноничен.

---

## 8. Follow-up (не реализовано)

- Явные правила **revive / downed** без аварийного 1 HP.
- **Camp / rest screen**, стоимость отдыха, **safe zones**.
- **Consumables**, inventory, bench/reserve.
- Углублённая **injury** / медицинский UI.
