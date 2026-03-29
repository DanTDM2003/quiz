package domain

import "sort"

func BuildLeaderboard(participants []ParticipantStanding) []LeaderboardEntry {
	sorted := append([]ParticipantStanding(nil), participants...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].TotalScore != sorted[j].TotalScore {
			return sorted[i].TotalScore > sorted[j].TotalScore
		}
		return sorted[i].ParticipantID < sorted[j].ParticipantID
	})
	out := make([]LeaderboardEntry, 0, len(sorted))
	for i := range sorted {
		var rank int
		if i == 0 || sorted[i].TotalScore != sorted[i-1].TotalScore {
			rank = i + 1
		} else {
			rank = out[i-1].Rank
		}
		out = append(out, LeaderboardEntry{
			Rank:          rank,
			ParticipantID: sorted[i].ParticipantID,
			DisplayName:   sorted[i].DisplayName,
			TotalScore:    sorted[i].TotalScore,
		})
	}
	return out
}
