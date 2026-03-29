package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"quiz/internal/quiz"
	"quiz/internal/realtime"
)

type submitAnswerRequest struct {
	QuestionID       string `json:"questionId"`
	SelectedOptionID string `json:"selectedOptionId"`
}

type submitAnswerResponse struct {
	QuizID        string `json:"quizId"`
	ParticipantID string `json:"participantId"`
	QuestionID    string `json:"questionId"`
	ScoreDelta    int    `json:"scoreDelta"`
	TotalScore    int    `json:"totalScore"`
	Correct       bool   `json:"correct"`
}

func SubmitAnswerHandler(reg *quiz.Registry, hub *realtime.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid body")
			return
		}
		var body submitAnswerRequest
		if err := json.Unmarshal(raw, &body); err != nil {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid body")
			return
		}
		questionID := strings.TrimSpace(body.QuestionID)
		selectedOptionID := strings.TrimSpace(body.SelectedOptionID)
		if questionID == "" || selectedOptionID == "" {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid body")
			return
		}

		participantID := strings.TrimSpace(r.Header.Get("X-Participant-Id"))
		if participantID == "" {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "X-Participant-Id is required")
			return
		}

		idem := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if idem != "" && (len(idem) < 8 || len(idem) > 128) {
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid Idempotency-Key")
			return
		}

		quizID := r.PathValue("quizId")
		result, errCode := reg.SubmitAnswer(quizID, participantID, idem, questionID, selectedOptionID)
		switch errCode {
		case "":
			writeJSON(w, http.StatusOK, submitAnswerResponse{
				QuizID:        result.QuizID,
				ParticipantID: result.ParticipantID,
				QuestionID:    result.QuestionID,
				ScoreDelta:    result.ScoreDelta,
				TotalScore:    result.TotalScore,
				Correct:       result.Correct,
			})
			if hub != nil {
				ts := time.Now().UTC()
				hub.BroadcastScoreUpdated(quizID, result, ts)
				hub.BroadcastLeaderboardUpdated(quizID, reg, ts)
			}
		case "QUIZ_NOT_FOUND":
			writeErr(w, http.StatusNotFound, "QUIZ_NOT_FOUND", "Quiz not found")
		case "PARTICIPANT_NOT_IN_QUIZ":
			writeErr(w, http.StatusNotFound, "PARTICIPANT_NOT_IN_QUIZ", "Participant not in quiz")
		case "QUESTION_NOT_FOUND":
			writeErr(w, http.StatusNotFound, "QUESTION_NOT_FOUND", "Question not found")
		case "INVALID_OPTION":
			writeErr(w, http.StatusBadRequest, "INVALID_OPTION", "Invalid option")
		case "QUIZ_NOT_ACTIVE":
			writeErr(w, http.StatusConflict, "QUIZ_NOT_ACTIVE", "Quiz not active")
		case "QUESTION_ALREADY_ANSWERED":
			writeErr(w, http.StatusConflict, "QUESTION_ALREADY_ANSWERED", "Question already answered")
		case "IDEMPOTENCY_KEY_REUSE_MISMATCH":
			writeErr(w, http.StatusConflict, "IDEMPOTENCY_KEY_REUSE_MISMATCH", "Idempotency key reuse mismatch")
		case "INTERNAL_ERROR":
			writeErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal error")
		default:
			writeErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal error")
		}
	}
}
