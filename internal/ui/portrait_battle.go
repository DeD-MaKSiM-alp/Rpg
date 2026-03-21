package ui

import (
	"bytes"
	_ "embed"
	"image/jpeg"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed data/squire_placeholder.jpg
var embeddedSquirePortraitJPEG []byte

var (
	squirePortraitOnce sync.Once
	squirePortraitImg  *ebiten.Image
)

// SquirePortraitImage возвращает загруженный портрет «оруженосец» или nil при ошибке декодирования.
func SquirePortraitImage() *ebiten.Image {
	squirePortraitOnce.Do(func() {
		if len(embeddedSquirePortraitJPEG) == 0 {
			return
		}
		img, err := jpeg.Decode(bytes.NewReader(embeddedSquirePortraitJPEG))
		if err != nil {
			return
		}
		squirePortraitImg = ebiten.NewImageFromImage(img)
	})
	return squirePortraitImg
}
