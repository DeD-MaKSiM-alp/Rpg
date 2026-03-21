# Отчёт: ветка group-healer — минимальное массовое лечение в бою

## 1. Модель group-heal

**Все живые союзники** получают одну и ту же величину лечения за одно действие: `GroupHealPower() = 1 + Def.Base.HealPower` (на союзника).  
Цель в бою **не выбирается**: правило таргета `TargetAllyTeam`, запрос с `TargetKindNone`, подтверждение сразу после выбора способности (как у «no-target» ветки в `update.go`).

## 2. Почему так

- Соответствует приоритету 1 из ТЗ: без сложного UI выбора, без AoE по клеткам.
- Отличие от одиночного лечения задаётся **типом способности** `AbilityGroupHeal`, а не `UnitID` в `resolve.go`.
- Синхронизация с party: HP по-прежнему живут в `CombatUnitState`; массовое лечение только меняет HP нескольких юнитов в одном `ResolveAbility` — пост-бой `syncPartyFromBattle` не менялся.

## 3. Изменённые файлы

| Файл | Суть |
|------|------|
| `internal/battle/ability.go` | `AbilityGroupHeal`, `TargetAllyTeam`, реестр (имя «Массовое лечение»). |
| `internal/battle/validation.go` | Список валидных целей `[NoTarget()]`, валидация без unit-target. |
| `internal/battle/resolve.go` | Ветка `AbilityGroupHeal`: обход `LivingUnits(actor.Side)`, кап до MaxHP, лог. |
| `internal/battle/action.go` | `HealApplication` + срез `HealApplications` в `ActionResult`. |
| `internal/battle/battle.go` | `ApplyActionResult`: несколько `pushHealFeedback` по срезу. |
| `internal/battle/unit.go` | `GroupHealPower()`. |
| `internal/battle/preview.go` | Превью лечения для group heal. |
| `internal/battle/update.go` / `hud_mouse.go` | Явный `case TargetAllyTeam` → немедленное выполнение. |
| `internal/unitdata/unitdata.go` | `empire_healer_group_1`: `AbilityGroupHeal`, `HealPower: 0`. |
| `internal/hero/hero.go` | Роль healer при `AbilityGroupHeal`. |
| `internal/ui/character_inspect.go` | Подпись способности. |
| `internal/ui/battle_panels.go` | Цель «все союзники» в summary/target для массового лечения. |
| `internal/battle/group_heal_test.go` | Тесты поведения. |
| `internal/unitdata/unitdata_test.go` | Различие single vs group в данных. |
| `internal/hero/hero_test.go` | `CombatUnitSeed` для ветки group. |

## 4. Single vs group в данных и в бою

| | Single (`empire_healer_single_1`) | Group (`empire_healer_group_1`) |
|---|--------------------------------|--------------------------------|
| Способность | `AbilityHeal` | `AbilityGroupHeal` |
| Формула | `HealPower()` = 2 + bonus (bonus 1 → **3** одной цели) | `GroupHealPower()` = 1 + bonus (bonus 0 → **1** каждому живому союзнику) |
| Таргет | Один союзник | Вся сторона (без клика по цели) |

## 5. Встраивание без переписывания battle

- Один новый `AbilityID` и одно правило `TargetRule`; исполнение — дополнительный `case` в `ResolveAbility`.
- Существующий поток «способность без цели» уже был в `default` ввода игрока; для ясности добавлен `case TargetAllyTeam` с `fallthrough`.

## 6. Тесты

- Данные/разница профилей, разрешение без unit-target, групповое лечение двух союзников, враг не трогается, кламп MaxHP, одиночный heal одной цели, seed героя group-ветки.

## 7. Ограничения

- Нет приоритета «самые раненые», нет splash — только равномерный «всем живым».
- Имя в реестре у старых способностей по-прежнему на английском; для group heal задано русское имя в HUD.
- Прогрессия `RewardAbilityHeal` по-прежнему про `AbilityHeal`, не про group (ветка group уже получает способность из шаблона).

## 8. Следующие шаги (3–5)

1. Локализовать подписи **Heal / Shoot** в battle HUD в том же стиле, что «Массовое лечение».
2. При необходимости — слабый баланс `HealPower` bonus для group vs single после плейтеста.
3. Отдельная награда прогрессии на усиление `GroupHealPower`, если понадобится.
4. AI врагов с heal (если появятся) — выбор между single/group по эвристике.
5. Опционально: heal только **раненым** союзникам (меньше «пустых» нажатий), без усложнения таргетинга.
