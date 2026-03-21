package game

// applyVictoryTrainingMarks — начисление знаков обучения за победу в бою (сессия).
// Вызывается из updateBattleMode только при BattleOutcomeVictory.
func (g *Game) applyVictoryTrainingMarks() {
	g.TrainingMarks += TrainingMarksPerVictory
}
