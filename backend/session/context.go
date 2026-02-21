package session

import (
	"context"

	"albatross-2026-backend/db"
)

type sessionIDContextKey struct{}
type userContextKey struct{}
type clientIPContextKey struct{}

func GetSessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDContextKey{}).(string)
	return sessionID, ok
}

func SetSessionIDInContext(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDContextKey{}, sessionID)
}

func GetUserFromContext(ctx context.Context) (*db.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(*db.User)
	return user, ok
}

// SetUserInContext sets a user in the context. Intended for testing.
func SetUserInContext(ctx context.Context, user *db.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func GetClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPContextKey{}).(string)
	return ip
}

func SetClientIPInContext(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPContextKey{}, ip)
}
