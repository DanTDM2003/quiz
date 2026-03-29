package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"quiz/internal/quiz"
	"quiz/internal/realtime"
)

func TestStreamJoinWelcome(t *testing.T) {
	t.Parallel()
	reg := quiz.SeededRegistry()
	hub := realtime.NewHub()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", JoinHandler(reg))
	mux.HandleFunc("GET /v1/quizzes/{quizId}/stream", StreamHandler(reg, hub))
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

	wsURL := strings.Replace(srv.URL, "http", "ws", 1) + "/v1/quizzes/sample-quiz/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	payload := `{"type":"join","quizId":"sample-quiz","participantId":"` + joined.ParticipantID + `"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
		t.Fatal(err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var welcome struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &welcome); err != nil {
		t.Fatal(err)
	}
	if welcome.Type != "welcome" {
		t.Fatalf("got %s", welcome.Type)
	}
}

func TestSubmitBroadcastsToSubscriber(t *testing.T) {
	t.Parallel()
	reg := quiz.SeededRegistry()
	hub := realtime.NewHub()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", JoinHandler(reg))
	mux.HandleFunc("POST /v1/quizzes/{quizId}/answers", SubmitAnswerHandler(reg, hub))
	mux.HandleFunc("GET /v1/quizzes/{quizId}/stream", StreamHandler(reg, hub))
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

	wsURL := strings.Replace(srv.URL, "http", "ws", 1) + "/v1/quizzes/sample-quiz/stream"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wsConn.Close() })
	joinPayload := `{"type":"join","quizId":"sample-quiz","participantId":"` + joined.ParticipantID + `"}`
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte(joinPayload)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := wsConn.ReadMessage(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := wsConn.ReadMessage(); err != nil {
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

	var sawScore, sawBoard bool
	for i := 0; i < 4 && !(sawScore && sawBoard); i++ {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		var head struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &head); err != nil {
			t.Fatal(err)
		}
		switch head.Type {
		case "score_updated":
			sawScore = true
		case "leaderboard_updated":
			sawBoard = true
		}
	}
	if !sawScore || !sawBoard {
		t.Fatalf("score=%v board=%v", sawScore, sawBoard)
	}
}
