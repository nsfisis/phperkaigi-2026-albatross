package api

import (
	"github.com/oapi-codegen/nullable"

	"albatross-2026-backend/game"
	"albatross-2026-backend/tournament"
)

func toAPIUser(p game.Player) User {
	return User{
		UserID:      p.UserID,
		Username:    p.Username,
		DisplayName: p.DisplayName,
		IconPath:    p.IconPath,
		IsAdmin:     p.IsAdmin,
		Label:       toNullable(p.Label),
	}
}

func toAPIGame(g game.GameDetail) Game {
	var startedAt *int64
	if g.StartedAt != nil {
		ts := g.StartedAt.Unix()
		startedAt = &ts
	}
	mainPlayers := make([]User, len(g.MainPlayers))
	for i, p := range g.MainPlayers {
		mainPlayers[i] = toAPIUser(p)
	}
	return Game{
		GameID:          g.GameID,
		GameType:        GameType(g.GameType),
		IsPublic:        g.IsPublic,
		DisplayName:     g.DisplayName,
		DurationSeconds: g.DurationSeconds,
		StartedAt:       startedAt,
		Problem: Problem{
			ProblemID:   g.Problem.ProblemID,
			Title:       g.Problem.Title,
			Description: g.Problem.Description,
			Language:    ProblemLanguage(g.Problem.Language),
			SampleCode:  g.Problem.SampleCode,
		},
		MainPlayers: mainPlayers,
	}
}

func toAPILatestState(s game.LatestState) LatestGameState {
	var score nullable.Nullable[int]
	if s.Score != nil {
		score = nullable.NewNullableWithValue(*s.Score)
	} else {
		score = nullable.NewNullNullable[int]()
	}
	var submittedAt nullable.Nullable[int64]
	if s.BestScoreSubmittedAt != nil {
		submittedAt = nullable.NewNullableWithValue(*s.BestScoreSubmittedAt)
	} else {
		submittedAt = nullable.NewNullNullable[int64]()
	}
	return LatestGameState{
		Code:                 s.Code,
		Score:                score,
		BestScoreSubmittedAt: submittedAt,
		Status:               ExecutionStatus(s.Status),
	}
}

func toAPIRankingEntry(r game.RankingEntry) RankingEntry {
	var code nullable.Nullable[string]
	if r.Code != nil {
		code = nullable.NewNullableWithValue(*r.Code)
	} else {
		code = nullable.NewNullNullable[string]()
	}
	return RankingEntry{
		Player: toAPIUser(r.Player),
		Score:  r.Score,
		SubmittedAt: r.SubmittedAt,
		Code:   code,
	}
}

func toAPISubmission(s game.SubmissionDetail) Submission {
	return Submission{
		SubmissionID: s.SubmissionID,
		GameID:       s.GameID,
		Code:         s.Code,
		CodeSize:     s.CodeSize,
		Status:       ExecutionStatus(s.Status),
		CreatedAt:    s.CreatedAt,
	}
}

func toAPITournamentUser(p tournament.Player) User {
	return User{
		UserID:      p.UserID,
		Username:    p.Username,
		DisplayName: p.DisplayName,
		IconPath:    p.IconPath,
		IsAdmin:     p.IsAdmin,
		Label:       toNullable(p.Label),
	}
}

func toAPITournamentPlayerPtr(p *tournament.Player) *User {
	if p == nil {
		return nil
	}
	u := toAPITournamentUser(*p)
	return &u
}

func toAPITournament(t tournament.TournamentBracket) Tournament {
	entries := make([]TournamentEntry, len(t.Entries))
	for i, e := range t.Entries {
		entries[i] = TournamentEntry{
			User: toAPITournamentUser(e.User),
			Seed: e.Seed,
		}
	}
	matches := make([]TournamentMatch, len(t.Matches))
	for i, m := range t.Matches {
		matches[i] = TournamentMatch{
			TournamentMatchID: m.TournamentMatchID,
			Round:             m.Round,
			Position:          m.Position,
			GameID:            m.GameID,
			Player1:           toAPITournamentPlayerPtr(m.Player1),
			Player2:           toAPITournamentPlayerPtr(m.Player2),
			Player1Score:      m.Player1Score,
			Player2Score:      m.Player2Score,
			WinnerUserID:      m.WinnerUserID,
			IsBye:             m.IsBye,
		}
	}
	return Tournament{
		TournamentID: t.TournamentID,
		DisplayName:  t.DisplayName,
		BracketSize:  t.BracketSize,
		NumRounds:    t.NumRounds,
		Entries:      entries,
		Matches:      matches,
	}
}

func toNullable[T any](p *T) nullable.Nullable[T] {
	if p == nil {
		return nullable.NewNullNullable[T]()
	}
	return nullable.NewNullableWithValue(*p)
}

func toNullableWith[T, U any](p *T, f func(T) U) nullable.Nullable[U] {
	if p == nil {
		return nullable.NewNullNullable[U]()
	}
	return nullable.NewNullableWithValue(f(*p))
}
