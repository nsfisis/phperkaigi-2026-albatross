package api

import (
	"github.com/labstack/echo/v4"

	"albatross-2026-backend/auth"
	"albatross-2026-backend/db"
	"albatross-2026-backend/session"
)

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
			ctx = session.SetSessionIDInContext(ctx, hashedID)
			ctx = session.SetUserInContext(ctx, &user)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// ClientIPMiddleware extracts the client IP from echo.Context.RealIP()
// and stores it in the request's context.Context so that handlers
// receiving only context.Context (via generated code) can access it.
func ClientIPMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			ctx := session.SetClientIPInContext(c.Request().Context(), ip)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
