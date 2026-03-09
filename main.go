package main

import (
	"log"

	"mygame/internal/game"
)

// Точка входа. Вся инициализация и запуск — через пакет game
// (internal/game подключает internal/battle, ui, input, player и world).
func main() {
	worldSeed := 3
	playerGridX, playerGridY := 2, 2

	if err := game.Run(worldSeed, playerGridX, playerGridY, ""); err != nil {
		log.Fatal(err)
	}
}
