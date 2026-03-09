package input

// Direction задаёт направление движения (смещение по сетке).
// Нулевое значение Direction{} означает отсутствие направления.
type Direction struct {
	Dx, Dy int
}
