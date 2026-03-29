package quiz

import (
	"sync"

	"quiz/internal/id"
)

type Participant struct {
	DisplayName string
}

type Quiz struct {
	ID           string
	Active       bool
	Participants map[string]Participant
}

type Registry struct {
	mu      sync.RWMutex
	quizzes map[string]*Quiz
}

func NewRegistry() *Registry {
	return &Registry{quizzes: make(map[string]*Quiz)}
}

func (r *Registry) Register(q *Quiz) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quizzes[q.ID] = q
}

func (r *Registry) Join(quizID, displayName string) (participantID string, errCode string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	q, ok := r.quizzes[quizID]
	if !ok {
		return "", "QUIZ_NOT_FOUND"
	}
	if !q.Active {
		return "", "QUIZ_NOT_ACTIVE"
	}
	pid, err := id.RandomUUID()
	if err != nil {
		return "", "INTERNAL_ERROR"
	}
	if q.Participants == nil {
		q.Participants = make(map[string]Participant)
	}
	q.Participants[pid] = Participant{DisplayName: displayName}
	return pid, ""
}

func SeededRegistry() *Registry {
	r := NewRegistry()
	r.Register(&Quiz{
		ID:           "sample-quiz",
		Active:       true,
		Participants: make(map[string]Participant),
	})
	return r
}
