package game

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

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
