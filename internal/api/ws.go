package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"quiz/internal/quiz"
	"quiz/internal/realtime"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type clientJoinMsg struct {
	Type          string `json:"type"`
	QuizID        string `json:"quizId"`
	ParticipantID string `json:"participantId"`
}

func StreamHandler(reg *quiz.Registry, hub *realtime.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		quizID := r.PathValue("quizId")
		_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, payload, err := conn.ReadMessage()
		if err != nil {
			_ = conn.Close()
			return
		}
		var join clientJoinMsg
		if err := json.Unmarshal(payload, &join); err != nil {
			_ = realtime.WriteError(conn, "INVALID_OPTION", "Invalid body")
			_ = conn.Close()
			return
		}
		if join.Type != "join" {
			_ = realtime.WriteError(conn, "INVALID_OPTION", "Expected join")
			_ = conn.Close()
			return
		}
		if strings.TrimSpace(join.QuizID) != quizID {
			_ = realtime.WriteError(conn, "INVALID_OPTION", "quizId mismatch")
			_ = conn.Close()
			return
		}
		participantID := strings.TrimSpace(join.ParticipantID)
		if participantID == "" {
			_ = realtime.WriteError(conn, "INVALID_OPTION", "participantId is required")
			_ = conn.Close()
			return
		}
		displayName, errCode := reg.LookupParticipant(quizID, participantID)
		if errCode != "" {
			switch errCode {
			case "QUIZ_NOT_FOUND":
				_ = realtime.WriteError(conn, "QUIZ_NOT_FOUND", "Quiz not found")
			case "QUIZ_NOT_ACTIVE":
				_ = realtime.WriteError(conn, "QUIZ_NOT_ACTIVE", "Quiz not active")
			case "PARTICIPANT_NOT_IN_QUIZ":
				_ = realtime.WriteError(conn, "PARTICIPANT_NOT_IN_QUIZ", "Participant not in quiz")
			default:
				_ = realtime.WriteError(conn, "INTERNAL_ERROR", "Internal error")
			}
			_ = conn.Close()
			return
		}
		if hub == nil {
			_ = realtime.WriteError(conn, "INTERNAL_ERROR", "Internal error")
			_ = conn.Close()
			return
		}
		hub.Add(quizID, participantID, conn)
		if err := realtime.WriteWelcome(conn, quizID, participantID); err != nil {
			hub.Remove(quizID, participantID)
			_ = conn.Close()
			return
		}
		hub.BroadcastParticipantJoined(quizID, participantID, displayName)
		_ = conn.SetReadDeadline(time.Time{})
		go func() {
			defer hub.Remove(quizID, participantID)
			defer conn.Close()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
	}
}
