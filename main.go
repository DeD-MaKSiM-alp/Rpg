package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"

	"mygame/world"
)

// main — точка входа в приложение.
// Здесь мы настраиваем окно, создаём экземпляр Game
// и передаём его в ebiten, который берёт на себя главный цикл игры.
func main() {
	// устанавливаем размер окна игры;
	// значения screenWidth/screenHeight берутся из игровых констант в game.go
	ebiten.SetWindowSize(screenWidth, screenHeight)
	// устанавливаем заголовок окна игры — то, что отображается в рамке окна ОС
	ebiten.SetWindowTitle("My Game")

	// worldSeed — зерно процедурной генерации мира.
	// Меняя это число, можно быстро получать разные варианты карты,
	// не меняя остальной код игры.
	worldSeed := 2

	// создаём структуру Game, которая хранит всё состояние игры
	game := &Game{
		// создаём игрока и ставим его в клетку (2, 2) сетки
		player:  *NewPlayer(2, 2),
		world:   world.NewWorld(worldSeed),
		hudFace: text.NewGoXFace(basicfont.Face7x13),
	}
	// Запускаем основной цикл игры.
	// Если Ebiten вернёт ошибку, логируем её и завершаем программу.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
