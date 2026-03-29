package domain

import (
	"reflect"
	"testing"
)

func TestBuildLeaderboardOrderAndRanks(t *testing.T) {
	t.Parallel()
	rows := BuildLeaderboard([]ParticipantStanding{
		{ParticipantID: "b", DisplayName: "B", TotalScore: 10},
		{ParticipantID: "a", DisplayName: "A", TotalScore: 10},
		{ParticipantID: "c", DisplayName: "C", TotalScore: 5},
	})
	ids := make([]string, len(rows))
	for i := range rows {
		ids[i] = rows[i].ParticipantID
	}
	if !reflect.DeepEqual(ids, []string{"a", "b", "c"}) {
		t.Fatalf("order: %v", ids)
	}
	if rows[0].Rank != 1 || rows[1].Rank != 1 || rows[2].Rank != 3 {
		t.Fatalf("ranks: %+v", rows)
	}
}

func TestBuildLeaderboardCompetitionRanks(t *testing.T) {
	t.Parallel()
	rows := BuildLeaderboard([]ParticipantStanding{
		{ParticipantID: "p1", DisplayName: "P1", TotalScore: 100},
		{ParticipantID: "p2", DisplayName: "P2", TotalScore: 80},
		{ParticipantID: "p3", DisplayName: "P3", TotalScore: 80},
		{ParticipantID: "p4", DisplayName: "P4", TotalScore: 50},
	})
	ranks := make([]int, len(rows))
	for i := range rows {
		ranks[i] = rows[i].Rank
	}
	if !reflect.DeepEqual(ranks, []int{1, 2, 2, 4}) {
		t.Fatalf("ranks: %v", ranks)
	}
}
