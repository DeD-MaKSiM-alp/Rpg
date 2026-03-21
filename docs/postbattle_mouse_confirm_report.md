# Post-battle: mouse confirm (краткий отчёт)

## 1. Где был keyboard-only confirm

- **`internal/postbattle/postbattle.go`**, `Flow.Update`, ветка **`StepResult`**: переход к награде или выход в мир обрабатывался **только** через `inpututil.IsKeyJustPressed` для `Space` и `Enter`.
- Шаг **`StepReward`** уже поддерживал мышь: клик по строке награды вызывал `ApplyReward` (через `rewardOptionAtCursor` + `ComputePostBattleLayout`).

## 2. Выбранный mouse confirm path

- **Экран результата (`StepResult`)**: явная кнопка с подписью **«Продолжить»** (победа) или **«В мир»** (поражение / побег и т.д.) — hit-test по `PostBattleLayout.ResultContinueButton`, тот же прямоугольник, что и в `DrawPostBattleOverlay`.
- **Экран выбора награды (`StepReward`)**: добавлена кнопка **«Подтвердить»** (`RewardConfirmButton`) — эквивалент `Space`/`Enter` для **текущего** выбранного индекса; клик по строке награды по-прежнему сразу применяет эту награду (приоритет строк выше, чем кнопка — строки выше по экрану, пересечений нет).
- **`Space` / `Enter`** не убирались: по-прежнему вызывают те же функции **`confirmResultStep`** и **`confirmRewardSelection`**.

## 3. Изменённые файлы

| Файл | Назначение |
|------|------------|
| `internal/ui/postbattle.go` | Расчёт `ResultContinueButton` / `RewardConfirmButton` в `ComputePostBattleLayout`; отрисовка кнопок с hover; подсказки на русском; хелпер `drawPostBattlePrimaryButton`. |
| `internal/postbattle/postbattle.go` | `confirmResultStep`, `confirmRewardSelection`; мышь на результате и на кнопке подтверждения награды; удалён дублирующий `rewardOptionAtCursor`. |
| `internal/postbattle/draw.go` | `BuildPostBattleParams`: один вызов `ComputePostBattleLayout`, подписи кнопок, hover с курсора. |

## 4. Суть изменений

- **Один канонический путь** для клавиатуры и мыши: логика «что произошло при подтверждении» сосредоточена в `confirmResultStep` и `confirmRewardSelection`.
- **Draw и hit-test** используют **один** `ComputePostBattleLayout` с теми же флагами шага и числом опций.

## 5. Почему не ломается reward flow

- Клик по **опции награды** обрабатывается **раньше** клика по «Подтвердить» (`RewardOptionIndexAt` → иначе `HitRewardConfirm`).
- Нет «клика в пустоту» для закрытия: только **явная кнопка** на экране результата и **кнопка подтверждения** или **строка награды** на шаге награды.
- `ApplyReward` вызывается только из `confirmRewardSelection` с валидным индексом.

## 6. Ручная проверка

1. Пройти бой до конца → открыть post-battle.
2. На экране **результата**: нажать **«Продолжить»** / **«В мир»** мышью — то же поведение, что **Space**.
3. На экране **награды** (победа): выбрать строку **мышью** — награда применяется; **или** выбрать стрелками и нажать **«Подтвердить»** мышью; **или** **Space**.
4. Убедиться, что клик **мимо** кнопок и строк **не** завершает экран.
5. Поражение: с экрана результата выйти мышью через **«В мир»**.

## 7. Follow-up (не делалось)

- Локализация заголовков «Victory!» / «Defeat» в том же стиле, что кнопки.
- Таб-фокус / полная доступность только с клавиатуры (уже есть стрелки + Space).
