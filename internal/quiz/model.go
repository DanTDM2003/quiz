package quiz

type AnswerResult struct {
	QuizID        string
	ParticipantID string
	QuestionID    string
	ScoreDelta    int
	TotalScore    int
	Correct       bool
}

type IdempotencyRecord struct {
	QuestionID       string
	SelectedOptionID string
	Result           AnswerResult
}

type Participant struct {
	DisplayName string
	TotalScore  int
	Answered    map[string]struct{}
	Idempotency map[string]IdempotencyRecord
}

type Question struct {
	ID              string
	CorrectOptionID string
	Points          int
	Options         map[string]struct{}
}

type Quiz struct {
	ID           string
	Active       bool
	Questions    map[string]*Question
	Participants map[string]*Participant
}
