package ui

import (
	"bytes"
	"embed"
	"log"

	text "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Встраиваем читаемый UI-шрифт с поддержкой кириллицы.
// Для HUD и battle overlay лучше использовать нейтральный шрифт,
// а не сильно стилизованный ретро-вариант.
//
//go:embed assets/fonts/*.ttf
var hudFontFS embed.FS

// LoadHUDFace создаёт шрифт для HUD и боевого overlay.
//
// Размер 14 лучше подходит для плотного интерфейса:
// текст остаётся читаемым, но не выглядит слишком громоздким.
func LoadHUDFace() *text.GoTextFace {
	hudFontBytes, err := hudFontFS.ReadFile("assets/fonts/JetBrainsMono-Regular.ttf")
	if err != nil {
		log.Fatalf("failed to read HUD font: %v", err)
	}
	source, err := text.NewGoTextFaceSource(bytes.NewReader(hudFontBytes))
	if err != nil {
		log.Fatalf("failed to load HUD font: %v", err)
	}

	return &text.GoTextFace{
		Source: source,
		Size:   14,
	}
}
