package domain

import "testing"

func TestScoreAnswer(t *testing.T) {
	t.Parallel()
	out := ScoreAnswer(Question{ID: "q1", CorrectOptionID: "a", Points: 10}, "a")
	if !out.Correct || out.ScoreDelta != 10 {
		t.Fatalf("expected correct and 10, got %+v", out)
	}
	out = ScoreAnswer(Question{ID: "q1", CorrectOptionID: "a", Points: 10}, "b")
	if out.Correct || out.ScoreDelta != 0 {
		t.Fatalf("expected incorrect and 0, got %+v", out)
	}
}
