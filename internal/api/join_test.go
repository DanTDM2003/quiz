package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"quiz/internal/quiz"
)

func TestJoinSampleQuiz(t *testing.T) {
	t.Parallel()
	reg := quiz.NewRegistry()
	reg.Register(&quiz.Quiz{
		ID:           "sample-quiz",
		Active:       true,
		Participants: make(map[string]*quiz.Participant),
	})
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", JoinHandler(reg))
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
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body struct {
		QuizID        string `json:"quizId"`
		ParticipantID string `json:"participantId"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.QuizID != "sample-quiz" || !looksLikeUUID(body.ParticipantID) {
		t.Fatalf("body %+v", body)
	}
}

func TestJoinNotFound(t *testing.T) {
	t.Parallel()
	reg := quiz.NewRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", JoinHandler(reg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Post(
		srv.URL+"/v1/quizzes/missing/participants",
		"application/json",
		bytes.NewBufferString(`{"displayName":"Ada"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestJoinInactive(t *testing.T) {
	t.Parallel()
	reg := quiz.NewRegistry()
	reg.Register(&quiz.Quiz{
		ID:           "closed",
		Active:       false,
		Participants: make(map[string]*quiz.Participant),
	})
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", JoinHandler(reg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Post(
		srv.URL+"/v1/quizzes/closed/participants",
		"application/json",
		bytes.NewBufferString(`{"displayName":"Ada"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func looksLikeUUID(s string) bool {
	parts := strings.Split(s, "-")
	return len(parts) == 5 && len(parts[0]) == 8 && len(parts[1]) == 4
}
