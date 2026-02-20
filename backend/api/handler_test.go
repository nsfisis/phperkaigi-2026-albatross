package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"albatross-2026-backend/config"
	"albatross-2026-backend/db"
)

// mockQuerier implements db.Querier for testing.
type mockQuerier struct {
	db.Querier
	getGameByIDFunc                     func(ctx context.Context, gameID int32) (db.GetGameByIDRow, error)
	listMainPlayersFunc                 func(ctx context.Context, gameIDs []int32) ([]db.ListMainPlayersRow, error)
	listPublicGamesFunc                 func(ctx context.Context) ([]db.ListPublicGamesRow, error)
	deleteSessionFunc                   func(ctx context.Context, sessionID string) error
	getLatestStateFunc                  func(ctx context.Context, arg db.GetLatestStateParams) (db.GetLatestStateRow, error)
	updateCodeFunc                      func(ctx context.Context, arg db.UpdateCodeParams) error
	getRankingFunc                      func(ctx context.Context, gameID int32) ([]db.GetRankingRow, error)
	getLatestStatesFunc                 func(ctx context.Context, gameID int32) ([]db.GetLatestStatesOfMainPlayersRow, error)
	getTournamentByIDFunc               func(ctx context.Context, tournamentID int32) (db.Tournament, error)
	listTournamentEntriesFunc           func(ctx context.Context, tournamentID int32) ([]db.ListTournamentEntriesRow, error)
	listTournamentMatchesFunc           func(ctx context.Context, tournamentID int32) ([]db.TournamentMatch, error)
	getSubmissionsByGameIDAndUserIDFunc func(ctx context.Context, arg db.GetSubmissionsByGameIDAndUserIDParams) ([]db.Submission, error)
}

func (m *mockQuerier) GetGameByID(ctx context.Context, gameID int32) (db.GetGameByIDRow, error) {
	if m.getGameByIDFunc != nil {
		return m.getGameByIDFunc(ctx, gameID)
	}
	return db.GetGameByIDRow{}, pgx.ErrNoRows
}

func (m *mockQuerier) ListMainPlayers(ctx context.Context, gameIDs []int32) ([]db.ListMainPlayersRow, error) {
	if m.listMainPlayersFunc != nil {
		return m.listMainPlayersFunc(ctx, gameIDs)
	}
	return nil, nil
}

func (m *mockQuerier) ListPublicGames(ctx context.Context) ([]db.ListPublicGamesRow, error) {
	if m.listPublicGamesFunc != nil {
		return m.listPublicGamesFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) DeleteSession(ctx context.Context, sessionID string) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *mockQuerier) GetLatestState(ctx context.Context, arg db.GetLatestStateParams) (db.GetLatestStateRow, error) {
	if m.getLatestStateFunc != nil {
		return m.getLatestStateFunc(ctx, arg)
	}
	return db.GetLatestStateRow{}, pgx.ErrNoRows
}

func (m *mockQuerier) UpdateCode(ctx context.Context, arg db.UpdateCodeParams) error {
	if m.updateCodeFunc != nil {
		return m.updateCodeFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) GetRanking(ctx context.Context, gameID int32) ([]db.GetRankingRow, error) {
	if m.getRankingFunc != nil {
		return m.getRankingFunc(ctx, gameID)
	}
	return nil, nil
}

func (m *mockQuerier) GetLatestStatesOfMainPlayers(ctx context.Context, gameID int32) ([]db.GetLatestStatesOfMainPlayersRow, error) {
	if m.getLatestStatesFunc != nil {
		return m.getLatestStatesFunc(ctx, gameID)
	}
	return nil, nil
}

func (m *mockQuerier) GetSubmissionsByGameIDAndUserID(ctx context.Context, arg db.GetSubmissionsByGameIDAndUserIDParams) ([]db.Submission, error) {
	if m.getSubmissionsByGameIDAndUserIDFunc != nil {
		return m.getSubmissionsByGameIDAndUserIDFunc(ctx, arg)
	}
	return nil, nil
}

func (m *mockQuerier) GetTournamentByID(ctx context.Context, tournamentID int32) (db.Tournament, error) {
	if m.getTournamentByIDFunc != nil {
		return m.getTournamentByIDFunc(ctx, tournamentID)
	}
	return db.Tournament{}, pgx.ErrNoRows
}

func (m *mockQuerier) ListTournamentEntries(ctx context.Context, tournamentID int32) ([]db.ListTournamentEntriesRow, error) {
	if m.listTournamentEntriesFunc != nil {
		return m.listTournamentEntriesFunc(ctx, tournamentID)
	}
	return nil, nil
}

func (m *mockQuerier) ListTournamentMatches(ctx context.Context, tournamentID int32) ([]db.TournamentMatch, error) {
	if m.listTournamentMatchesFunc != nil {
		return m.listTournamentMatchesFunc(ctx, tournamentID)
	}
	return nil, nil
}

// mockTxManager implements db.TxManager for testing.
type mockTxManager struct{}

func (m *mockTxManager) RunInTx(_ context.Context, fn func(q db.Querier) error) error {
	return fn(&mockQuerier{})
}

// mockGameHub implements GameHubInterface for testing.
type mockGameHub struct {
	calcCodeSizeResult int
	enqueueErr         error
}

func (m *mockGameHub) CalcCodeSize(_ string, _ string) int {
	return m.calcCodeSizeResult
}

func (m *mockGameHub) EnqueueTestTasks(_ context.Context, _, _, _ int, _, _ string) error {
	return m.enqueueErr
}

// mockAuthenticator implements AuthenticatorInterface for testing.
type mockAuthenticator struct {
	loginResult int
	loginErr    error
}

func (m *mockAuthenticator) Login(_ context.Context, _, _ string) (int, error) {
	return m.loginResult, m.loginErr
}

func TestGetGamePlaySubmissions_GameNotFound(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGamePlaySubmissions(context.Background(), GetGamePlaySubmissionsRequestObject{
		GameID: 999,
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(GetGamePlaySubmissions404JSONResponse); !ok {
		t.Errorf("expected 404 response, got %T", resp)
	}
}

func TestGetGamePlaySubmissions_Empty(t *testing.T) {
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:   1,
					Language: "php",
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGamePlaySubmissions(context.Background(), GetGamePlaySubmissionsRequestObject{
		GameID: 1,
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGamePlaySubmissions200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if len(okResp.Submissions) != 0 {
		t.Errorf("expected 0 submissions, got %d", len(okResp.Submissions))
	}
}

func TestGetGamePlaySubmissions_WithSubmissions(t *testing.T) {
	now := time.Now()
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:   1,
					Language: "php",
				}, nil
			},
			getSubmissionsByGameIDAndUserIDFunc: func(_ context.Context, arg db.GetSubmissionsByGameIDAndUserIDParams) ([]db.Submission, error) {
				if arg.GameID != 1 || arg.UserID != 42 {
					t.Errorf("unexpected query params: game_id=%d, user_id=%d", arg.GameID, arg.UserID)
				}
				return []db.Submission{
					{
						SubmissionID: 10,
						GameID:       1,
						UserID:       42,
						Code:         "<?php echo 1;",
						CodeSize:     14,
						Status:       "success",
						CreatedAt:    pgtype.Timestamp{Time: now, Valid: true},
					},
					{
						SubmissionID: 9,
						GameID:       1,
						UserID:       42,
						Code:         "<?php echo 'hello';",
						CodeSize:     20,
						Status:       "wrong_answer",
						CreatedAt:    pgtype.Timestamp{Time: now.Add(-5 * time.Minute), Valid: true},
					},
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 42}
	resp, err := h.GetGamePlaySubmissions(context.Background(), GetGamePlaySubmissionsRequestObject{
		GameID: 1,
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGamePlaySubmissions200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if len(okResp.Submissions) != 2 {
		t.Fatalf("expected 2 submissions, got %d", len(okResp.Submissions))
	}

	s0 := okResp.Submissions[0]
	if s0.SubmissionID != 10 {
		t.Errorf("expected submission_id 10, got %d", s0.SubmissionID)
	}
	if s0.GameID != 1 {
		t.Errorf("expected game_id 1, got %d", s0.GameID)
	}
	if s0.Code != "<?php echo 1;" {
		t.Errorf("expected code '<?php echo 1;', got %q", s0.Code)
	}
	if s0.CodeSize != 14 {
		t.Errorf("expected code_size 14, got %d", s0.CodeSize)
	}
	if s0.Status != Success {
		t.Errorf("expected status 'success', got %q", s0.Status)
	}
	if s0.CreatedAt != now.Unix() {
		t.Errorf("expected created_at %d, got %d", now.Unix(), s0.CreatedAt)
	}

	s1 := okResp.Submissions[1]
	if s1.SubmissionID != 9 {
		t.Errorf("expected submission_id 9, got %d", s1.SubmissionID)
	}
	if s1.Status != WrongAnswer {
		t.Errorf("expected status 'wrong_answer', got %q", s1.Status)
	}
}

func TestPostGamePlaySubmit_GameNotFound(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.PostGamePlaySubmit(context.Background(), PostGamePlaySubmitRequestObject{
		GameID: 999,
		Body:   &PostGamePlaySubmitJSONRequestBody{Code: "test"},
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(PostGamePlaySubmit404JSONResponse); !ok {
		t.Errorf("expected 404 response, got %T", resp)
	}
}

func TestPostGamePlaySubmit_GameNotRunning(t *testing.T) {
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:   1,
					Language: "php",
					StartedAt: pgtype.Timestamp{
						Valid: false,
					},
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{calcCodeSizeResult: 10},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.PostGamePlaySubmit(context.Background(), PostGamePlaySubmitRequestObject{
		GameID: 1,
		Body:   &PostGamePlaySubmitJSONRequestBody{Code: "<?php echo 1;"},
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r, ok := resp.(PostGamePlaySubmit403JSONResponse); !ok {
		t.Errorf("expected 403 response, got %T", resp)
	} else if r.Message != "Game is not running" {
		t.Errorf("unexpected message: %s", r.Message)
	}
}

func TestIsGameRunning(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		game db.GetGameByIDRow
		want bool
	}{
		{
			name: "not started",
			game: db.GetGameByIDRow{
				StartedAt:       pgtype.Timestamp{Valid: false},
				DurationSeconds: 300,
			},
			want: false,
		},
		{
			name: "running",
			game: db.GetGameByIDRow{
				StartedAt:       pgtype.Timestamp{Time: now.Add(-1 * time.Minute), Valid: true},
				DurationSeconds: 300,
			},
			want: true,
		},
		{
			name: "finished",
			game: db.GetGameByIDRow{
				StartedAt:       pgtype.Timestamp{Time: now.Add(-10 * time.Minute), Valid: true},
				DurationSeconds: 300,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGameRunning(tt.game)
			if got != tt.want {
				t.Errorf("isGameRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGameFinished(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		game db.GetGameByIDRow
		want bool
	}{
		{
			name: "not started",
			game: db.GetGameByIDRow{
				StartedAt:       pgtype.Timestamp{Valid: false},
				DurationSeconds: 300,
			},
			want: false,
		},
		{
			name: "still running",
			game: db.GetGameByIDRow{
				StartedAt:       pgtype.Timestamp{Time: now.Add(-1 * time.Minute), Valid: true},
				DurationSeconds: 300,
			},
			want: false,
		},
		{
			name: "finished",
			game: db.GetGameByIDRow{
				StartedAt:       pgtype.Timestamp{Time: now.Add(-10 * time.Minute), Valid: true},
				DurationSeconds: 300,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGameFinished(tt.game)
			if got != tt.want {
				t.Errorf("isGameFinished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToNullable(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result := toNullable[string](nil)
		if !result.IsNull() {
			t.Error("expected null for nil input")
		}
	})
	t.Run("non-nil value", func(t *testing.T) {
		s := "hello"
		result := toNullable(&s)
		if result.IsNull() {
			t.Error("expected non-null for non-nil input")
		}
		v, err := result.Get()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "hello" {
			t.Errorf("expected 'hello', got %q", v)
		}
	})
}

func TestGetMe(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{
		UserID:      1,
		Username:    "testuser",
		DisplayName: "Test User",
		IsAdmin:     false,
	}
	resp, err := h.GetMe(context.Background(), GetMeRequestObject{}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetMe200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.User.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", okResp.User.UserID)
	}
	if okResp.User.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", okResp.User.Username)
	}
	if okResp.User.IsAdmin {
		t.Error("expected non-admin user")
	}
}

func TestGetGame_NotFound(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGame(context.Background(), GetGameRequestObject{GameID: 999}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(GetGame404JSONResponse); !ok {
		t.Errorf("expected 404 response, got %T", resp)
	}
}

func TestGetGame_NonPublicAsNonAdmin(t *testing.T) {
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:   1,
					IsPublic: false,
					Language: "php",
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1, IsAdmin: false}
	resp, err := h.GetGame(context.Background(), GetGameRequestObject{GameID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(GetGame404JSONResponse); !ok {
		t.Errorf("expected 404 for non-public game as non-admin, got %T", resp)
	}
}

func TestGetGame_PublicGameSuccess(t *testing.T) {
	now := time.Now()
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:          1,
					IsPublic:        true,
					Language:        "php",
					DisplayName:     "Test Game",
					DurationSeconds: 300,
					StartedAt:       pgtype.Timestamp{Time: now, Valid: true},
					GameType:        "golf",
					ProblemID:       10,
					Title:           "Test Problem",
					Description:     "desc",
					SampleCode:      "<?php",
				}, nil
			},
			listMainPlayersFunc: func(_ context.Context, _ []int32) ([]db.ListMainPlayersRow, error) {
				return []db.ListMainPlayersRow{
					{UserID: 1, Username: "player1", DisplayName: "Player 1", IsAdmin: false},
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1, IsAdmin: false}
	resp, err := h.GetGame(context.Background(), GetGameRequestObject{GameID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGame200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Game.GameID != 1 {
		t.Errorf("expected game ID 1, got %d", okResp.Game.GameID)
	}
	if len(okResp.Game.MainPlayers) != 1 {
		t.Fatalf("expected 1 main player, got %d", len(okResp.Game.MainPlayers))
	}
	if okResp.Game.MainPlayers[0].Username != "player1" {
		t.Errorf("expected username 'player1', got %q", okResp.Game.MainPlayers[0].Username)
	}
}

func TestPostLogin_AuthFailure(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{loginErr: errors.New("invalid credentials")},
		conf: &config.Config{},
	}
	resp, err := h.PostLogin(context.Background(), PostLoginRequestObject{
		Body: &PostLoginJSONRequestBody{Username: "user", Password: "wrong"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(PostLogin401JSONResponse); !ok {
		t.Errorf("expected 401 response, got %T", resp)
	}
}

func TestPostLogout(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{BasePath: "/"},
	}
	user := &db.User{UserID: 1}
	// Set session ID in context
	ctx := context.WithValue(context.Background(), sessionIDContextKey{}, "hashed-session")
	resp, err := h.PostLogout(ctx, PostLogoutRequestObject{}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(postLogoutCookieResponse); !ok {
		t.Errorf("expected postLogoutCookieResponse, got %T", resp)
	}
}

func TestGetGames_Empty(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGames(context.Background(), GetGamesRequestObject{}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGames200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if len(okResp.Games) != 0 {
		t.Errorf("expected 0 games, got %d", len(okResp.Games))
	}
}

func TestGetGames_WithGames(t *testing.T) {
	now := time.Now()
	h := Handler{
		q: &mockQuerier{
			listPublicGamesFunc: func(_ context.Context) ([]db.ListPublicGamesRow, error) {
				return []db.ListPublicGamesRow{
					{
						GameID:          1,
						GameType:        "golf",
						IsPublic:        true,
						DisplayName:     "Game 1",
						DurationSeconds: 300,
						StartedAt:       pgtype.Timestamp{Time: now, Valid: true},
						ProblemID:       10,
						Title:           "Problem 1",
						Description:     "desc",
						Language:        "php",
						SampleCode:      "<?php",
					},
				}, nil
			},
			listMainPlayersFunc: func(_ context.Context, _ []int32) ([]db.ListMainPlayersRow, error) {
				return nil, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGames(context.Background(), GetGamesRequestObject{}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGames200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if len(okResp.Games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(okResp.Games))
	}
	if okResp.Games[0].DisplayName != "Game 1" {
		t.Errorf("expected display name 'Game 1', got %q", okResp.Games[0].DisplayName)
	}
	if okResp.Games[0].StartedAt == nil {
		t.Error("expected non-nil StartedAt")
	}
}

func TestGetGamePlayLatestState_NoState(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGamePlayLatestState(context.Background(), GetGamePlayLatestStateRequestObject{GameID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGamePlayLatestState200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.State.Code != "" {
		t.Errorf("expected empty code, got %q", okResp.State.Code)
	}
	if okResp.State.Status != None {
		t.Errorf("expected status 'none', got %q", okResp.State.Status)
	}
}

func TestPostGamePlayCode_GameNotFound(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.PostGamePlayCode(context.Background(), PostGamePlayCodeRequestObject{
		GameID: 999,
		Body:   &PostGamePlayCodeJSONRequestBody{Code: "test"},
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(PostGamePlayCode404JSONResponse); !ok {
		t.Errorf("expected 404 response, got %T", resp)
	}
}

func TestPostGamePlayCode_GameNotRunning(t *testing.T) {
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:   1,
					Language: "php",
					StartedAt: pgtype.Timestamp{
						Valid: false,
					},
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.PostGamePlayCode(context.Background(), PostGamePlayCodeRequestObject{
		GameID: 1,
		Body:   &PostGamePlayCodeJSONRequestBody{Code: "<?php echo 1;"},
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(PostGamePlayCode403JSONResponse); !ok {
		t.Errorf("expected 403 response, got %T", resp)
	}
}

func TestPostGamePlayCode_Success(t *testing.T) {
	now := time.Now()
	var updatedCode string
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:          1,
					Language:        "php",
					StartedAt:       pgtype.Timestamp{Time: now, Valid: true},
					DurationSeconds: 600,
				}, nil
			},
			updateCodeFunc: func(_ context.Context, arg db.UpdateCodeParams) error {
				updatedCode = arg.Code
				return nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.PostGamePlayCode(context.Background(), PostGamePlayCodeRequestObject{
		GameID: 1,
		Body:   &PostGamePlayCodeJSONRequestBody{Code: "<?php echo 42;"},
	}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(PostGamePlayCode200Response); !ok {
		t.Errorf("expected 200 response, got %T", resp)
	}
	if updatedCode != "<?php echo 42;" {
		t.Errorf("expected code '<?php echo 42;', got %q", updatedCode)
	}
}

func TestGetGameWatchRanking_NotFound(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGameWatchRanking(context.Background(), GetGameWatchRankingRequestObject{GameID: 999}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(GetGameWatchRanking404JSONResponse); !ok {
		t.Errorf("expected 404 response, got %T", resp)
	}
}

func TestGetGameWatchRanking_EmptyRanking(t *testing.T) {
	now := time.Now()
	h := Handler{
		q: &mockQuerier{
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:          1,
					Language:        "php",
					StartedAt:       pgtype.Timestamp{Time: now.Add(-10 * time.Minute), Valid: true},
					DurationSeconds: 300,
				}, nil
			},
			getRankingFunc: func(_ context.Context, _ int32) ([]db.GetRankingRow, error) {
				return nil, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGameWatchRanking(context.Background(), GetGameWatchRankingRequestObject{GameID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGameWatchRanking200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if len(okResp.Ranking) != 0 {
		t.Errorf("expected empty ranking, got %d entries", len(okResp.Ranking))
	}
}

func TestGetGameWatchLatestStates_Empty(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetGameWatchLatestStates(context.Background(), GetGameWatchLatestStatesRequestObject{GameID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetGameWatchLatestStates200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if len(okResp.States) != 0 {
		t.Errorf("expected 0 states, got %d", len(okResp.States))
	}
}

func TestToNullableWith(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result := toNullableWith(nil, func(_ int) string { return "x" })
		if !result.IsNull() {
			t.Error("expected null for nil input")
		}
	})
	t.Run("non-nil value", func(t *testing.T) {
		x := 42
		result := toNullableWith(&x, func(_ int) string { return "hello" })
		if result.IsNull() {
			t.Error("expected non-null for non-nil input")
		}
		v, err := result.Get()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "hello" {
			t.Errorf("expected 'hello', got %q", v)
		}
	})
}

// --- Tournament tests ---

func TestStandardBracketSeeds(t *testing.T) {
	tests := []struct {
		name        string
		bracketSize int
		expected    []int
	}{
		{
			name:        "bracket_size=2",
			bracketSize: 2,
			expected:    []int{1, 2},
		},
		{
			name:        "bracket_size=4",
			bracketSize: 4,
			expected:    []int{1, 4, 2, 3},
		},
		{
			name:        "bracket_size=8",
			bracketSize: 8,
			expected:    []int{1, 8, 4, 5, 2, 7, 3, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := standardBracketSeeds(tt.bracketSize)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(got))
			}
			for i, v := range tt.expected {
				if got[i] != v {
					t.Errorf("position %d: expected seed %d, got %d", i, v, got[i])
				}
			}
		})
	}
}

func TestStandardBracketSeeds_Seed1And2OppositeSides(t *testing.T) {
	seeds := standardBracketSeeds(8)
	// Seed 1 should be in the first half, Seed 2 in the second half
	seed1Pos := -1
	seed2Pos := -1
	for i, s := range seeds {
		if s == 1 {
			seed1Pos = i
		}
		if s == 2 {
			seed2Pos = i
		}
	}
	if seed1Pos >= 4 {
		t.Errorf("Seed 1 should be in first half, but at position %d", seed1Pos)
	}
	if seed2Pos < 4 {
		t.Errorf("Seed 2 should be in second half, but at position %d", seed2Pos)
	}
}

func TestStandardBracketSeeds_AllSeedsPresent(t *testing.T) {
	for _, size := range []int{2, 4, 8, 16} {
		seeds := standardBracketSeeds(size)
		seen := make(map[int]bool)
		for _, s := range seeds {
			if s < 1 || s > size {
				t.Errorf("bracket_size=%d: seed %d out of range", size, s)
			}
			if seen[s] {
				t.Errorf("bracket_size=%d: duplicate seed %d", size, s)
			}
			seen[s] = true
		}
		if len(seen) != size {
			t.Errorf("bracket_size=%d: expected %d unique seeds, got %d", size, size, len(seen))
		}
	}
}

func TestFindSeedByUserID(t *testing.T) {
	entries := []TournamentEntry{
		{User: User{UserID: 10}, Seed: 1},
		{User: User{UserID: 20}, Seed: 2},
		{User: User{UserID: 30}, Seed: 3},
	}

	if got := findSeedByUserID(entries, 10); got != 1 {
		t.Errorf("expected seed 1 for user 10, got %d", got)
	}
	if got := findSeedByUserID(entries, 20); got != 2 {
		t.Errorf("expected seed 2 for user 20, got %d", got)
	}
	if got := findSeedByUserID(entries, 999); got != 0 {
		t.Errorf("expected seed 0 for unknown user, got %d", got)
	}
}

func TestGetTournament_NotFound(t *testing.T) {
	h := Handler{
		q:    &mockQuerier{},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetTournament(context.Background(), GetTournamentRequestObject{TournamentID: 999}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(GetTournament404JSONResponse); !ok {
		t.Errorf("expected 404 response, got %T", resp)
	}
}

func TestGetTournament_Success_NoEntries(t *testing.T) {
	h := Handler{
		q: &mockQuerier{
			getTournamentByIDFunc: func(_ context.Context, _ int32) (db.Tournament, error) {
				return db.Tournament{
					TournamentID: 1,
					DisplayName:  "Test Tournament",
					BracketSize:  4,
					NumRounds:    2,
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetTournament(context.Background(), GetTournamentRequestObject{TournamentID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetTournament200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Tournament.TournamentID != 1 {
		t.Errorf("expected tournament ID 1, got %d", okResp.Tournament.TournamentID)
	}
	if okResp.Tournament.DisplayName != "Test Tournament" {
		t.Errorf("expected display name 'Test Tournament', got %q", okResp.Tournament.DisplayName)
	}
	if okResp.Tournament.BracketSize != 4 {
		t.Errorf("expected bracket size 4, got %d", okResp.Tournament.BracketSize)
	}
	if len(okResp.Tournament.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(okResp.Tournament.Entries))
	}
}

func TestGetTournament_WithEntriesAndMatches(t *testing.T) {
	gameID := int32(10)
	h := Handler{
		q: &mockQuerier{
			getTournamentByIDFunc: func(_ context.Context, _ int32) (db.Tournament, error) {
				return db.Tournament{
					TournamentID: 1,
					DisplayName:  "Test",
					BracketSize:  4,
					NumRounds:    2,
				}, nil
			},
			listTournamentEntriesFunc: func(_ context.Context, _ int32) ([]db.ListTournamentEntriesRow, error) {
				return []db.ListTournamentEntriesRow{
					{Seed: 1, UserID: 100, Username: "alice", DisplayName: "Alice", IsAdmin: false},
					{Seed: 2, UserID: 200, Username: "bob", DisplayName: "Bob", IsAdmin: false},
					{Seed: 3, UserID: 300, Username: "carol", DisplayName: "Carol", IsAdmin: false},
				}, nil
			},
			listTournamentMatchesFunc: func(_ context.Context, _ int32) ([]db.TournamentMatch, error) {
				return []db.TournamentMatch{
					{TournamentMatchID: 1, TournamentID: 1, Round: 0, Position: 0, GameID: &gameID},
					{TournamentMatchID: 2, TournamentID: 1, Round: 0, Position: 1, GameID: nil},
					{TournamentMatchID: 3, TournamentID: 1, Round: 1, Position: 0, GameID: nil},
				}, nil
			},
			getGameByIDFunc: func(_ context.Context, _ int32) (db.GetGameByIDRow, error) {
				return db.GetGameByIDRow{
					GameID:    10,
					StartedAt: pgtype.Timestamp{Valid: false},
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetTournament(context.Background(), GetTournamentRequestObject{TournamentID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := resp.(GetTournament200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}

	// Check entries
	if len(okResp.Tournament.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(okResp.Tournament.Entries))
	}

	// Check matches: bracket_size=4, num_rounds=2 → round 0: 2 matches, round 1: 1 match
	if len(okResp.Tournament.Matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(okResp.Tournament.Matches))
	}

	// Round 0, Position 0: Seed 1 (Alice) vs Seed 4 (bye)
	m0 := okResp.Tournament.Matches[0]
	if m0.Round != 0 || m0.Position != 0 {
		t.Errorf("match 0: expected round=0, pos=0, got round=%d, pos=%d", m0.Round, m0.Position)
	}
	if m0.Player1 == nil || m0.Player1.Username != "alice" {
		t.Errorf("match 0 player1: expected alice")
	}
	if m0.Player2 != nil {
		t.Errorf("match 0 player2: expected nil (bye), got %v", m0.Player2)
	}
	if !m0.IsBye {
		t.Error("match 0: expected is_bye=true")
	}
	if m0.WinnerUserID == nil || *m0.WinnerUserID != 100 {
		t.Error("match 0: expected winner to be Alice (user_id=100)")
	}

	// Round 0, Position 1: Seed 2 (Bob) vs Seed 3 (Carol)
	m1 := okResp.Tournament.Matches[1]
	if m1.Round != 0 || m1.Position != 1 {
		t.Errorf("match 1: expected round=0, pos=1, got round=%d, pos=%d", m1.Round, m1.Position)
	}
	if m1.Player1 == nil || m1.Player1.Username != "bob" {
		t.Errorf("match 1 player1: expected bob, got %v", m1.Player1)
	}
	if m1.Player2 == nil || m1.Player2.Username != "carol" {
		t.Errorf("match 1 player2: expected carol, got %v", m1.Player2)
	}
	if m1.IsBye {
		t.Error("match 1: expected is_bye=false")
	}
}

func TestGetTournament_ByeAutoWinner(t *testing.T) {
	// 3 players in bracket_size=4: seed 4 is empty → round 0, pos 0 is a bye
	// The bye winner should propagate to round 1
	h := Handler{
		q: &mockQuerier{
			getTournamentByIDFunc: func(_ context.Context, _ int32) (db.Tournament, error) {
				return db.Tournament{
					TournamentID: 1,
					DisplayName:  "Bye Test",
					BracketSize:  4,
					NumRounds:    2,
				}, nil
			},
			listTournamentEntriesFunc: func(_ context.Context, _ int32) ([]db.ListTournamentEntriesRow, error) {
				return []db.ListTournamentEntriesRow{
					{Seed: 1, UserID: 100, Username: "alice", DisplayName: "Alice"},
					{Seed: 2, UserID: 200, Username: "bob", DisplayName: "Bob"},
					{Seed: 3, UserID: 300, Username: "carol", DisplayName: "Carol"},
				}, nil
			},
			listTournamentMatchesFunc: func(_ context.Context, _ int32) ([]db.TournamentMatch, error) {
				return []db.TournamentMatch{
					{TournamentMatchID: 1, Round: 0, Position: 0},
					{TournamentMatchID: 2, Round: 0, Position: 1},
					{TournamentMatchID: 3, Round: 1, Position: 0},
				}, nil
			},
		},
		txm:  &mockTxManager{},
		hub:  &mockGameHub{},
		auth: &mockAuthenticator{},
		conf: &config.Config{},
	}
	user := &db.User{UserID: 1}
	resp, err := h.GetTournament(context.Background(), GetTournamentRequestObject{TournamentID: 1}, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp := resp.(GetTournament200JSONResponse)

	// Round 1, Position 0 (final): player1 should be Alice (bye winner from round 0 pos 0)
	final := okResp.Tournament.Matches[2]
	if final.Round != 1 || final.Position != 0 {
		t.Fatalf("expected final at round=1, pos=0")
	}
	if final.Player1 == nil || final.Player1.UserID != 100 {
		t.Error("final player1: expected Alice (bye winner)")
	}
}
