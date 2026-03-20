package game

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"albatross-2026-backend/db"
)

// stubQuerier implements db.Querier with only the methods needed for tests.
// All unimplemented methods panic so missing stubs are caught immediately.
type stubQuerier struct {
	db.Querier
	getGameByID                  func(ctx context.Context, gameID int32) (db.GetGameByIDRow, error)
	getLatestStatesOfMainPlayers func(ctx context.Context, gameID int32) ([]db.GetLatestStatesOfMainPlayersRow, error)
}

func (s *stubQuerier) GetGameByID(ctx context.Context, gameID int32) (db.GetGameByIDRow, error) {
	return s.getGameByID(ctx, gameID)
}

func (s *stubQuerier) GetLatestStatesOfMainPlayers(ctx context.Context, gameID int32) ([]db.GetLatestStatesOfMainPlayersRow, error) {
	return s.getLatestStatesOfMainPlayers(ctx, gameID)
}

func TestIsGameRunning(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		startedAt       pgtype.Timestamp
		durationSeconds int32
		want            bool
	}{
		{
			name:            "not started",
			startedAt:       pgtype.Timestamp{Valid: false},
			durationSeconds: 300,
			want:            false,
		},
		{
			name:            "running",
			startedAt:       pgtype.Timestamp{Time: now.Add(-1 * time.Minute), Valid: true},
			durationSeconds: 300,
			want:            true,
		},
		{
			name:            "finished",
			startedAt:       pgtype.Timestamp{Time: now.Add(-10 * time.Minute), Valid: true},
			durationSeconds: 300,
			want:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGameRunning(tt.startedAt, tt.durationSeconds)
			if got != tt.want {
				t.Errorf("IsGameRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGameFinished(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		startedAt       pgtype.Timestamp
		durationSeconds int32
		want            bool
	}{
		{
			name:            "not started",
			startedAt:       pgtype.Timestamp{Valid: false},
			durationSeconds: 300,
			want:            false,
		},
		{
			name:            "still running",
			startedAt:       pgtype.Timestamp{Time: now.Add(-1 * time.Minute), Valid: true},
			durationSeconds: 300,
			want:            false,
		},
		{
			name:            "finished",
			startedAt:       pgtype.Timestamp{Time: now.Add(-10 * time.Minute), Valid: true},
			durationSeconds: 300,
			want:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGameFinished(tt.startedAt, tt.durationSeconds)
			if got != tt.want {
				t.Errorf("IsGameFinished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWatchLatestStates_ParticipantRestriction(t *testing.T) {
	now := time.Now()
	var playerID int32 = 1
	var otherID int32 = 2

	code := "<?php echo 1;"
	status := "pass"
	var codeSize int32 = 14

	mainPlayerRows := []db.GetLatestStatesOfMainPlayersRow{
		{
			GameID:   1,
			UserID:   playerID,
			Code:     &code,
			Status:   &status,
			CodeSize: &codeSize,
		},
		{
			GameID:   1,
			UserID:   otherID,
			Code:     &code,
			Status:   &status,
			CodeSize: &codeSize,
		},
	}

	tests := []struct {
		name      string
		startedAt pgtype.Timestamp
		userID    *int32
		isAdmin   bool
		wantErr   error
	}{
		{
			name:      "participant blocked while game is running",
			startedAt: pgtype.Timestamp{Time: now.Add(-1 * time.Minute), Valid: true},
			userID:    &playerID,
			isAdmin:   false,
			wantErr:   ErrForbidden,
		},
		{
			name:      "participant allowed after game finished",
			startedAt: pgtype.Timestamp{Time: now.Add(-10 * time.Minute), Valid: true},
			userID:    &playerID,
			isAdmin:   false,
			wantErr:   nil,
		},
		{
			name:      "participant blocked before game starts",
			startedAt: pgtype.Timestamp{Valid: false},
			userID:    &playerID,
			isAdmin:   false,
			wantErr:   ErrForbidden,
		},
		{
			name:      "admin always allowed even while running",
			startedAt: pgtype.Timestamp{Time: now.Add(-1 * time.Minute), Valid: true},
			userID:    &playerID,
			isAdmin:   true,
			wantErr:   nil,
		},
		{
			name:      "non-participant allowed while running",
			startedAt: pgtype.Timestamp{Time: now.Add(-1 * time.Minute), Valid: true},
			userID:    nil,
			isAdmin:   false,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &stubQuerier{
				getGameByID: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
					return db.GetGameByIDRow{
						StartedAt:       tt.startedAt,
						DurationSeconds: 300,
					}, nil
				},
				getLatestStatesOfMainPlayers: func(_ context.Context, _ int32) ([]db.GetLatestStatesOfMainPlayersRow, error) {
					return mainPlayerRows, nil
				},
			}
			svc := NewService(q, nil, nil)

			_, err := svc.GetWatchLatestStates(context.Background(), 1, tt.userID, tt.isAdmin)
			if err != tt.wantErr {
				t.Errorf("got err=%v, want %v", err, tt.wantErr)
			}
		})
	}
}
