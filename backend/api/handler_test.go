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
	getGameByIDFunc     func(ctx context.Context, gameID int32) (db.GetGameByIDRow, error)
	listMainPlayersFunc func(ctx context.Context, gameIDs []int32) ([]db.ListMainPlayersRow, error)
	listPublicGamesFunc func(ctx context.Context) ([]db.ListPublicGamesRow, error)
	deleteSessionFunc   func(ctx context.Context, sessionID string) error
	getLatestStateFunc  func(ctx context.Context, arg db.GetLatestStateParams) (db.GetLatestStateRow, error)
	updateCodeFunc      func(ctx context.Context, arg db.UpdateCodeParams) error
	getRankingFunc      func(ctx context.Context, gameID int32) ([]db.GetRankingRow, error)
	getLatestStatesFunc func(ctx context.Context, gameID int32) ([]db.GetLatestStatesOfMainPlayersRow, error)
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
		result := toNullableWith[int, string](nil, func(_ int) string { return "x" })
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
