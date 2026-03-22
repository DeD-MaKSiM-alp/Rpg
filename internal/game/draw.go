package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	battlepkg "mygame/internal/battle"
	"mygame/internal/hero"
	"mygame/internal/postbattle"
	"mygame/internal/ui"
)

// Draw рисует один кадр игры.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	g.world.Draw(screen, g.cameraX, g.cameraY, WorldViewport.WidthTiles, WorldViewport.HeightTiles, tileSize)
	g.drawGrid(screen)
	if g.mode == ModeExplore {
		g.world.DrawExploreCues(screen, g.player.GridX, g.player.GridY, g.cameraX, g.cameraY, WorldViewport.WidthTiles, WorldViewport.HeightTiles, tileSize)
	}

	if debugShowChunkOverlay {
		viewportW := WorldViewport.WidthTiles * tileSize
		viewportH := WorldViewport.HeightTiles * tileSize
		g.world.DrawChunkDebugOverlay(screen, g.cameraX, g.cameraY, WorldViewport.WidthTiles, WorldViewport.HeightTiles, tileSize, viewportW, viewportH)
	}

	g.player.Draw(screen, g.cameraX, g.cameraY, tileSize)

	if debugShowChunkOverlay {
		g.drawDebugInfo(screen)
	}

	atCamp := g.world.PlayerStandsOnActiveRecruitCamp(g.player.GridX, g.player.GridY)
	promoHUD := PromotionExploreHUDLine(&g.party, atCamp, g.TrainingMarks)
	ui.DrawHUD(screen, g.pickupCount, g.TrainingMarks, g.hudFace, g.party.Leader(), ScreenWidth, promoHUD)

	if g.mode == ModeExplore || g.mode == ModeRecruitOffer || g.mode == ModePOIChoice {
		promoStrip := PromotionExploreStripLine(&g.party, atCamp, g.TrainingMarks)
		ui.DrawExplorePartyStrip(screen, g.hudFace, &g.party, ScreenWidth, promoStrip)
	}
	if g.mode == ModeExplore {
		ui.DrawExploreFormationHint(screen, g.hudFace, ScreenWidth, ScreenHeight, g.world.ZoneHUDLine(g.player.GridX, g.player.GridY), g.exploreRestMsg, g.exploreRecruitMsg, g.explorePOIMsg, g.world.ExploreHUDHintLine(g.player.GridX, g.player.GridY))
	}

	if g.mode == ModeRecruitOffer {
		ui.DrawRecruitOfferOverlay(screen, g.hudFace, ScreenWidth, ScreenHeight, g.recruitOfferHoverBtn)
	}
	if g.mode == ModePOIChoice {
		ui.DrawPOIChoiceOverlay(screen, g.hudFace, ScreenWidth, ScreenHeight, g.poiChoiceKind, g.poiChoiceSel, altarBoldHPLossPreview(&g.party), g.poiChoiceHoverOpt, g.poiChoiceHoverConfirm, g.poiChoiceHoverCancel)
	}

	if g.mode == ModeFormation {
		promoHints := PromotionFormationRowHints(&g.party, atCamp, g.TrainingMarks)
		ui.DrawFormationOverlay(screen, g.hudFace, &g.party, g.formationSel, ScreenWidth, ScreenHeight, g.formationInspectOpen, g.inspectHoverFormationGlobalIdx, promoHints)
		if g.formationInspectOpen {
			var promoTargets []string
			promoCosts := []int(nil)
			var promoHead string
			if h := g.party.HeroAtGlobalIndex(g.formationSel); h != nil {
				promoTargets, _ = hero.PromotionTargetUnitIDs(h)
				promoCosts = make([]int, len(promoTargets))
				for i, id := range promoTargets {
					c, ok := PromotionTrainingMarkCostForHeroTarget(h, id)
					if ok {
						promoCosts[i] = c
					}
				}
				promoHead = PromotionInspectHeadline(h, atCamp, g.TrainingMarks, promoTargets, g.formationPromoteBranchIdx)
			}
			ui.DrawCharacterInspectOverlay(screen, g.hudFace, &g.party, g.formationSel, ScreenWidth, ScreenHeight, g.formationMsg, atCamp, g.TrainingMarks, promoTargets, promoCosts, g.formationPromoteBranchIdx, promoHead)
		}
	}

	if g.mode == ModeExplore && debugShowInputDirection {
		rawX, rawY := g.input.DebugRaw()
		emitX, emitY := g.input.DebugEmit()
		ui.DrawDebugInputDirection(screen, rawX, rawY, emitX, emitY)
	}

	if g.mode == ModeBattle {
		ui.DrawBattleOverlay(screen, g.hudFace, g.battle, ScreenWidth, ScreenHeight, g.battleInspectUnitID, g.battleInspectOpen)
		if g.battle != nil && !g.postBattle.IsActive() {
			ui.DrawBattleInspectHighlights(screen, g.battle, ScreenWidth, ScreenHeight, g.inspectHoverBattleUnitID, g.battleInspectOpen, g.battleInspectUnitID)
		}
		if g.postBattle.IsActive() {
			params := postbattle.BuildPostBattleParams(&g.postBattle, ScreenWidth, ScreenHeight)
			ui.DrawPostBattleOverlay(screen, g.hudFace, params)
		} else if g.battleInspectOpen && g.battle != nil {
			u := g.battle.Units[g.battleInspectUnitID]
			if u != nil {
				var promoTargets []string
				promoCosts := []int(nil)
				promoHead := ""
				if u.Side == battlepkg.TeamPlayer && u.Origin.PartyActiveIndex >= 0 {
					if h := g.party.HeroAtGlobalIndex(u.Origin.PartyActiveIndex); h != nil {
						promoTargets, _ = hero.PromotionTargetUnitIDs(h)
						promoCosts = make([]int, len(promoTargets))
						for i, id := range promoTargets {
							c, ok := PromotionTrainingMarkCostForHeroTarget(h, id)
							if ok {
								promoCosts[i] = c
							}
						}
						promoHead = PromotionInspectHeadline(h, false, g.TrainingMarks, promoTargets, g.formationPromoteBranchIdx)
					}
				}
				ui.DrawBattleInspectOverlay(screen, g.hudFace, &g.party, u, ScreenWidth, ScreenHeight, g.TrainingMarks, promoTargets, promoCosts, g.formationPromoteBranchIdx, promoHead)
			}
		}
	}

	ui.DrawResolutionIndicator(screen, g.hudFace, ScreenWidth, ScreenHeight)
}
