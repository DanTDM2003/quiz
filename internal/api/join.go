package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"quiz/internal/quiz"
)

type joinRequest struct {
	DisplayName string `json:"displayName"`
}

type joinResponse struct {
	QuizID        string `json:"quizId"`
	ParticipantID string `json:"participantId"`
}

func JoinHandler(reg *quiz.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid body")
			return
		}
		var body joinRequest
		if err := json.Unmarshal(raw, &body); err != nil {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid body")
			return
		}
		displayName := strings.TrimSpace(body.DisplayName)
		if displayName == "" || len(displayName) > 64 {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid displayName")
			return
		}

		quizID := r.PathValue("quizId")
		participantID, errCode := reg.Join(quizID, displayName)
		switch errCode {
		case "":
			writeJSON(w, http.StatusCreated, joinResponse{QuizID: quizID, ParticipantID: participantID})
		case "QUIZ_NOT_FOUND":
			writeErr(w, http.StatusNotFound, "QUIZ_NOT_FOUND", "Quiz not found")
		case "QUIZ_NOT_ACTIVE":
			writeErr(w, http.StatusConflict, "QUIZ_NOT_ACTIVE", "Quiz not active")
		case "INTERNAL_ERROR":
			writeErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal error")
		default:
			writeErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal error")
		}
	}
}
