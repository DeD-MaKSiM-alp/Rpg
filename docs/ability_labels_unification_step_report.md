# Ability labels unification — step report

## 1. Где были несогласованные названия

- **Inspect-карточка** использовала локальный `abilityNameRu` в `character_inspect.go` (короткие русские строки).
- **Battle HUD** (панель способностей, summary v1/v2, блок ACTION) брал **`GetAbility(id).Name`** из `internal/battle/ability.go`: там английские/смешанные имена (`Attack`, `Shoot`, `Heal`, `Buff`, частично русское «Массовое лечение»).
- **Боевой лог** в `resolve.go`: английские фразы (`hits`, `heals`, `buffs`) и отдельные русские строки про массовое лечение — без единых коротких имён способностей.

## 2. Единый источник правды

Введена функция **`battle.PlayerAbilityLabelRU(id AbilityID) string`** в файле `internal/battle/ability_labels_ru.go`.

- Один маппинг в пакете `battle` (доменные id, presentation-строки).
- Пакет `ui` уже импортирует `battle` — дублирования второго маппинга в `ui` нет.

## 3. Canonical player-facing short labels

| AbilityID | Подпись |
|-----------|---------|
| `AbilityBasicAttack` | Базовый удар |
| `AbilityRangedAttack` | Дальний удар |
| `AbilityHeal` | Лечение |
| `AbilityGroupHeal` | Масс-лечение |
| `AbilityBuff` | Усиление |
| неизвестное значение | Способность |

## 4. Куда подключено

- **Inspect** (`abilityLinesBullet` в `battle_inspect.go`): `PlayerAbilityLabelRU`; удалён `abilityNameRu` из `character_inspect.go`.
- **Battle HUD v1**: `drawAbilityPanel`, `drawConfirmPanel` (строка «Способность: …»), список способностей v2 в `drawBattleScreenV2`.
- **V2 summary**: строка «Базовый удар · клик по врагу» и превью `способность → цель` через `PlayerAbilityLabelRU`.
- **Боевой лог** (`resolve.go`): сообщения переписаны на короткий русский формат с тем же `PlayerAbilityLabelRU`, чтобы имя способности совпадало с карточкой и HUD.

Реестр `abilityRegistry` в `ability.go` **не менялся** (внутренние имена и логика без изменений).

## 5. Fallback

- Неизвестный `AbilityID` → **«Способность»** (не пусто, без raw id).

## 6. Изменённые файлы

| Файл | Изменение |
|------|-----------|
| `internal/battle/ability_labels_ru.go` | новый helper |
| `internal/battle/resolve.go` | логи с едиными подписями |
| `internal/ui/battle_inspect.go` | inspect → `PlayerAbilityLabelRU` |
| `internal/ui/character_inspect.go` | удалён `abilityNameRu` |
| `internal/ui/battle_panels.go` | HUD/summary на `PlayerAbilityLabelRU` |
| `internal/battle/ability_labels_ru_test.go` | тесты маппинга и fallback |
| `internal/ui/ability_labels_ui_test.go` | согласованность inspect + helper |

## 7. Тесты

- Все известные id дают ожидаемые строки.
- Неизвестный id → «Способность».
- Inspect flatten содержит ту же подпись, что и `PlayerAbilityLabelRU` для базового удара.

## 8. Ограничения и временное

- Строки вроде `Step:` / `Target:` в v1 ACTION summary по-прежнему на английском — это не имена способностей; вынос на отдельный шаг при желании.
- `internal/battle/ability.go` поле `Name` у записей реестра остаётся для внутренней совместимости; UI не показывает его напрямую.

## 9. Следующие логичные UX-шаги

- При необходимости локализовать остальные подписи панели ACTION (фаза/цель) на русский в том же тоне.
- Проверить, не осталось ли в других экранах (debug, тесты) ссылок на `GetAbility(...).Name` для игрока.
