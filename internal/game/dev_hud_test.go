package game

import "testing"

func TestDevHUDOverlayDefaultOff(t *testing.T) {
	if DevHUDOverlay {
		t.Fatal("DevHUDOverlay должен быть false по умолчанию (служебный оверлей только после F10)")
	}
}
