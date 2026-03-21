package battle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// BattleKeyboardIntents — семантические действия клавиатуры за один кадр боя (после edge detection).
// Единственное место, где сочетаются физические клавиши в UX-смыслы; доменный код опирается только на эти флаги.
type BattleKeyboardIntents struct {
	Confirm bool // Space или Enter
	Back    bool // Backspace (отмена спецспособности / выход из выбора цели)
	Prev    bool // стрелка вверх или влево
	Next    bool // стрелка вниз или вправо
	Escape  bool
}

// BuildBattleKeyboardIntents собирает интенты из флагов «клавиша только что нажата».
// Используется PollBattleKeyboardIntents; экспортировано для юнит-тестов без Ebiten.
func BuildBattleKeyboardIntents(space, enter, backspace, up, left, down, right, esc bool) BattleKeyboardIntents {
	return BattleKeyboardIntents{
		Confirm: space || enter,
		Back:    backspace,
		Prev:    up || left,
		Next:    down || right,
		Escape:  esc,
	}
}

// PollBattleKeyboardIntents читает raw input Ebiten один раз за кадр и возвращает интенты.
func PollBattleKeyboardIntents() BattleKeyboardIntents {
	return BuildBattleKeyboardIntents(
		inpututil.IsKeyJustPressed(ebiten.KeySpace),
		inpututil.IsKeyJustPressed(ebiten.KeyEnter),
		inpututil.IsKeyJustPressed(ebiten.KeyBackspace),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowDown),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowRight),
		inpututil.IsKeyJustPressed(ebiten.KeyEscape),
	)
}

// BattleMouseButtons — нажатия кнопок мыши за кадр (edge). Координаты курсора по-прежнему читаются там, где нужен hit-test.
type BattleMouseButtons struct {
	LeftJustPressed  bool
	RightJustPressed bool
}

// BuildBattleMouseButtons собирает кадр из флагов edge (для тестов; pollBattleMouseButtons использует Ebiten).
func BuildBattleMouseButtons(leftJust, rightJust bool) BattleMouseButtons {
	return BattleMouseButtons{LeftJustPressed: leftJust, RightJustPressed: rightJust}
}

func pollBattleMouseButtons() BattleMouseButtons {
	return BuildBattleMouseButtons(
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight),
	)
}
