package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/nsfisis/iosdc-japan-2025-albatross/backend/account"
	"github.com/nsfisis/iosdc-japan-2025-albatross/backend/db"
	"github.com/nsfisis/iosdc-japan-2025-albatross/backend/fortee"
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
	return verifyForteeAccountOrSignup(ctx, queries, username, password)
}

func verifyForteeAccountOrSignup(
	ctx context.Context,
	queries *db.Queries,
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
	username string,
) (int, error) {
	// TODO: transaction
	userID, err := queries.CreateUser(ctx, username)
	if err != nil {
		return 0, err
	}
	if err := queries.CreateUserAuth(ctx, db.CreateUserAuthParams{
		UserID:   userID,
		AuthType: "fortee",
	}); err != nil {
		return 0, err
	}
	go func() {
		err := account.FetchIcon(context.Background(), queries, int(userID))
		if err != nil {
			log.Printf("%v", err)
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
