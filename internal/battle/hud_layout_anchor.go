package battle

// Привязка battle HUD к той же экранной модели, что и ui.ScreenLayout (safe inset + tier).
// Логика pad/tier дублирует ui.TierFromScreen / presetForTier — при изменении foundation синхронизировать.

func offsetHUDRect(r HUDRect, dx, dy float32) HUDRect {
	return HUDRect{X: r.X + dx, Y: r.Y + dy, W: r.W, H: r.H}
}

// battleContentRectFromScreen — safe-фрейм (аналог ui.ScreenLayout.Safe).
func battleContentRectFromScreen(screenW, screenH int) (x, y, w, h float32) {
	sw := float32(screenW)
	sh := float32(screenH)
	minDim := screenW
	if screenH < minDim {
		minDim = screenH
	}
	var pad float32 = 12
	switch {
	case minDim < 720 || screenW < 960:
		pad = 8
	case minDim >= 900 && screenW >= 1440:
		pad = 16
	default:
		pad = 12
	}
	return pad, pad, sw - 2*pad, sh - 2*pad
}

// battleTierOrdinal — 0 small, 1 medium, 2 large (совпадает с ui.ResolutionTier порядком).
func battleTierOrdinal(screenW, screenH int) int {
	minDim := screenW
	if screenH < minDim {
		minDim = screenH
	}
	if minDim < 720 || screenW < 960 {
		return 0
	}
	if minDim < 900 || screenW < 1440 {
		return 1
	}
	return 2
}

// OffsetBattleHUDLayout сдвигает все прямоугольники layout (draw + hit-test).
func OffsetBattleHUDLayout(l BattleHUDLayout, dx, dy float32) BattleHUDLayout {
	o := l
	o.TopBar = offsetHUDRect(l.TopBar, dx, dy)
	o.LeftRoster = offsetHUDRect(l.LeftRoster, dx, dy)
	o.RightRoster = offsetHUDRect(l.RightRoster, dx, dy)
	o.Battlefield = offsetHUDRect(l.Battlefield, dx, dy)
	o.BottomPanel = offsetHUDRect(l.BottomPanel, dx, dy)
	o.V2TopBarLine = offsetHUDRect(l.V2TopBarLine, dx, dy)
	o.V2BottomActive = offsetHUDRect(l.V2BottomActive, dx, dy)
	o.V2BottomTarget = offsetHUDRect(l.V2BottomTarget, dx, dy)
	o.V2BottomSummary = offsetHUDRect(l.V2BottomSummary, dx, dy)
	o.V2BottomLog = offsetHUDRect(l.V2BottomLog, dx, dy)
	o.Overlay = offsetHUDRect(l.Overlay, dx, dy)
	o.Content = offsetHUDRect(l.Content, dx, dy)
	o.TitleRow = offsetHUDRect(l.TitleRow, dx, dy)
	o.InfoRow1 = offsetHUDRect(l.InfoRow1, dx, dy)
	o.InfoRow2 = offsetHUDRect(l.InfoRow2, dx, dy)
	o.Formation = offsetHUDRect(l.Formation, dx, dy)
	o.Middle = offsetHUDRect(l.Middle, dx, dy)
	o.Footer = offsetHUDRect(l.Footer, dx, dy)
	o.PlayerFormation = offsetHUDRect(l.PlayerFormation, dx, dy)
	o.EnemyFormation = offsetHUDRect(l.EnemyFormation, dx, dy)
	o.Abilities = offsetHUDRect(l.Abilities, dx, dy)
	o.Action = offsetHUDRect(l.Action, dx, dy)
	o.AbilitiesTitleRow = offsetHUDRect(l.AbilitiesTitleRow, dx, dy)
	o.AbilityList = offsetHUDRect(l.AbilityList, dx, dy)
	o.ActionTitleRow = offsetHUDRect(l.ActionTitleRow, dx, dy)
	o.ActionSummary = offsetHUDRect(l.ActionSummary, dx, dy)
	o.ActionButtons = offsetHUDRect(l.ActionButtons, dx, dy)
	o.FooterTitleRow = offsetHUDRect(l.FooterTitleRow, dx, dy)
	o.CombatLog = offsetHUDRect(l.CombatLog, dx, dy)
	o.HintLine = offsetHUDRect(l.HintLine, dx, dy)
	o.PlayerFormationTitleRow = offsetHUDRect(l.PlayerFormationTitleRow, dx, dy)
	o.EnemyFormationTitleRow = offsetHUDRect(l.EnemyFormationTitleRow, dx, dy)
	o.BackButton = offsetHUDRect(l.BackButton, dx, dy)
	o.TopInfoPrimary = offsetHUDRect(l.TopInfoPrimary, dx, dy)
	o.TopInfoSecondary = offsetHUDRect(l.TopInfoSecondary, dx, dy)
	o.InfoLine = offsetHUDRect(l.InfoLine, dx, dy)
	o.AbilityHeader = offsetHUDRect(l.AbilityHeader, dx, dy)
	o.AbilityTooltip = offsetHUDRect(l.AbilityTooltip, dx, dy)
	o.ActionMain = offsetHUDRect(l.ActionMain, dx, dy)
	o.ActorInfo = offsetHUDRect(l.ActorInfo, dx, dy)
	o.HoverInfo = offsetHUDRect(l.HoverInfo, dx, dy)
	o.BattlefieldPlayerBack = offsetHUDRect(l.BattlefieldPlayerBack, dx, dy)
	o.BattlefieldPlayerFront = offsetHUDRect(l.BattlefieldPlayerFront, dx, dy)
	o.BattlefieldEnemyFront = offsetHUDRect(l.BattlefieldEnemyFront, dx, dy)
	o.BattlefieldEnemyBack = offsetHUDRect(l.BattlefieldEnemyBack, dx, dy)

	if len(l.AbilityItemRects) > 0 {
		o.AbilityItemRects = make([]HUDRect, len(l.AbilityItemRects))
		for i := range l.AbilityItemRects {
			o.AbilityItemRects[i] = offsetHUDRect(l.AbilityItemRects[i], dx, dy)
		}
	}
	if len(l.BattlefieldSlotCells) > 0 {
		o.BattlefieldSlotCells = make([]HUDRect, len(l.BattlefieldSlotCells))
		for i := range l.BattlefieldSlotCells {
			o.BattlefieldSlotCells[i] = offsetHUDRect(l.BattlefieldSlotCells[i], dx, dy)
		}
	}
	if len(l.BattlefieldRowLabels) > 0 {
		o.BattlefieldRowLabels = make([]BattlefieldRowLabel, len(l.BattlefieldRowLabels))
		for i := range l.BattlefieldRowLabels {
			o.BattlefieldRowLabels[i] = BattlefieldRowLabel{
				Rect: offsetHUDRect(l.BattlefieldRowLabels[i].Rect, dx, dy),
				Text: l.BattlefieldRowLabels[i].Text,
			}
		}
	}
	if l.UnitRects != nil {
		o.UnitRects = make(map[UnitID]HUDRect, len(l.UnitRects))
		for k, v := range l.UnitRects {
			o.UnitRects[k] = offsetHUDRect(v, dx, dy)
		}
	}
	if l.BattlefieldTokens != nil {
		o.BattlefieldTokens = make(map[UnitID]HUDRect, len(l.BattlefieldTokens))
		for k, v := range l.BattlefieldTokens {
			o.BattlefieldTokens[k] = offsetHUDRect(v, dx, dy)
		}
	}
	return o
}

// ComputeBattleHUDLayoutWithBounds строит layout в прямоугольнике контента (cw×ch), затем сдвигает на (cx,cy).
// fullW/fullH — полный размер окна для computeHUDMetrics (масштаб читаемости); tierHUD — плотность текста (см. computeHUDMetrics).
func (b *BattleContext) ComputeBattleHUDLayoutWithBounds(fullW, fullH int, cx, cy, cw, ch float32, tierHUD int) BattleHUDLayout {
	iw, ih := int(cw), int(ch)
	if iw < 1 {
		iw = 1
	}
	if ih < 1 {
		ih = 1
	}
	metrics := computeHUDMetrics(fullW, fullH, tierHUD)
	var lay BattleHUDLayout
	if b != nil && b.LayoutStyle == LayoutStyleV2Disciples {
		lay = b.computeLayoutV2(iw, ih, metrics)
	} else {
		lay = b.computeLayoutV1(iw, ih, metrics)
	}
	return OffsetBattleHUDLayout(lay, cx, cy)
}

// ComputeBattleHUDLayoutAnchored — раскладка внутри safe-зоны (как ui.ScreenLayout) + tier-aware metrics.
// Используйте для отрисовки и hit-test вместо сырого полноэкранного ComputeBattleHUDLayout, чтобы HUD совпадал с foundation.
func (b *BattleContext) ComputeBattleHUDLayoutAnchored(screenW, screenH int) BattleHUDLayout {
	cx, cy, cw, ch := battleContentRectFromScreen(screenW, screenH)
	tier := battleTierOrdinal(screenW, screenH)
	return b.ComputeBattleHUDLayoutWithBounds(screenW, screenH, cx, cy, cw, ch, tier)
}

// BattleContentRectForHUDAnchor — safe-фрейм контента для anchored battle HUD (совпадает с ui.ScreenLayout.Safe при тех же W/ H).
func BattleContentRectForHUDAnchor(screenW, screenH int) (x, y, w, h float32) {
	return battleContentRectFromScreen(screenW, screenH)
}

// BattleHUDTierOrdinal — 0 = small, 1 = medium, 2 = large (порядок как у ui.ResolutionTier).
func BattleHUDTierOrdinal(screenW, screenH int) int {
	return battleTierOrdinal(screenW, screenH)
}
