# World-integrated recruit sources — отчёт этапа

## 1. Executive summary

- **Цель:** встроить **реальные источники рекрута** в explore/world loop поверх foundation (`hero.NewRecruitHero`, `party.AddToReserve`), не ломая battle / Active+Reserve / recovery / progression.
- **Реализация:** новый тип пикапа **`PickupKindRecruitCamp`**; детерминированный **один** лагерь в чанке **(1,0)** (к востоку от стартовой зоны); **лазурный** маркер на карте; при шаге на клетку — **`ModeRecruitOffer`** с подтверждением (Y/Enter vs N/Esc); при успехе — `AddToReserve` + **`MarkRecruitPickupCollected`**.
- **F9:** сохранён как **демо/debug** (подсказка внизу экрана обновлена).
- **Переполнение:** при полном ростере сообщение, **лагерь не исчезает** (можно вернуться позже).

## 2. Состояние до изменений

- Пикапы: `entity.Pickup` без типа; `World.CollectPickupAt` сразу помечал собранным; жёлтый квадрат в `render/draw.go`.
- Рекрут только через **F9** в explore (вне мира).
- Точка интеграции: `TryMovePlayer` → после `Move` вызов сбора пикапа.

## 3. Целевая модель этапа

| Элемент | Правило |
|--------|---------|
| Источник | Специальный **пикап** `RecruitCamp` в данных чанка. |
| Размещение | **Один** лагерь на мир: чанк `(1,0)`, позиция от **hash(seed, attempt)**, затем **fallback** — скан чанка по первой подходящей клетке (вне зоны 0–6 стартового квадрата), без пересечения с другим пикапом. |
| Взаимодействие | Шаг на клетку → `PickupInteractRecruitOffer` → overlay подтверждения. |
| Успех | `AddToReserve(NewRecruitHero())` → `MarkRecruitPickupCollected` — источник **исчезает**. |
| Отказ | Esc/N → ход мира продолжается, пикап **остаётся**. |
| Full party | Сообщение, пикап **остаётся**, ход мира продолжается. |
| Визуал | Жёлтый = ресурс, **лазурный** = лагерь. |

**Design choice (этап):** не NPC и не квест — **минимальный** расширяемый пикап; генерация **контролируемая** (чанк 1,0), не полноценная procedural recruit-сеть.

## 4. Изменённые файлы

| Файл | Содержание |
|------|------------|
| `world/entity/pickup.go` | `PickupKind`, `PickupKindRecruitCamp`, поле `Kind`. |
| `world/pickup_interaction.go` | **Новый:** `PickupInteractionResult`. |
| `world/state.go` | `InteractPickupAfterMove`, `MarkRecruitPickupCollected`, `generateRecruitCampPickups`, ключ для обычных пикапов `Kind: Resource`. |
| `world/render/draw.go` | Цвет пикапа по `Kind`. |
| `internal/player/movement.go` | Интерфейс `InteractPickupAfterMove`, возврат `PickupInteractionResult`. |
| `internal/game/game.go` | `ModeRecruitOffer`, `recruitOfferX/Y`. |
| `internal/game/update.go` | Обработка режима; ветка `RecruitOffer` после move без `advanceWorldTurn` до решения. |
| `internal/game/recruit_offer.go` | **Новый:** `updateRecruitOfferMode`. |
| `internal/game/draw.go` | Party strip в recruit mode; overlay. |
| `internal/ui/recruit_offer.go` | **Новый:** `DrawRecruitOfferOverlay`. |
| `internal/ui/explore_hud.go` | Подсказка: лагерь на карте + F9 демо. |

## 5. Инварианты

- Новый член — **`hero.Hero`** через `NewRecruitHero`, только **`AddToReserve`**.
- **Active** не меняется автоматически.
- **Battle** — только `PlayerCombatSeeds` из Active (без изменений).
- Нет параллельной сущности «recruit» вне `Hero`/`Party`.

## 6. Упрощения / follow-up

- Один фиксированный чанк для лагеря; нет плотной сети лагерей по миру.
- Нет экономики, диалогов, пост-боя как источника.
- Нет сохранения — при будущем save нужно сериализовать `PickupKind` и `collectedPickups`.

## 7. Ручная проверка

- [ ] Идти **вправо** от старта в чанк (1,0), найти **лазурный** квадрат.
- [ ] Встать на клетку → overlay → Y → новобранец в резерве, маркер исчез.
- [ ] F5 — виден в резерве; перевод в active → бой.
- [ ] Full party (12) на лагере → сообщение, маркер остаётся.
- [ ] Отказ Esc → маркер остаётся, можно уйти и вернуться.
- [ ] F9 — по-прежнему демо-рекрут.
- [ ] Обычные жёлтые пикапы считаются в HUD как раньше.

## 8. Follow-up

- Несколько типов лагерей / редкость / экономика.
- NPC-взаимодействие, события после боя.
- Процедурное размещение с правилами биома.
- Уникальные шаблоны новобранцев.
