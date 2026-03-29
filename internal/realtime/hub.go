package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"quiz/internal/domain"
	"quiz/internal/quiz"
)

type Hub struct {
	mu    sync.Mutex
	rooms map[string]map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[string]*websocket.Conn)}
}

func (h *Hub) Add(quizID, participantID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[quizID] == nil {
		h.rooms[quizID] = make(map[string]*websocket.Conn)
	}
	if old := h.rooms[quizID][participantID]; old != nil {
		_ = old.Close()
	}
	h.rooms[quizID][participantID] = c
}

func (h *Hub) Remove(quizID, participantID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[quizID]
	if room == nil {
		return
	}
	delete(room, participantID)
	if len(room) == 0 {
		delete(h.rooms, quizID)
	}
}

func (h *Hub) broadcast(quizID string, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.mu.Lock()
	room := h.rooms[quizID]
	type conn struct {
		id string
		c  *websocket.Conn
	}
	var list []conn
	for id, c := range room {
		list = append(list, conn{id: id, c: c})
	}
	h.mu.Unlock()
	deadline := time.Now().Add(10 * time.Second)
	for _, x := range list {
		_ = x.c.SetWriteDeadline(deadline)
		if err := x.c.WriteMessage(websocket.TextMessage, data); err != nil {
			h.Remove(quizID, x.id)
			_ = x.c.Close()
		}
	}
}

func (h *Hub) BroadcastScoreUpdated(quizID string, res quiz.AnswerResult, ts time.Time) {
	if h == nil {
		return
	}
	h.broadcast(quizID, scoreUpdatedMsg{
		Type:          "score_updated",
		QuizID:        quizID,
		ParticipantID: res.ParticipantID,
		QuestionID:    res.QuestionID,
		ScoreDelta:    res.ScoreDelta,
		TotalScore:    res.TotalScore,
		Correct:       res.Correct,
		UpdatedAt:     ts.Format(time.RFC3339Nano),
	})
}

func (h *Hub) BroadcastLeaderboardUpdated(quizID string, reg *quiz.Registry, ts time.Time) {
	if h == nil {
		return
	}
	standings, code := reg.ParticipantStandings(quizID)
	if code != "" {
		return
	}
	rows := domain.BuildLeaderboard(standings)
	entries := make([]leaderboardEntryMsg, len(rows))
	for i := range rows {
		entries[i] = leaderboardEntryMsg{
			Rank:          rows[i].Rank,
			ParticipantID: rows[i].ParticipantID,
			DisplayName:   rows[i].DisplayName,
			TotalScore:    rows[i].TotalScore,
		}
	}
	h.broadcast(quizID, leaderboardUpdatedMsg{
		Type:      "leaderboard_updated",
		QuizID:    quizID,
		Entries:   entries,
		UpdatedAt: ts.Format(time.RFC3339Nano),
	})
}

func (h *Hub) BroadcastParticipantJoined(quizID, participantID, displayName string) {
	if h == nil {
		return
	}
	h.broadcast(quizID, participantJoinedMsg{
		Type:          "participant_joined",
		QuizID:        quizID,
		ParticipantID: participantID,
		DisplayName:   displayName,
	})
}

type scoreUpdatedMsg struct {
	Type          string `json:"type"`
	QuizID        string `json:"quizId"`
	ParticipantID string `json:"participantId"`
	QuestionID    string `json:"questionId"`
	ScoreDelta    int    `json:"scoreDelta"`
	TotalScore    int    `json:"totalScore"`
	Correct       bool   `json:"correct"`
	UpdatedAt     string `json:"updatedAt"`
}

type leaderboardEntryMsg struct {
	Rank          int    `json:"rank"`
	ParticipantID string `json:"participantId"`
	DisplayName   string `json:"displayName"`
	TotalScore    int    `json:"totalScore"`
}

type leaderboardUpdatedMsg struct {
	Type      string                `json:"type"`
	QuizID    string                `json:"quizId"`
	Entries   []leaderboardEntryMsg `json:"entries"`
	UpdatedAt string                `json:"updatedAt"`
}

type participantJoinedMsg struct {
	Type          string `json:"type"`
	QuizID        string `json:"quizId"`
	ParticipantID string `json:"participantId"`
	DisplayName   string `json:"displayName"`
}

type welcomeMsg struct {
	Type          string `json:"type"`
	QuizID        string `json:"quizId"`
	ParticipantID string `json:"participantId"`
}

type wsErrMsg struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
