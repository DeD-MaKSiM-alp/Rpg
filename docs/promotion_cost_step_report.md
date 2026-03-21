# Отчёт: стоимость promotion (знаки обучения)

## 1. Какой ресурс выбран

**Знаки обучения** (`TrainingMarks`) — одно целочисленное поле на сессию в структуре `Game`.

## 2. Почему именно он

- В проекте не было отдельной валюты; `BattlesWon` используется для эскалации боёв, не как «кошелёк».
- Новый счётчик на `Game` — минимальное место без инвентаря и без изменения `hero` / `TryPromoteHero`.
- Название понятно в UI: связь с «обучением» в лагере и повышением ранга.

## 3. Где хранится

- `internal/game/game.go`: поле `TrainingMarks int` (сессия, сбрасывается при новой игре через `NewGame`).

## 4. Как начисляется

- Константа `TrainingMarksPerVictory = 1` в `internal/game/promotion_gate.go`.
- `Game.applyVictoryTrainingMarks()` в `internal/game/victory_marks.go` увеличивает счётчик.
- Вызов из `internal/game/update.go` в ветке `BattleOutcomeVictory` сразу после `BattlesWon++` (без изменения `resolveBattleResult` и без переписывания post-battle).

## 5. Как считается стоимость promotion

- Фиксированная цена: `PromotionCostTrainingMarks = 2` (одинакова для всех promotion в этом шаге).
- Списание только после успешного `hero.TryPromoteHero` в `updateFormationMode` (ветка по клавише **P**).

## 6. Policy availability

Единая точка: `EvaluatePromotionGate(h *hero.Hero, atCamp bool, trainingMarks int)`.

Порядок проверок:

1. `hero.ValidatePromotionDomain(h)` — legacy / нет пути / шаблон и т.д.
2. `atCamp` — стоимость на активном `PickupKindRecruitCamp`.
3. `trainingMarks >= PromotionCostTrainingMarks` — иначе сообщение вида «не хватает знаков обучения (N/M)».

## 7. Почему `TryPromoteHero` остался чистым

- В `internal/hero/promotion.go` нет изменений: по-прежнему только домен шаблона и пересборка героя.
- Ресурс, лагерь и списание не импортируются в пакет `hero`.

## 8. Изменённые файлы

| Файл | Назначение |
|------|------------|
| `internal/game/game.go` | Поле `TrainingMarks` |
| `internal/game/promotion_gate.go` | Константы, расширенный `EvaluatePromotionGate` |
| `internal/game/victory_marks.go` | Начисление за победу |
| `internal/game/update.go` | Вызов начисления; gate + списание при promotion |
| `internal/game/draw.go` | Передача знаков и цены в HUD и inspect |
| `internal/game/promotion_gate_test.go` | Обновление вызовов + тест нехватки знаков |
| `internal/game/promotion_cost_test.go` | Тесты начисления и списания |
| `internal/ui/draw.go`, `internal/ui/panels.go` | HUD: строка про знаки |
| `internal/ui/character_inspect.go` | Строка статуса promotion с ценой и запасом |

## 9. Какие тесты добавлены

- `TestEvaluatePromotionGate_insufficientMarks`
- `TestApplyVictoryTrainingMarks`
- `TestPromotionSuccessDeductsMarks`
- `TestPromotionGateBlocksWithInsufficientMarks_NoDeductSimulated`

Доменные тесты `hero` (`PromotionUILine`, `TryPromoteHero`) без изменений.

## 10. Следующие логичные шаги (3–5)

1. Баланс: другие `TrainingMarksPerVictory` / `PromotionCostTrainingMarks` или простая привязка к tier цели.
2. Показ знаков в explore bar (нижняя панель), если нужно заметнее, чем в верхнем HUD.
3. Сохранение сессии / перенос знаков между запусками — только если появится save.
4. Альтернативный источник знаков (редкий пикап) — без расширения экономики до инвентаря.
5. Локализация строк policy/UI в одном месте.

## Путь к отчёту

`docs/promotion_cost_step_report.md`

## Временное / ограничения шага

- Цена фиксирована глобально; нет полей в `unitdata` (намеренно).
- Знаки не начисляются за поражение / отступление.
- HUD сдвинул вторую строку вниз (y=40); при очень маленьком окне возможен overlap с другими элементами — при необходимости вынести в компактную одну строку.
