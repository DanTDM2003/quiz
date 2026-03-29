package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"quiz/internal/quiz"
)

func TestSubmitAnswerHTTP(t *testing.T) {
	t.Parallel()
	reg := quiz.SeededRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", JoinHandler(reg))
	mux.HandleFunc("POST /v1/quizzes/{quizId}/answers", SubmitAnswerHandler(reg, nil))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Post(
		srv.URL+"/v1/quizzes/sample-quiz/participants",
		"application/json",
		bytes.NewBufferString(`{"displayName":"Ada"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var joined struct {
		ParticipantID string `json:"participantId"`
	}
	if err := json.NewDecoder(res.Body).Decode(&joined); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		srv.URL+"/v1/quizzes/sample-quiz/answers",
		bytes.NewBufferString(`{"questionId":"q1","selectedOptionId":"a"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Participant-Id", joined.ParticipantID)

	ans, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer ans.Body.Close()
	if ans.StatusCode != http.StatusOK {
		t.Fatalf("status %d", ans.StatusCode)
	}
	var body struct {
		ScoreDelta int  `json:"scoreDelta"`
		TotalScore int  `json:"totalScore"`
		Correct    bool `json:"correct"`
	}
	if err := json.NewDecoder(ans.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ScoreDelta != 10 || body.TotalScore != 10 || !body.Correct {
		t.Fatalf("%+v", body)
	}
}
