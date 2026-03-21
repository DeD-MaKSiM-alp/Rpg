# Отчёт: исправление принятия найма из лагеря (recruit camp)

## 1. Симптом

На лазурной точке лагеря наёмников открывалось предложение найма (overlay), но подтверждение (**Enter / Space / Y**) не добавляло юнита в отряд — казалось, что accept не работает.

## 2. Точная причина

В `Game.Update()` обрабатывались только режимы `ModeBattle` и `ModeFormation`; для любого другого режима вызывался **`updateExploreMode()`** без проверки.

Режим **`ModeRecruitOffer`** никогда не попадал в **`updateRecruitOfferMode()`** — функция была реализована в `recruit_offer.go`, но **нигде не вызывалась** из `Update()`. Визуально overlay рисовался (`draw.go` для `ModeRecruitOffer`), а логика подтверждения не выполнялась.

Связь с недавним gating promotion **косвенная**: баг не в `PlayerStandsOnActiveRecruitCamp` и не в collected-пикапе; корневая причина — **отсутствие dispatch** для `ModeRecruitOffer` в главном цикле обновления. Ошибка, вероятно, существовала с момента появления отдельной моды без маршрутизации (или после рефакторинга `Update`).

## 3. Исправление (минимальный diff)

В `internal/game/update.go` после ветки `ModeFormation` добавлена ветка:

```go
if g.mode == ModeRecruitOffer {
    g.updateRecruitOfferMode()
    return nil
}
```

До вызова `updateExploreMode()`. Так ввод подтверждения обрабатывается только в `updateRecruitOfferMode`, а explore-движение и прочие клавиши не перехватывают кадр в режиме оффера.

## 4. Изменённые файлы

| Файл | Изменение |
|------|-----------|
| `internal/game/update.go` | Dispatch `ModeRecruitOffer` → `updateRecruitOfferMode` |
| `internal/game/recruit_offer_flow_test.go` | **Новый:** регрессия пути add-to-reserve (как в recruit offer) |
| `docs/recruit_camp_accept_bugfix_report.md` | Этот отчёт |

## 5. Lifecycle лагеря (ожидаемое поведение)

1. Игрок наступает на клетку с несобранным `PickupKindRecruitCamp` → `InteractPickupAfterMove` → `ModeRecruitOffer`, сохраняются `recruitOfferX/Y`.
2. Каждый кадр: **`updateRecruitOfferMode()`** обрабатывает Escape/N (отказ + `advanceWorldTurn`) или Enter/Space/Y (принятие).
3. При принятии: `AddToReserve`, `MarkRecruitPickupCollected`, `ModeExplore`, баннер.
4. **Promotion** по-прежнему использует `PlayerStandsOnActiveRecruitCamp` только при **P** на карточке состава; на accept recruit это не влияет.

## 6. Почему не ломается promotion gating

Gating не менялся: проверка «стоим на активном recruit camp» остаётся в formation + `draw` для inspect. Исправление только подключает **`updateRecruitOfferMode`** к игровому циклу.

## 7. Тесты

- `TestRecruitCampOffer_addsHeroToReserve` — тот же сценарий данных, что успешный accept в `recruit_offer.go`, без ebiten.

Полный e2e с `inpututil` в проекте не добавлялся (без тяжёлой инфраструктуры).

## 8. Спорные места

- Пока нет автоматического теста, что `Update()` **всегда** вызывает `updateRecruitOfferMode` при `ModeRecruitOffer` — инвариант закреплён комментарием в коде; при рефакторинге `Update` ветку нужно сохранить.
