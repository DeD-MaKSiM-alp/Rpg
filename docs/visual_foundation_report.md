# Отчёт: visual / UI foundation (battle, formation, explore)

## 1. Executive summary

**Цель этапа:** заложить единый визуальный foundation для уже стабилизированных контуров (battle HUD v1/v2, formation overlay, explore hints / recovery), без финального production-art и без изменения gameplay-логики.

**Результат:** введён **`internal/ui/theme.go`** — палитра `Theme` (панели, текст, ally/enemy, active/hover/target/dead, кнопки, HP-bars, post-battle, explore-bar) и хелперы **`DrawHPBarMicro`**, **`DrawThinAccentLine`**. Все перечисленные экраны переведены на эту систему; добавлены **микро-HP-бары** на слотах v1, карточках v2 и в formation/explore. Добавлен **`DrawExplorePartyStrip`** — компактная панель отряда в explore.

**Затронуто:** battle (v1/v2), formation overlay, explore (party strip + hints), post-battle overlay, `drawPanelBox`, HUD pickup/resolution.

**Совместимость:** геометрия layout по-прежнему из `battle.ComputeBattleHUDLayout` / `ComputePostBattleLayout`; hit-test и input не менялись.

---

## 2. Состояние до изменений

| Область | Было |
|--------|------|
| **Battle** | Разрозненные `color.RGBA` в `battle_panels.go`; v2 уже структурирован, но без общей палитры; HP только текстом. |
| **Formation** | Плоский текст на затемнении, без группировки и без визуального «списка карт». |
| **Explore** | Подсказки F5/R текстом без подложки; нет сводки партии на экране мира. |
| **Post-battle** | Собственные локальные цвета, близкие к панелям, но не именованные. |

**Главные проблемы:** отсутствие единого языка состояний (ally vs enemy, acting vs wait), слабая иерархия, «техпрототипный» вид, дублирование числовых RGB.

---

## 3. Целевая визуальная модель этапа

- **Панели:** `Theme.PanelBG` / `PanelBorder` / `PanelTitleSep` — единый каркас.
- **Текст:** `TextPrimary` → заголовки и важное; `TextSecondary`/`TextMuted` для фаз и подсказок.
- **Бой:** `ActiveTurn` (ход), `WaitAlly` (очередь союзника), `HoverTarget` / `ValidTarget` / `SelectedKill` (цели), `EnemyAccent` vs `AllyAccent` на слотах.
- **HP:** трек `HPBarTrack` + заливка `HPAllyFill` / `HPEnemyFill` (микро-бары без отдельного asset pipeline).
- **Explore:** нижняя полоса `ExploreBarBG` + `RecoveryBanner` для успешного отдыха; **party strip** отдельно сверху.
- **Post-battle:** `PostBattlePanelBG`, выбор строки `PostBattleRowSelect` / `PostBattleRowBrd`.

---

## 4. Изменённые файлы

| Файл | Изменения |
|------|-----------|
| `internal/ui/theme.go` | **Новый:** `Theme`, `DrawHPBarMicro`, `DrawThinAccentLine`. |
| `internal/ui/formation_overlay.go` | Formation: центрированная панель, акцентная линия, карточки строк, HP-бар, `Theme`. |
| `internal/ui/explore_hud.go` | **Новый:** `DrawExplorePartyStrip`, `DrawExploreHintPanelLayout`, `DrawExploreFormationHintLines`, `DrawExploreFormationHint`. |
| `internal/ui/battle_panels.go` | Все цвета → `Theme`; v1/v2 HP-бары; v2 ростеры с акцентной линией; кнопки/способности из темы. |
| `internal/ui/panels.go` | `drawPanelBox` использует `Theme`; pickup HUD — `Theme.TextPrimary`. |
| `internal/ui/postbattle.go` | Панель и строки награды из `Theme`; удалён неиспользуемый импорт `image/color`. |
| `internal/ui/draw.go` | Resolution indicator — `Theme.TextMuted`. |
| `internal/game/draw.go` | В explore: `DrawExplorePartyStrip` перед подсказками. |

---

## 5. Ключевые инварианты после изменений

- **Layout / input:** неизменны; отрисовка только читает `battle` state.
- **Состояния юнитов:** те же правила подсветки (active, hover, target), цвета вынесены в `Theme`.
- **HP:** канонические числа по-прежнему в тексте; бары — визуальное усиление, не второй источник истины.
- **Foundation без ассетов:** только `vector` + существующий шрифт HUD.

---

## 6. Что осталось упрощённым / временным

- Поле боя — placeholder (тёмные прямоугольники), без персонажных спрайтов.
- Нет портретов, анимаций, VFX, скруглений панелей (только прямоугольники).
- Типографика — один HUD face; без отдельного scale/line-height тюнинга.

---

## 7. Ручная проверка (checklist)

- [ ] Explore: party strip, pickups, F5/R подложка, баннер после `R`.
- [ ] Formation: карточки, выбор, HP, Esc/F5.
- [ ] Battle v2 (default): ростеры, бары, top bar, bottom panel, abilities, post-battle поверх.
- [ ] Battle v1 (F8): overlay сетка, слоты, бары.
- [ ] Acting / dead / enemy / target hover — согласованные цвета.
- [ ] Post-battle: выбор награды, hit-test строк (как раньше).
- [ ] Разрешение F6/F7 — строка в углу читаема.

---

## 8. Follow-up

- Отдельный **battle skin pass** (арт поля, рамки слотов).
- **Portrait / token** система для карт и мира.
- **Анимации** хода, урона, heal numbers.
- **World visual pass** (тайлы, сущности).
- Расширение **theme** (night mode, accessibility contrast).
