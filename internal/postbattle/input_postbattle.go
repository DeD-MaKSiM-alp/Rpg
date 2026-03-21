package postbattle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PostBattleKeyboardIntents — семантика клавиатуры за кадр post-battle (edge detection уже учтена в poll).
type PostBattleKeyboardIntents struct {
	Confirm bool // Space или Enter — продолжить результат / подтвердить выбор награды
	Prev    bool // стрелка вверх или влево — предыдущая награда
	Next    bool // стрелка вниз или вправо — следующая награда
}

// BuildPostBattleKeyboardIntents собирает интенты из флагов «клавиша только что нажата» (для тестов без Ebiten).
func BuildPostBattleKeyboardIntents(space, enter, up, left, down, right bool) PostBattleKeyboardIntents {
	return PostBattleKeyboardIntents{
		Confirm: space || enter,
		Prev:    up || left,
		Next:    down || right,
	}
}

// PollPostBattleKeyboardIntents читает raw input один раз за кадр.
func PollPostBattleKeyboardIntents() PostBattleKeyboardIntents {
	return BuildPostBattleKeyboardIntents(
		inpututil.IsKeyJustPressed(ebiten.KeySpace),
		inpututil.IsKeyJustPressed(ebiten.KeyEnter),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowDown),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowRight),
	)
}

// PostBattleMouseButtons — нажатия кнопок мыши за кадр (edge). Координаты и hit-test остаются в Update.
type PostBattleMouseButtons struct {
	LeftJustPressed  bool
	RightJustPressed bool
}

// BuildPostBattleMouseButtons собирает кадр из флагов (для тестов).
func BuildPostBattleMouseButtons(leftJust, rightJust bool) PostBattleMouseButtons {
	return PostBattleMouseButtons{LeftJustPressed: leftJust, RightJustPressed: rightJust}
}

func pollPostBattleMouseButtons() PostBattleMouseButtons {
	return BuildPostBattleMouseButtons(
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight),
	)
}
