package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"albatross-2026-backend/db"
)

type mockQuerier struct {
	db.Querier
	getUserAuthByUsernameFunc func(ctx context.Context, username string) (db.GetUserAuthByUsernameRow, error)
}

func (m *mockQuerier) GetUserAuthByUsername(ctx context.Context, username string) (db.GetUserAuthByUsernameRow, error) {
	if m.getUserAuthByUsernameFunc != nil {
		return m.getUserAuthByUsernameFunc(ctx, username)
	}
	return db.GetUserAuthByUsernameRow{}, pgx.ErrNoRows
}

type mockTxManager struct{}

func (m *mockTxManager) RunInTx(_ context.Context, fn func(q db.Querier) error) error {
	return fn(&mockQuerier{})
}

func TestLogin_PasswordAuth_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to generate hash: %v", err)
	}
	hashStr := string(hash)

	a := NewAuthenticator(
		&mockQuerier{
			getUserAuthByUsernameFunc: func(_ context.Context, _ string) (db.GetUserAuthByUsernameRow, error) {
				return db.GetUserAuthByUsernameRow{
					UserID:       42,
					AuthType:     "password",
					PasswordHash: &hashStr,
				}, nil
			},
		},
		&mockTxManager{},
	)

	userID, err := a.Login(context.Background(), "testuser", "correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 42 {
		t.Errorf("expected userID 42, got %d", userID)
	}
}

func TestLogin_PasswordAuth_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to generate hash: %v", err)
	}
	hashStr := string(hash)

	a := NewAuthenticator(
		&mockQuerier{
			getUserAuthByUsernameFunc: func(_ context.Context, _ string) (db.GetUserAuthByUsernameRow, error) {
				return db.GetUserAuthByUsernameRow{
					UserID:       42,
					AuthType:     "password",
					PasswordHash: &hashStr,
				}, nil
			},
		},
		&mockTxManager{},
	)

	_, err = a.Login(context.Background(), "testuser", "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestLogin_PasswordAuth_NilHash(t *testing.T) {
	a := NewAuthenticator(
		&mockQuerier{
			getUserAuthByUsernameFunc: func(_ context.Context, _ string) (db.GetUserAuthByUsernameRow, error) {
				return db.GetUserAuthByUsernameRow{
					UserID:       42,
					AuthType:     "password",
					PasswordHash: nil,
				}, nil
			},
		},
		&mockTxManager{},
	)

	_, err := a.Login(context.Background(), "testuser", "any")
	if err == nil {
		t.Fatal("expected error for nil password hash, got nil")
	}
	if err.Error() != "inconsistent data: password auth type but no password hash" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLogin_DBError(t *testing.T) {
	dbErr := errors.New("database connection failed")
	a := NewAuthenticator(
		&mockQuerier{
			getUserAuthByUsernameFunc: func(_ context.Context, _ string) (db.GetUserAuthByUsernameRow, error) {
				return db.GetUserAuthByUsernameRow{}, dbErr
			},
		},
		&mockTxManager{},
	)

	_, err := a.Login(context.Background(), "testuser", "any")
	if !errors.Is(err, dbErr) {
		t.Errorf("expected db error, got: %v", err)
	}
}
