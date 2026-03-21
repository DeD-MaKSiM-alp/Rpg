package game

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/party"
	playerpkg "mygame/internal/player"
	"mygame/internal/progression"
	"mygame/internal/ui"
	"mygame/internal/unitdata"
	"mygame/world"
)

// exploreRestFeedbackDurationTicks — длительность баннера после отдыха (R) в explore (~3s @ 60fps).
const exploreRestFeedbackDurationTicks = 180

// formationMsgDurationTicks — длительность баннера promotion на экране состава.
const formationMsgDurationTicks = 180

// Update обрабатывает один кадр игры.
func (g *Game) Update() error {
	if g.mode != ModeFormation {
		g.inspectHoverFormationGlobalIdx = -1
	}
	if g.mode != ModeBattle {
		g.inspectHoverBattleUnitID = 0
	}

	// Runtime resolution switch: F6 = previous preset, F7 = next preset (cyclic).
	n := len(ResolutionPresets)
	if n > 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyF6) {
			ActivePresetIndex = (ActivePresetIndex - 1 + n) % n
			applyResolutionPreset()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF7) {
			ActivePresetIndex = (ActivePresetIndex + 1) % n
			applyResolutionPreset()
		}
	}

	if g.mode == ModeBattle {
		if inpututil.IsKeyJustPressed(ebiten.KeyF8) {
			if g.BattleHUDStyle == 0 {
				g.BattleHUDStyle = 1
			} else {
				g.BattleHUDStyle = 0
			}
		}
		g.updateBattleMode()
		return nil
	}
	if g.mode == ModeFormation {
		g.updateFormationMode()
		return nil
	}
	// Recruit offer — отдельная мода: иначе updateExploreMode съедает ввод и не вызывается accept (см. recruit_camp_accept_bugfix_report).
	if g.mode == ModeRecruitOffer {
		g.updateRecruitOfferMode()
		return nil
	}
	return g.updateExploreMode()
}

// readPlayerAction — единственное место чтения explore input; контракт: Input.ReadExploreInput().
func (g *Game) readPlayerAction() PlayerAction {
	dx, dy, wait := g.input.ReadExploreInput()
	g.debugInputDX, g.debugInputDY = dx, dy // временный debug: для отрисовки "Input: dx= dy="
	if wait {
		return PlayerAction{Type: ActionWait}
	}
	if dx != 0 || dy != 0 {
		return PlayerAction{Type: ActionMove, DX: dx, DY: dy}
	}
	return PlayerAction{Type: ActionNone}
}

// advanceWorldTurn — единственная точка вызова AdvanceTurn: ход врагов, затем обновление камеры и стриминга.
func (g *Game) advanceWorldTurn() {
	px, py := g.player.Position()
	enemyID, startedBattle := g.world.AdvanceTurn(px, py)
	if startedBattle && enemyID != 0 {
		g.startBattle(enemyID)
		return
	}
	g.updateCamera()
	g.updateStreamingWorld()
}

// updateExploreMode: Input → PlayerAction → применение действия → при успехе завершение хода → world turn → возможный бой.
func (g *Game) updateExploreMode() error {
	if g.exploreRestMsgTicks > 0 {
		g.exploreRestMsgTicks--
		if g.exploreRestMsgTicks <= 0 {
			g.exploreRestMsg = ""
		}
	}
	if g.exploreRecruitMsgTicks > 0 {
		g.exploreRecruitMsgTicks--
		if g.exploreRecruitMsgTicks <= 0 {
			g.exploreRecruitMsg = ""
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		if g.party.TotalMembers() >= party.MaxPartyMembers {
			g.exploreRecruitMsg = fmt.Sprintf("Отряд полон (макс. %d)", party.MaxPartyMembers)
			g.exploreRecruitMsgTicks = exploreRestFeedbackDurationTicks
			return nil
		}
		idx := len(g.party.Reserve) + 1
		h := hero.RecruitHeroFromEarlyPool(idx)
		h.RecruitLabel = hero.RecruitDisplayName(idx)
		if err := g.party.AddToReserve(h); err != nil {
			if errors.Is(err, party.ErrPartyFull) {
				g.exploreRecruitMsg = fmt.Sprintf("Отряд полон (макс. %d)", party.MaxPartyMembers)
			} else {
				g.exploreRecruitMsg = err.Error()
			}
		} else {
			g.exploreRecruitMsg = fmt.Sprintf("В резерв: %s (F5 — состав)", hero.RecruitDisplayName(idx))
		}
		g.exploreRecruitMsgTicks = exploreRestFeedbackDurationTicks
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		g.mode = ModeFormation
		g.formationSel = 0
		g.formationInspectOpen = false
		if n := len(g.party.Active); n == 0 {
			g.mode = ModeExplore
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if len(g.party.Active) > 0 {
			g.party.ApplyWorldRest()
			g.exploreRestMsg = party.RestExploreBanner
			g.exploreRestMsgTicks = exploreRestFeedbackDurationTicks
			g.advanceWorldTurn()
		}
		return nil
	}

	action := g.readPlayerAction()

	if action.Type == ActionNone {
		g.updateCamera()
		g.updateStreamingWorld()
		return nil
	}

	switch action.Type {
	case ActionMove:
		moved, enemyID, pu := playerpkg.TryMovePlayer(&g.player, g.world, action.DX, action.DY)
		if pu == world.PickupInteractResource {
			g.pickupCount++
		}
		if !moved {
			return nil
		}
		if enemyID != 0 {
			g.startBattle(enemyID)
			return nil
		}
		if pu == world.PickupInteractRecruitOffer {
			g.mode = ModeRecruitOffer
			g.recruitOfferX = g.player.GridX
			g.recruitOfferY = g.player.GridY
			return nil
		}
		g.advanceWorldTurn()

	case ActionWait:
		g.advanceWorldTurn()
	}

	return nil
}

func (g *Game) updateFormationMode() {
	mx, my := ebiten.CursorPosition()
	g.inspectHoverFormationGlobalIdx = ui.FormationHitTestGlobalIndex(ScreenWidth, ScreenHeight, mx, my, &g.party)

	if g.formationMsgTicks > 0 {
		g.formationMsgTicks--
		if g.formationMsgTicks <= 0 {
			g.formationMsg = ""
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		if gidx := ui.FormationHitTestGlobalIndex(ScreenWidth, ScreenHeight, mx, my, &g.party); gidx >= 0 {
			if g.formationInspectOpen && gidx == g.formationSel {
				g.formationInspectOpen = false
			} else {
				g.formationSel = gidx
				g.formationInspectOpen = true
				g.syncPromotionBranchForInspectHero()
			}
			return
		}
	}

	na := len(g.party.Active)
	nr := len(g.party.Reserve)
	total := na + nr
	if total == 0 {
		g.mode = ModeExplore
		g.formationInspectOpen = false
		return
	}
	if g.formationSel >= total {
		g.formationSel = total - 1
	}
	if g.formationSel < 0 {
		g.formationSel = 0
	}

	// Карточка бойца: только навигация по списку и выход с листа.
	if g.formationInspectOpen {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyI) ||
			inpututil.IsKeyJustPressed(ebiten.KeyF5) {
			g.formationInspectOpen = false
			return
		}
		h := g.party.HeroAtGlobalIndex(g.formationSel)
		if h != nil {
			targets, _ := hero.PromotionTargetUnitIDs(h)
			if len(targets) >= 2 {
				if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
					g.formationPromoteBranchIdx = 0
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
					g.formationPromoteBranchIdx = 1
				}
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			if h != nil {
				atCamp := g.world.PlayerStandsOnActiveRecruitCamp(g.player.GridX, g.player.GridY)
				targets, _ := hero.PromotionTargetUnitIDs(h)
				var sel string
				if len(targets) == 1 {
					sel = targets[0]
				} else if len(targets) >= 2 {
					if g.formationPromoteBranchIdx < 0 {
						sel = ""
					} else {
						sel = targets[g.formationPromoteBranchIdx%len(targets)]
					}
				}
				gate := EvaluatePromotionGate(h, atCamp, g.TrainingMarks, sel)
				if !gate.Allowed {
					g.formationMsg = gate.Message
				} else if err := hero.TryPromoteHeroTo(h, sel); err != nil {
					g.formationMsg = hero.PromotionErrUserMessage(err)
				} else {
					g.TrainingMarks -= gate.Cost
					if tpl, ok := unitdata.GetUnitTemplate(h.UnitID); ok {
						g.formationMsg = "Повышение: «" + tpl.DisplayName + "»"
					} else {
						g.formationMsg = "Повышение выполнено."
					}
				}
				g.formationMsgTicks = formationMsgDurationTicks
			}
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && g.formationSel > 0 {
			g.formationSel--
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && g.formationSel < total-1 {
			g.formationSel++
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		g.mode = ModeExplore
		g.formationInspectOpen = false
		g.formationMsg = ""
		g.formationMsgTicks = 0
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.formationInspectOpen = true
		g.syncPromotionBranchForInspectHero()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && g.formationSel > 0 {
		g.formationSel--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && g.formationSel < total-1 {
		g.formationSel++
	}

	if na >= 2 && g.formationSel < na {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			if g.party.MoveActiveEarlier(g.formationSel) {
				g.formationSel--
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			if g.party.MoveActiveLater(g.formationSel) {
				g.formationSel++
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.formationSel < na {
			if g.party.MoveActiveToReserve(g.formationSel) {
				if g.formationSel >= len(g.party.Active) {
					g.formationSel = len(g.party.Active) - 1
				}
			}
		} else {
			rj := g.formationSel - na
			if g.party.MoveReserveToActive(rj) {
				g.formationSel = len(g.party.Active) - 1
			}
		}
	}
}

func (g *Game) updateBattleMode() {
	if g.battle == nil {
		g.endBattle()
		return
	}

	// Post-battle flow: result screen → (on victory) reward selection → return to world.
	if g.postBattle.IsActive() {
		g.inspectHoverBattleUnitID = 0
		if g.postBattle.Update(&g.party, ScreenWidth, ScreenHeight) {
			g.endBattle()
		}
		return
	}

	mx, my := ebiten.CursorPosition()
	g.inspectHoverBattleUnitID = battlepkg.HitTestUnitUnderCursor(g.battle, ScreenWidth, ScreenHeight, mx, my)

	g.battle.LayoutStyle = g.BattleHUDStyle
	g.battle.SuppressEscThisFrame = false
	g.battle.SuppressMouseRightThisFrame = false
	g.battle.BlockPlayerInput = g.battleInspectOpen

	if g.battleInspectOpen && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.battleInspectOpen = false
		g.battleInspectUnitID = 0
		g.battle.SuppressEscThisFrame = true
	}

	g.handleBattleInspectInput()

	if g.battleInspectOpen {
		u := g.battle.Units[g.battleInspectUnitID]
		if u != nil && u.Origin.PartyActiveIndex >= 0 {
			if h := g.party.HeroAtGlobalIndex(u.Origin.PartyActiveIndex); h != nil {
				targets, _ := hero.PromotionTargetUnitIDs(h)
				if len(targets) >= 2 {
					if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
						g.formationPromoteBranchIdx = 0
					}
					if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
						g.formationPromoteBranchIdx = 1
					}
				}
			}
		}
	}

	outcome := g.battle.Update()

	switch outcome {
	case battlepkg.BattleOutcomeVictory:
		g.syncPartyFromBattle()
		progression.ApplyVictoryCombatXPForActiveSurvivors(g.battle, &g.party)
		summary := progression.BuildVictoryProgressionSummary(g.battle, &g.party, TrainingMarksPerVictory)
		g.resolveBattleResult(outcome)
		g.BattlesWon++
		g.applyVictoryTrainingMarks()
		g.postBattle.Begin(outcome, summary.Lines)
		return
	case battlepkg.BattleOutcomeDefeat:
		g.syncPartyFromBattle()
		g.resolveBattleResult(outcome)
		g.postBattle.Begin(outcome, nil)
		return
	case battlepkg.BattleOutcomeRetreat:
		g.syncPartyFromBattle()
		g.resolveBattleResult(outcome)
		g.postBattle.Begin(outcome, nil)
		return
	case battlepkg.BattleOutcomeNone:
		return
	}
}

// resolveBattleResult применяет результат боя к миру (удаление врагов при победе и т.д.).
func (g *Game) handleBattleInspectInput() {
	if g.battle == nil || !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		return
	}
	mx, my := ebiten.CursorPosition()
	if hit := battlepkg.HitTestUnitUnderCursor(g.battle, ScreenWidth, ScreenHeight, mx, my); hit != 0 {
		if g.battleInspectOpen && hit == g.battleInspectUnitID {
			g.battleInspectOpen = false
			g.battleInspectUnitID = 0
		} else {
			g.battleInspectOpen = true
			g.battleInspectUnitID = hit
			g.syncPromotionBranchForBattleInspect()
		}
		g.battle.SuppressMouseRightThisFrame = true
		return
	}
	if g.battleInspectOpen {
		g.battleInspectOpen = false
		g.battleInspectUnitID = 0
		g.battle.SuppressMouseRightThisFrame = true
	}
}

func (g *Game) resolveBattleResult(outcome battlepkg.BattleOutcome) {
	switch outcome {
	case battlepkg.BattleOutcomeVictory:
		for _, e := range g.battle.Encounter.Enemies {
			g.world.RemoveEnemy(e.EnemyID)
		}
	case battlepkg.BattleOutcomeDefeat, battlepkg.BattleOutcomeRetreat:
		// Пока ничего не делаем; позже: respawn, потеря прогресса и т.д.
	}
}
