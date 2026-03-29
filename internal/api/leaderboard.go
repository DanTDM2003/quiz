package api

import (
	"net/http"
	"time"

	"quiz/internal/domain"
	"quiz/internal/quiz"
)

type leaderboardHTTPResponse struct {
	QuizID    string               `json:"quizId"`
	Entries   []leaderboardHTTPRow `json:"entries"`
	UpdatedAt string               `json:"updatedAt"`
}

type leaderboardHTTPRow struct {
	Rank          int    `json:"rank"`
	ParticipantID string `json:"participantId"`
	DisplayName   string `json:"displayName"`
	TotalScore    int    `json:"totalScore"`
}

func LeaderboardHandler(reg *quiz.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quizID := r.PathValue("quizId")
		standings, errCode := reg.ParticipantStandings(quizID)
		switch errCode {
		case "":
		case "QUIZ_NOT_FOUND":
			writeErr(w, http.StatusNotFound, "QUIZ_NOT_FOUND", "Quiz not found")
			return
		default:
			writeErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal error")
			return
		}
		rows := domain.BuildLeaderboard(standings)
		entries := make([]leaderboardHTTPRow, len(rows))
		for i := range rows {
			entries[i] = leaderboardHTTPRow{
				Rank:          rows[i].Rank,
				ParticipantID: rows[i].ParticipantID,
				DisplayName:   rows[i].DisplayName,
				TotalScore:    rows[i].TotalScore,
			}
		}
		writeJSON(w, http.StatusOK, leaderboardHTTPResponse{
			QuizID:    quizID,
			Entries:   entries,
			UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		})
	}
}
