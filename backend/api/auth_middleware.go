package api

import (
	"context"

	"github.com/labstack/echo/v4"

	"albatross-2026-backend/auth"
	"albatross-2026-backend/db"
)

type sessionIDContextKey struct{}
type userContextKey struct{}

func SessionCookieMiddleware(q db.Querier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("albatross_session")
			if err != nil {
				return next(c)
			}
			hashedID := auth.HashSessionID(cookie.Value)
			user, err := q.GetUserBySession(c.Request().Context(), hashedID)
			if err != nil {
				return next(c)
			}
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, sessionIDContextKey{}, hashedID)
			ctx = context.WithValue(ctx, userContextKey{}, &user)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func GetSessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDContextKey{}).(string)
	return sessionID, ok
}

func GetUserFromContext(ctx context.Context) (*db.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(*db.User)
	return user, ok
}
