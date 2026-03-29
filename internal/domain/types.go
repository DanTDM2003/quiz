package domain

type Question struct {
	ID              string
	CorrectOptionID string
	Points          int
}

type ParticipantStanding struct {
	ParticipantID string
	DisplayName   string
	TotalScore    int
}

type LeaderboardEntry struct {
	Rank          int
	ParticipantID string
	DisplayName   string
	TotalScore    int
}
