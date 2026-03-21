# Battle design contract (Disciples-like), project baseline

This document defines the **battle design contract** for the next development stages of the project’s turn-based party combat system (Disciples-inspired, adapted to current codebase).

It is written as an engineering contract: terminology, invariants, responsibilities, and data flow boundaries. Implementation can evolve, but **must preserve the semantics** defined here unless this document is updated.

---

## 1) Goal of the battle system

### Battle as a dedicated game mode
- **Battle is a separate game mode** (distinct from overworld exploration), with its own update loop, input rules, UI and state machine.
- Battle is initiated from overworld when an **Encounter** is created (e.g. contact with an enemy) and is resolved back into the overworld via a **BattleResult**.

### Overworld ↔ battle boundary
- Overworld owns persistent entities/state (world map, enemies as world entities, player persistent state).
- Battle owns only battle-runtime state and applies effects within battle. When battle ends, the game layer applies the outcome to overworld (remove defeated enemies, award rewards, etc.).

### Why party + slot-oriented
- Combat is **party-based**: both sides consist of multiple units.
- Units occupy fixed **slots/positions** in a formation (front/back rows). Slotting makes:
  - target rules explicit and deterministic;
  - formation tactics meaningful (protect backline, melee screening);
  - UI and AI simpler (enumerable positions, stable references).

### Disciples-like inspirations (adapted)
- Two opposing sides with **front/back formation rows**.
- Targeting constraints that differ between melee/ranged/support.
- Turn-based flow with per-unit turns and deterministic resolution.
- Adaptations for this project:
  - the battle module is a pure gameplay mode (no tactical grid movement inside battle);
  - expansion path for abilities/effects/status systems aligned with current code style.

---

## 2) Core entities (canonical definitions)

> Names below are *domain contract names*. Code may use different type names, but must map cleanly.

### Battle
- Orchestrator of a single combat instance.
- Owns battle loop, phase transitions, and produces a final `BattleResult`.

### BattleState
- The mutable runtime state of a battle: units, slots, turn order, ongoing effects/statuses, round number.
- Must be self-contained: given an initial state and a deterministic RNG seed (if needed), it can simulate the battle.

### BattleSide
- One of two sides: **player side** or **enemy side**.
- Owns its formation slots and references to units occupying them.

### CombatUnit
- A runtime battle unit with combat stats and a set of abilities.
- Must support: alive/dead, disable states, current HP, modifiers/statuses, and identity stable during the battle.

### Slot / Position
- A fixed place in a formation.
- Has a side, row (front/back), and index within row.
- Must be targetable/inspectable even if empty (important for summons and some effects later).

### Ability
- A unit action definition with:
  - targeting rule (what can be targeted);
  - delivery form (single-target, AoE, self, slot-based);
  - range (melee/ranged/etc.);
  - effect payload (damage/heal/buff/debuff/etc.).

### Effect / Resolution
- **Effect**: an atomic gameplay change (damage, heal, apply status, move unit, etc.).
- **Resolution**: the deterministic application of an ability into a sequence of effects and logs.
- Resolution is the **only place** allowed to change unit combat state (HP, statuses, etc.).

### TurnOrder / Initiative
- A deterministic ordering of unit turns (initiative-based; stable tie-breakers).
- Must be robust to unit death/disable: units that cannot act are skipped without breaking the loop.

### Status
- A persistent modifier on a unit (buff/debuff/conditions) with duration/stacking rules.
- Examples: defend, poison, stun, attack buff, armor debuff.

### BattleResult / Outcome
- The final result of a battle, returned to game/world:
  - outcome (victory/defeat/retreat/etc.);
  - summary of survivors/deaths;
  - rewards/xp/resource changes;
  - any persistent status changes (if designed).

### Encounter
- An overworld-to-battle boundary object that describes *what battle should be created*:
  - which enemies participate;
  - any metadata (zone, encounter type, modifiers).
- Encounter should not mutate the overworld by itself.

---

## 3) Battle sides

### Two sides only
- Exactly two sides:
  - **Player side**
  - **Enemy side**

### Fixed slot set per side
- Each side has a **fixed set of formation slots**.
- Slots are divided into:
  - **Front row**
  - **Back row**

### Attack capability is not uniform
- Not all units can attack all targets equally.
- Targeting depends on:
  - ability range (melee vs ranged);
  - ability target rule (enemy/ally/self/slot/group);
  - formation rules (front row screening, etc.).

---

## 4) Placement & formation rules (baseline)

### Front row
- Acts as the **screen**: melee attacks generally must hit front row first if it has living units.
- Typical roles: fighters/tanks/bruisers/melee attackers.

### Back row
- Protected by the front row by default.
- Typical roles: archers/casters/healers/support.

### Default melee restriction (screening)
- **Melee** abilities target:
  - enemy front row if any living front units exist;
  - otherwise enemy back row.

### Ranged/caster/support flexibility
- **Ranged** abilities may target any living enemy unit (front/back), unless the ability defines additional constraints.
- Support abilities may target allies (including self) regardless of row, unless constrained.

---

## 5) Turn order (initiative contract)

### Turn-based
- Battle is strictly turn-based.

### One unit = one primary turn
- Each unit that is alive and able to act gets **one primary turn** when reached in turn order.

### Skipping non-acting units
- If a unit is dead or disabled, it is **skipped**.

### Stable & robust queue
- The turn queue must remain valid when units die during the battle.
- Tie-breakers must be deterministic.
- The battle loop must safely terminate when battle ends (victory/defeat/retreat).

---

## 6) Actions (minimum early-game set)

Early version must support at least:
- **melee attack**
- **ranged attack**
- **heal**
- **defend**
- **skip / wait**

Architecture must be explicitly extensible to:
- buff
- debuff
- AoE
- DoT
- revive
- summon

Contract notes:
- “defend” should be modeled as a status/effect (e.g. reduce incoming damage until next turn).
- “skip/wait” is a valid action that consumes the unit’s turn without effects.

---

## 7) Targeting rules (validation contract)

### Abilities may or may not require a target
- Some abilities are self-only or no-target (e.g. defend).

### Target kinds
Ability can target:
- enemy unit
- ally unit
- self
- slot/position
- group (row, whole side, etc.) in future AoE rules

### Validation belongs to battle rules, not UI
- Target validity is determined by **battle rule validation**, not by UI.
- UI is a client: it asks for available targets and sends a chosen target.

### Shared validation for player and AI
- Player input and AI must use the **same target validation layer**.
- AI should not “cheat” by bypassing validation.

---

## 8) Battle lifecycle (flow contract)

### Flow
1. Overworld/game detects battle trigger and builds an `Encounter`.
2. Game passes `Encounter` + player party snapshot into battle.
3. Battle builds initial `BattleState` (units, sides, slots, turn order).
4. Battle executes a turn loop:
   - select active unit (initiative order);
   - request action (player input or AI);
   - validate action/target;
   - resolve ability → produce effects → apply to state;
   - check battle end conditions;
   - advance to next unit / round.
5. Battle finishes with an outcome and returns `BattleResult`.
6. Game/world applies `BattleResult` to persistent state and exits battle mode.

---

## 9) Data passed from overworld to battle (domain boundary)

At minimum:
- **Player side composition**:
  - party members (templates or persistent units snapshot);
  - their base stats + equipment-derived modifiers (if applicable);
  - their ability sets.
- **Enemy side composition**:
  - enemy templates or encounter composition; ideally no direct coupling to world entity internals.
- **Encounter metadata** (optional early, but contract-ready):
  - zone/biome info;
  - encounter type (ambush, elite, boss);
  - battle modifiers (weather, terrain tags).
- **Persistent player snapshot**:
  - enough data to build battle units deterministically;
  - battle should not directly mutate persistent player objects.

---

## 10) Data returned from battle (BattleResult contract)

Must include:
- **outcome**: victory/defeat/retreat/other
- **surviving units** (player side) and their post-battle state (HP, statuses if persistent)
- **defeated enemies** (identity/templates for removal and rewards)
- **xp** (if progression exists)
- **rewards/loot/resources**
- **persistent state deltas** (optional/explicit): injuries, permanent buffs/debuffs, etc.

---

## 11) Current code vs target model (gap analysis)

This section reflects the current implementation so future work can be planned safely.

### What already matches the target model
- **Battle as separate mode**: `internal/game` switches `ModeExplore` ↔ `ModeBattle`, battle updates in its own loop.
- **Encounter boundary exists**: `internal/battle/encounter.go` builds `Encounter` from world without mutating the world.
- **Two sides exist**: `TeamPlayer` / `TeamEnemy`.
- **Front/back rows exist**: `RowFront` / `RowBack`, with formation rules:
  - melee restricted to enemy front row when alive (`ReachableEnemyTargets`).
- **Initiative-based turn order exists**: `BuildTurnOrder` sorts by initiative with deterministic tie-breakers.
- **Resolution is centralized**: `ResolveAbility` is the single place that mutates HP/modifiers (good direction).
- **Outcome boundary exists**: internal `Result` mapped to external `BattleOutcome`, applied back to world in `resolveBattleResult`.

### What is simplified / temporary
- **Player side spawns one unit per active party member** (up to front+back capacity) from `party.Party` → `PlayerCombatSeeds()` in `BuildBattleContextFromEncounter`; slots map via `PlayerSlotForPartyIndex`.
  - Command UX is still per–active-unit (turn order); no party-level command mode.
  - Persistent post-battle sync of HP/injuries back into `hero.Hero` is still not implemented.
- **Player action selection is stubbed**:
  - player presses SPACE to execute the *first available ability* with the *first reachable target*.
  - no action menu, no target selection UI.
- **Enemy AI is trivial**:
  - chooses first ability with first reachable target.
- **Statuses are minimal**:
  - only `AttackModifier` exists (no duration, no stacking rules, no disable/stun).
- **Action set is incomplete**:
  - no explicit defend, skip/wait inside battle actions (only “retreat” via ESC).
- **Effects model is not yet explicit**:
  - `ActionResult` is a very small result struct; no effect queue/log/event stream, no multi-target resolution.
- **BattleResult is incomplete**:
  - world enemies removed on victory; rewards on leader; **hero.CurrentHP** synced from battle via `PartyActiveIndex` after each resolved outcome (minimal party persistence).

### Missing for the target contract
- Canonical **side/slot model** (fixed slot sets per row, empty slots, stable Position IDs).
- Canonical **CombatUnit** model detached from world and capable of party members + enemy templates.
- Action selection UX and a **player turn state machine** (choose unit/ability/target, confirm/cancel).
- Unified **target validation API** (enumerate valid targets; validate action request).
- Status/effect framework (durations, hooks, “defend”, disable, DoT, etc.).
- Reward/progression pipeline returned as `BattleResult`.

---

## 12) Next technical steps (ordered)

1. **Canonical combat unit model**
   - define `CombatUnit` seed/snapshot model for both player party members and enemies;
   - ensure battle runtime units are built from seeds (no direct world coupling).

2. **Side + slot model**
   - define fixed formation slots (front/back) with stable `PositionID`;
   - map units to slots; update targeting to reference positions (unit or slot as needed).

3. **Abilities / effects normalization**
   - introduce an explicit `Effect` list/event stream produced by resolution;
   - support multi-effect/multi-target outcomes without ad-hoc special cases.

4. **Player turn state machine**
   - implement action selection flow (ability → target → confirm/cancel);
   - keep it decoupled from rendering and input mapping.

5. **Target validation layer**
   - provide `ListValidTargets(actionRequest)` and `ValidateAction(actionRequest)` used by both UI and AI.

6. **AI upgrade path**
   - keep AI as a client of the same action/target APIs;
   - incrementally improve heuristics without breaking validation.

