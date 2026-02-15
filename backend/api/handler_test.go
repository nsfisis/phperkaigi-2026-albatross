package api

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"albatross-2026-backend/config"
	"albatross-2026-backend/db"
)

// mockQuerier implements db.Querier for testing.
type mockQuerier struct {
	db.Querier
	getGameByIDFunc func(ctx context.Context, gameID int32) (db.GetGameByIDRow, error)
}

func (m *mockQuerier) GetGameByID(ctx context.Context, gameID int32) (db.GetGameByIDRow, error) {
	if m.getGameByIDFunc != nil {
		return m.getGameByIDFunc(ctx, gameID)
	}
	return db.GetGameByIDRow{}, pgx.ErrNoRows
}

// mockTxManager implements db.TxManager for testing.
type mockTxManager struct{}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(q db.Querier) error) error {
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
