package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"albatross-2026-backend/account"
	"albatross-2026-backend/db"
	"albatross-2026-backend/fortee"
)

var (
	ErrForteeLoginTimeout = errors.New("fortee login timeout")
)

const (
	forteeAPITimeout = 3 * time.Second
)

func Login(
	ctx context.Context,
	queries *db.Queries,
	pool *pgxpool.Pool,
	username string,
	password string,
) (int, error) {
	userAuth, err := queries.GetUserAuthByUsername(ctx, username)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	if userAuth.AuthType == "password" {
		// Authenticate with password.
		passwordHash := userAuth.PasswordHash
		if passwordHash == nil {
			panic("inconsistant data")
		}
		err := bcrypt.CompareHashAndPassword([]byte(*passwordHash), []byte(password))
		if err != nil {
			return 0, err
		}
		return int(userAuth.UserID), nil
	}

	// Authenticate with fortee.
	return verifyForteeAccountOrSignup(ctx, queries, pool, username, password)
}

func verifyForteeAccountOrSignup(
	ctx context.Context,
	queries *db.Queries,
	pool *pgxpool.Pool,
	username string,
	password string,
) (int, error) {
	canonicalizedUsername, err := verifyForteeAccount(ctx, username, password)
	if err != nil {
		return 0, err
	}
	userID, err := queries.GetUserIDByUsername(ctx, canonicalizedUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return signup(
				ctx,
				queries,
				pool,
				canonicalizedUsername,
			)
		}
		return 0, err
	}
	return int(userID), nil
}

func signup(
	ctx context.Context,
	queries *db.Queries,
	pool *pgxpool.Pool,
	username string,
) (int, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	qtx := queries.WithTx(tx)
	userID, err := qtx.CreateUser(ctx, username)
	if err != nil {
		return 0, err
	}
	if err := qtx.CreateUserAuth(ctx, db.CreateUserAuthParams{
		UserID:   userID,
		AuthType: "fortee",
	}); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	go func() {
		err := account.FetchIcon(context.Background(), queries, int(userID))
		if err != nil {
			slog.Error("failed to fetch icon", "error", err)
			// The failure is intentionally ignored. Retry manually if needed.
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
