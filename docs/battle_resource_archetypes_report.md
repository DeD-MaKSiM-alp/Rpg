# Боевые resource-профили (архетипы маны / энергии)

## Три профиля

| Профиль | Смысл | Пулы (мана / энергия) | Regen за полный раунд |
|--------|--------|------------------------|------------------------|
| **ManaFocused** | Маги и целители | 18 / 4 | +1 мана, +0 энергии |
| **EnergyFocused** | Дальний / физический ритм | 4 / 18 | +1 мана, +2 энергии |
| **Striker** | Простые бойцы, упор на базовый удар и КД | 8 / 10 | +1 / +1 |

Профиль по умолчанию выводится из `Role` в `CombatUnitDefinition`. Если `Role == Fighter` (в т.ч. legacy `BuildPlayerCombatSeed`), роль уточняется по loadout (`Heal` → целитель, `Ranged` → лучник, `Buff` → маг). Явный override: поле `ResourceProfile` в `CombatUnitDefinition` (не `Unset`).

## Привязка к ролям и контенту

- **RoleHealer / RoleMage** → ManaFocused (героиня-целитель/маг из `unitdata`, враг-бандит как целитель в шаблоне).
- **RoleArcher** → EnergyFocused (лучники, враг «Волк»).
- **RoleFighter** → Striker (копейщик/новобранец, враг «Слайм»).

## Способности (playable slice)

- **Усиление (Buff):** только мана (3), без энергии — под мана-профиль мага.
- Остальные стоимости без массового рефакторинга реестра: лечение/масс-лечение — мана; дальний удар — энергия.

## Файлы

- `internal/battle/resource_profile.go` — типы, вывод профиля, пулы, regen.
- `internal/battle/resources.go` — `initCombatResources(st, def)`, `tickRoundResources` с regen по профилю.
- `internal/battle/unit.go` — поле `ResourceProfile` в определении юнита.
- `internal/battle/battle.go` — инициализация ресурсов при спавне.
- `internal/battle/ability.go` — правка `AbilityBuff`.
- `internal/battle/resource_profile_test.go` — тесты профиля и regen.

## Техдолг

- Явные профили в `EnemyTemplate` / data-driven таблица.
- Отдельная строка UI «архетип ресурса» в inspect (по желанию).
