package ui

import "testing"

func TestSquirePortraitImage_embeddedDecodes(t *testing.T) {
	img := SquirePortraitImage()
	if img == nil {
		t.Fatal("embedded squire portrait should decode")
	}
	if img.Bounds().Dx() < 8 || img.Bounds().Dy() < 8 {
		t.Fatalf("unexpected bounds %v", img.Bounds())
	}
}
