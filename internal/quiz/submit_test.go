package quiz

import "testing"

func TestSubmitAnswerScoring(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register(&Quiz{
		ID:     "q",
		Active: true,
		Questions: map[string]*Question{
			"x": {
				ID:              "x",
				CorrectOptionID: "a",
				Points:          5,
				Options:         map[string]struct{}{"a": {}, "b": {}},
			},
		},
		Participants: map[string]*Participant{
			"p1": {
				DisplayName: "U",
				Answered:    make(map[string]struct{}),
				Idempotency: make(map[string]IdempotencyRecord),
			},
		},
	})
	res, code := r.SubmitAnswer("q", "p1", "", "x", "a")
	if code != "" {
		t.Fatalf("code %s", code)
	}
	if res.ScoreDelta != 5 || res.TotalScore != 5 || !res.Correct {
		t.Fatalf("%+v", res)
	}
	res, code = r.SubmitAnswer("q", "p1", "", "x", "a")
	if code != "QUESTION_ALREADY_ANSWERED" {
		t.Fatalf("got %s", code)
	}
}

func TestSubmitIdempotencyReplay(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register(&Quiz{
		ID:     "q",
		Active: true,
		Questions: map[string]*Question{
			"x": {
				ID:              "x",
				CorrectOptionID: "a",
				Points:          5,
				Options:         map[string]struct{}{"a": {}, "b": {}},
			},
		},
		Participants: map[string]*Participant{
			"p1": {
				DisplayName: "U",
				Answered:    make(map[string]struct{}),
				Idempotency: make(map[string]IdempotencyRecord),
			},
		},
	})
	key := "idem-key-1"
	res1, code := r.SubmitAnswer("q", "p1", key, "x", "a")
	if code != "" {
		t.Fatalf("code %s", code)
	}
	res2, code := r.SubmitAnswer("q", "p1", key, "x", "a")
	if code != "" {
		t.Fatalf("code %s", code)
	}
	if res1 != res2 {
		t.Fatalf("mismatch %+v %+v", res1, res2)
	}
}

func TestSubmitIdempotencyMismatch(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register(&Quiz{
		ID:     "q",
		Active: true,
		Questions: map[string]*Question{
			"x": {
				ID:              "x",
				CorrectOptionID: "a",
				Points:          5,
				Options:         map[string]struct{}{"a": {}, "b": {}},
			},
			"y": {
				ID:              "y",
				CorrectOptionID: "a",
				Points:          1,
				Options:         map[string]struct{}{"a": {}, "b": {}},
			},
		},
		Participants: map[string]*Participant{
			"p1": {
				DisplayName: "U",
				Answered:    make(map[string]struct{}),
				Idempotency: make(map[string]IdempotencyRecord),
			},
		},
	})
	key := "idem-key-2"
	_, code := r.SubmitAnswer("q", "p1", key, "x", "a")
	if code != "" {
		t.Fatalf("code %s", code)
	}
	_, code = r.SubmitAnswer("q", "p1", key, "y", "a")
	if code != "IDEMPOTENCY_KEY_REUSE_MISMATCH" {
		t.Fatalf("got %s", code)
	}
}
