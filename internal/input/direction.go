package input

// Direction представляет намерение игрока двигаться по сетке.
// Dx — смещение по горизонтали (‑1, 0, 1),
// Dy — смещение по вертикали (‑1, 0, 1).
type Direction struct {
	Dx int
	Dy int
}
