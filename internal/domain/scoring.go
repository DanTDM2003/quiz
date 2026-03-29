package domain

type ScoreOutcome struct {
	Correct    bool
	ScoreDelta int
}

func ScoreAnswer(q Question, selectedOptionID string) ScoreOutcome {
	correct := selectedOptionID == q.CorrectOptionID
	delta := 0
	if correct {
		delta = q.Points
	}
	return ScoreOutcome{Correct: correct, ScoreDelta: delta}
}
