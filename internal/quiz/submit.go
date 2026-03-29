package quiz

import (
	"quiz/internal/domain"
)

func (r *Registry) SubmitAnswer(
	quizID string,
	participantID string,
	idempotencyKey string,
	questionID string,
	selectedOptionID string,
) (AnswerResult, string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	q, ok := r.quizzes[quizID]
	if !ok {
		return AnswerResult{}, "QUIZ_NOT_FOUND"
	}
	if !q.Active {
		return AnswerResult{}, "QUIZ_NOT_ACTIVE"
	}
	p, ok := q.Participants[participantID]
	if !ok {
		return AnswerResult{}, "PARTICIPANT_NOT_IN_QUIZ"
	}
	qdef, ok := q.Questions[questionID]
	if !ok {
		return AnswerResult{}, "QUESTION_NOT_FOUND"
	}
	if _, ok := qdef.Options[selectedOptionID]; !ok {
		return AnswerResult{}, "INVALID_OPTION"
	}

	if idempotencyKey != "" {
		if prev, ok := p.Idempotency[idempotencyKey]; ok {
			if prev.QuestionID == questionID && prev.SelectedOptionID == selectedOptionID {
				return prev.Result, ""
			}
			return AnswerResult{}, "IDEMPOTENCY_KEY_REUSE_MISMATCH"
		}
	}

	if _, ok := p.Answered[questionID]; ok {
		return AnswerResult{}, "QUESTION_ALREADY_ANSWERED"
	}

	dq := domain.Question{
		ID:              qdef.ID,
		CorrectOptionID: qdef.CorrectOptionID,
		Points:          qdef.Points,
	}
	out := domain.ScoreAnswer(dq, selectedOptionID)
	p.TotalScore += out.ScoreDelta
	p.Answered[questionID] = struct{}{}

	res := AnswerResult{
		QuizID:        quizID,
		ParticipantID: participantID,
		QuestionID:    questionID,
		ScoreDelta:    out.ScoreDelta,
		TotalScore:    p.TotalScore,
		Correct:       out.Correct,
	}

	if idempotencyKey != "" {
		if p.Idempotency == nil {
			p.Idempotency = make(map[string]IdempotencyRecord)
		}
		p.Idempotency[idempotencyKey] = IdempotencyRecord{
			QuestionID:       questionID,
			SelectedOptionID: selectedOptionID,
			Result:           res,
		}
	}

	return res, ""
}
