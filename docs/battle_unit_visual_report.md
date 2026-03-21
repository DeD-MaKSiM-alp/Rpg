# Отчёт: battle unit visual pass (токены на поле, v2)

## 1. Executive summary

**Цель:** сделать так, чтобы юниты на **battlefield** воспринимались как сущности (токены/состояния), а не только как строки в боковых ростерах, без финального арта и без ломки battle pipeline.

**Результат (v2):**
- В `battle` layout добавлены **`BattlefieldTokens`** (rect токена на поле) и **`BattlefieldSlotCells`** (12 ячеек сетки для подсветки пространства).
- **Hit-test** (`hud_mouse.go`) расширен: **`pointHitsUnit`** — клик по токену на поле эквивалентен клику по карточке в ростере.
- В **`internal/ui/battlefield_tokens.go`** — отрисовка: разделитель сторон, сетка слотов, **круглые токены** (ally/enemy заливка из `Theme`), рамки состояний (как в HUD), мёртвые — крест, индекс (партия `1..n` или `id%10`), микро-HP-бар, дополнительное кольцо для acting.
- В **`Theme`** добавлены **`BattlefieldTokenAlly`** / **`BattlefieldTokenEnemy`**.

**Совместимость:** v1 layout не задаёт `BattlefieldTokens` (nil); логика боя и `UnitRects` для ростеров не удалены.

---

## 2. Состояние до изменений

- Центральное поле (battlefield) было **пустым placeholder** (тёмные прямоугольники).
- Юниты визуально существовали только в **левом/правом ростере**; пространство боя не показывало formation/spatial presence.
- Клики по полю на юнита **не работали** — только по карточкам ростера.

---

## 3. Целевая визуальная модель этапа

| Элемент | Модель |
|--------|--------|
| Юнит на поле | **Круглый токен** + обводка состояния + короткий числовой лейбл |
| Ally / enemy | Заливка `BattlefieldTokenAlly` / `BattlefieldTokenEnemy` |
| Состояния | Те же семантические цвета, что у карточек: active, wait, hover, selected target, dead |
| HP | Микро-бар под токеном + согласование с ростером |
| Пространство | 12 ячеек (front/back × 3 × 2 сторон), линия между сторонами |
| Расширяемость | Отрисовка изолирована в `battlefield_tokens.go`; позиции — из layout |

---

## 4. Изменённые файлы

| Файл | Назначение |
|------|------------|
| `internal/battle/hud_layout.go` | Поля `BattlefieldTokens`, `BattlefieldSlotCells`; функция `computeBattlefieldPlacements`; заполнение в `computeLayoutV2`. |
| `internal/battle/hud_mouse.go` | `BattleHUDLayout.pointHitsUnit`; клики default attack и choose target учитывают токены. |
| `internal/ui/battlefield_tokens.go` | **Новый:** `DrawBattlefieldV2Scene`, `drawBattlefieldUnitToken`. |
| `internal/ui/battle_panels.go` | Вызов `DrawBattlefieldV2Scene` после фона battlefield. |
| `internal/ui/theme.go` | `BattlefieldTokenAlly`, `BattlefieldTokenEnemy`. |

---

## 5. Ключевые инварианты

- Источник геометрии — **`ComputeBattleHUDLayout`**; токены и ячейки согласованы с одной формулой `computeBattlefieldPlacements`.
- **Hit-test** для unit ID объединяет ростер и поле; поведение action pipeline не менялось.
- Multi-unit: до 6 юнитов на сторону; токены не пересекают боковые панели (центральная зона).

---

## 6. Что упрощено / временно

- Нет спрайтов, анимаций, VFX.
- Токены — **генерируемые** круги vector; враги без уникальных силуэтов (только цвет + индекс).
- v1 (табличный HUD) по-прежнему без полевых токенов.

---

## 7. Ручная проверка (checklist)

- [ ] v2: на поле видны сетка, разделитель, токены с HP-баром.
- [ ] Клик по **вражескому токену** в режиме basic attack — атака срабатывает как по карточке.
- [ ] Choose target для способности — клик по токену валидной цели.
- [ ] Hover подсветка на токене и в ростере (один `HoverTargetUnitID`).
- [ ] Active / wait / dead визуально согласованы с ростером.
- [ ] v1 (F8): без регрессий, поле как раньше в табличном режиме.

---

## 8. Follow-up

- Подмена круга на **sprite/token** через конфиг визуала без смены layout.
- Лёгкая **анимация** (pulse active), hit-flash.
- **Архетипы врагов** — разные формы/цвета.
- Опционально: полевые токены для **v1** или общий слой `BattlefieldPresenter`.
