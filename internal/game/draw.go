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

	if debugShowChunkOverlay {
		viewportW := WorldViewport.WidthTiles * tileSize
		viewportH := WorldViewport.HeightTiles * tileSize
		g.world.DrawChunkDebugOverlay(screen, g.cameraX, g.cameraY, WorldViewport.WidthTiles, WorldViewport.HeightTiles, tileSize, viewportW, viewportH)
	}

	g.player.Draw(screen, g.cameraX, g.cameraY, tileSize)

	if debugShowChunkOverlay {
		g.drawDebugInfo(screen)
	}

	ui.DrawHUD(screen, g.pickupCount, g.TrainingMarks, g.hudFace)

	if g.mode == ModeExplore || g.mode == ModeRecruitOffer {
		ui.DrawExplorePartyStrip(screen, g.hudFace, &g.party, ScreenWidth)
	}
	if g.mode == ModeExplore {
		ui.DrawExploreFormationHint(screen, g.hudFace, ScreenWidth, ScreenHeight, g.exploreRestMsg, g.exploreRecruitMsg)
	}

	if g.mode == ModeRecruitOffer {
		ui.DrawRecruitOfferOverlay(screen, g.hudFace, ScreenWidth, ScreenHeight)
	}

	if g.mode == ModeFormation {
		ui.DrawFormationOverlay(screen, g.hudFace, &g.party, g.formationSel, ScreenWidth, ScreenHeight, g.formationInspectOpen, g.inspectHoverFormationGlobalIdx)
		if g.formationInspectOpen {
			atCamp := g.world.PlayerStandsOnActiveRecruitCamp(g.player.GridX, g.player.GridY)
			var promoTargets []string
			promoCosts := []int(nil)
			if h := g.party.HeroAtGlobalIndex(g.formationSel); h != nil {
				promoTargets, _ = hero.PromotionTargetUnitIDs(h)
				promoCosts = make([]int, len(promoTargets))
				for i, id := range promoTargets {
					c, ok := PromotionTrainingMarkCostForHeroTarget(h, id)
					if ok {
						promoCosts[i] = c
					}
				}
			}
			ui.DrawCharacterInspectOverlay(screen, g.hudFace, &g.party, g.formationSel, ScreenWidth, ScreenHeight, g.formationMsg, atCamp, g.TrainingMarks, promoTargets, promoCosts, g.formationPromoteBranchIdx)
		}
	}

	if g.mode == ModeExplore && debugShowInputDirection {
		rawX, rawY := g.input.DebugRaw()
		emitX, emitY := g.input.DebugEmit()
		ui.DrawDebugInputDirection(screen, rawX, rawY, emitX, emitY)
	}

	if g.mode == ModeBattle {
		ui.DrawBattleOverlay(screen, g.hudFace, g.battle, ScreenWidth, ScreenHeight)
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
					}
				}
				ui.DrawBattleInspectOverlay(screen, g.hudFace, &g.party, u, ScreenWidth, ScreenHeight, g.TrainingMarks, promoTargets, promoCosts, g.formationPromoteBranchIdx)
			}
		}
	}

	ui.DrawResolutionIndicator(screen, g.hudFace, ScreenWidth, ScreenHeight)
}
