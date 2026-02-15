package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"albatross-2026-backend/account"
	"albatross-2026-backend/db"
	"albatross-2026-backend/fortee"
)

var ErrForteeLoginTimeout = errors.New("fortee login timeout")

const (
	forteeAPITimeout = 3 * time.Second
)

type Authenticator struct {
	q   db.Querier
	txm db.TxManager
}

func NewAuthenticator(q db.Querier, txm db.TxManager) *Authenticator {
	return &Authenticator{q: q, txm: txm}
}

func (a *Authenticator) Login(
	ctx context.Context,
	username string,
	password string,
) (int, error) {
	userAuth, err := a.q.GetUserAuthByUsername(ctx, username)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	if userAuth.AuthType == "password" {
		passwordHash := userAuth.PasswordHash
		if passwordHash == nil {
			return 0, errors.New("inconsistent data: password auth type but no password hash")
		}
		err := bcrypt.CompareHashAndPassword([]byte(*passwordHash), []byte(password))
		if err != nil {
			return 0, err
		}
		return int(userAuth.UserID), nil
	}

	return a.verifyForteeAccountOrSignup(ctx, username, password)
}

func (a *Authenticator) verifyForteeAccountOrSignup(
	ctx context.Context,
	username string,
	password string,
) (int, error) {
	canonicalizedUsername, err := verifyForteeAccount(ctx, username, password)
	if err != nil {
		return 0, err
	}
	userID, err := a.q.GetUserIDByUsername(ctx, canonicalizedUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return a.signup(ctx, canonicalizedUsername)
		}
		return 0, err
	}
	return int(userID), nil
}

func (a *Authenticator) signup(
	ctx context.Context,
	username string,
) (int, error) {
	var userID int32
	err := a.txm.RunInTx(ctx, func(qtx db.Querier) error {
		var err error
		userID, err = qtx.CreateUser(ctx, username)
		if err != nil {
			return err
		}
		return qtx.CreateUserAuth(ctx, db.CreateUserAuthParams{
			UserID:   userID,
			AuthType: "fortee",
		})
	})
	if err != nil {
		return 0, err
	}

	go func() {
		err := account.FetchIcon(context.Background(), a.q, int(userID))
		if err != nil {
			slog.Error("failed to fetch icon", "error", err)
		}
	}()
	return int(userID), nil
}

func verifyForteeAccount(ctx context.Context, username string, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, forteeAPITimeout)
	defer cancel()

	canonicalizedUsername, err := fortee.Login(ctx, username, password)
	if errors.Is(err, context.DeadlineExceeded) {
		return "", ErrForteeLoginTimeout
	}
	return canonicalizedUsername, err
}
