package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"quiz/internal/quiz"
)

func TestLeaderboardNotFound(t *testing.T) {
	t.Parallel()
	reg := quiz.NewRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/quizzes/{quizId}/leaderboard", LeaderboardHandler(reg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/v1/quizzes/unknown/leaderboard")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestLeaderboardOrdering(t *testing.T) {
	t.Parallel()
	reg := quiz.SeededRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", JoinHandler(reg))
	mux.HandleFunc("POST /v1/quizzes/{quizId}/answers", SubmitAnswerHandler(reg, nil))
	mux.HandleFunc("GET /v1/quizzes/{quizId}/leaderboard", LeaderboardHandler(reg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resA, err := http.Post(
		srv.URL+"/v1/quizzes/sample-quiz/participants",
		"application/json",
		bytes.NewBufferString(`{"displayName":"A"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resA.Body.Close()
	var bodyA struct {
		ParticipantID string `json:"participantId"`
	}
	if err := json.NewDecoder(resA.Body).Decode(&bodyA); err != nil {
		t.Fatal(err)
	}

	resB, err := http.Post(
		srv.URL+"/v1/quizzes/sample-quiz/participants",
		"application/json",
		bytes.NewBufferString(`{"displayName":"B"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resB.Body.Close()
	var bodyB struct {
		ParticipantID string `json:"participantId"`
	}
	if err := json.NewDecoder(resB.Body).Decode(&bodyB); err != nil {
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
	req.Header.Set("X-Participant-Id", bodyA.ParticipantID)
	ans, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	ans.Body.Close()

	res, err := http.Get(srv.URL + "/v1/quizzes/sample-quiz/leaderboard")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var out struct {
		QuizID  string `json:"quizId"`
		Entries []struct {
			Rank          int    `json:"rank"`
			ParticipantID string `json:"participantId"`
			DisplayName   string `json:"displayName"`
			TotalScore    int    `json:"totalScore"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.QuizID != "sample-quiz" || len(out.Entries) != 2 {
		t.Fatalf("%+v", out)
	}
	if out.Entries[0].TotalScore != 10 || out.Entries[0].Rank != 1 {
		t.Fatalf("%+v", out.Entries[0])
	}
	if out.Entries[1].TotalScore != 0 || out.Entries[1].Rank != 2 {
		t.Fatalf("%+v", out.Entries[1])
	}
}
