# Отчёт: мост template identity → battle (`battle_identity_bridge`)

## 1. Что изменено

- В **`battle.CombatUnitDefinition`** добавлены поля канонической identity шаблона (без импорта `unitdata` в battle): **`TemplateUnitID`**, **`FactionID`**, **`LineID`**, **`Tier`**, **`IdentityAttackKind`** (тип `TemplateAttackKind`), плюс **`ArchetypeID`** теперь заполняется из шаблона для игроков с валидным `hero.UnitID`.
- Сборка **`hero.CombatUnitSeed()`** при известном `UnitID` делает **lookup в `unitdata`** и переносит identity в сид; **роль** и **`IsRanged`** для нормального пути берутся из **шаблона** (`Role`, дальность — лучник = дальний бой).
- **LEGACY:** при пустом/неизвестном `UnitID` роль по-прежнему выводится из **способностей**; **`IdentityAttackKind`** = unknown.
- В **`unitdata.UnitTemplate`** добавлено поле **`UpgradeToUnitID`** (заполнено ссылками на следующий tier из design-doc id, без игровой логики).
- **Inspect (вне боя):** строка «Следующий шаг (данные): …» при непустом `UpgradeToUnitID`.
- **Battle HUD v2:** в строке «Ваш ход: …» добавлен суффикс **` · template_unit_id`** через **`battle.PlayerTemplateIdentitySuffix`** (только игрок, только если id задан).
- Тесты: hero (identity + legacy), battle (suffix), unitdata (upgrade link).

## 2. Поток identity: template → hero → seed → battle unit

1. **`unitdata.UnitTemplate`** — канон identity + `UpgradeToUnitID`.
2. **`hero.Hero`** хранит **`UnitID`** и runtime-статы.
3. **`CombatUnitSeed`** строится в **`hero.CombatUnitSeed`**: lookup шаблона, заполнение **`Def.*`** identity-полей; статы по-прежнему из героя через **`BuildPlayerCombatSeed`**.
4. **`BattleContext`** создаёт **`CombatUnit`** с копией **`Def`**; отображаемое имя в бою по-прежнему **`PlayerAllyDisplayName`** (слот), не `DisplayName` шаблона.

## 3. Канон vs runtime-derived

| Слой | Канон (шаблон / hero.UnitID) | Runtime (герой / бой) |
|------|------------------------------|------------------------|
| Стартовые статы в бою | — | `Hero` → `Def.Base` |
| Роль в battle Def | Из шаблона, если `UnitID` валиден | Иначе из способностей (legacy) |
| `IsRanged` | `Role == Archer` после выставления роли | Совпадает с механикой дальника |
| `ArchetypeID` в Def | Ключ архетипа из шаблона | Иначе остаётся `"player"` из `BuildPlayerCombatSeed` |
| `DisplayName` в бою | — | Слот: «Лидер», «Союзник N» |

## 4. Роль / дальность / тип атаки

- **Кто это (роль, архетип, faction, line, tier, attack kind в смысле дизайна):** из **шаблона**, если lookup успешен.
- **Дальность боя (`IsRanged`):** для канонического пути **`Role == Archer`** → дальний бой; иначе ближний. Legacy — то же от роли, полученной из абилок.
- **`IdentityAttackKind`:** копия design `AttackKind` для UI/отладки; **не** дублирует проверки таргетинга (они по-прежнему на способностях/`IsRanged`).

## 5. Upgrade / evolution (только данные)

- Поле **`UpgradeToUnitID`** в шаблоне; значения вроде `empire_warrior_squire`, `empire_archer_marksman_base`, `empire_healer_acolyte` — ориентиры для следующего этапа, без promotion UI и без изменения `Hero` в бою.

## 6. Fallback / legacy

- Пустой или неизвестный **`UnitID`**: нет `TemplateUnitID` в Def; роль из **`battleRoleFromAbilities`**; **`IdentityAttackKind`** = unknown.
- **`recruitHeroFallbackNoTemplate`** и старый **`DefaultHero`** fallback без id — тот же legacy-путь.

## 7. Сознательно не делалось

- Полная template-driven боевая модель, save/load, экран апгрейда.
- Импорт **`unitdata`** в пакет **`battle`** (lookup только в **`hero`**).
- Расширение enemy-side identity (враги по-прежнему с enemy-шаблонами как раньше).

## 8. Следующие логичные шаги (3–5)

1. Реализовать смену **`Hero.UnitID`** при promotion по **`UpgradeToUnitID`** (один шаг).
2. Синхронизировать отображаемое имя в бою с шаблоном (опционально, не ломая слоты).
3. Логи боя: всегда писать **`TemplateUnitID`** для союзников.
4. Валидация: предупреждение, если абилки героя не совпадают с шаблоном.
5. Вынести registry в файл, когда понадобится контент-пайплайн.
