package api

import (
	"context"

	"github.com/labstack/echo/v4"

	"albatross-2026-backend/auth"
)

type contextKey struct{}

func JWTCookieMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("albatross_token")
		if err != nil {
			return next(c)
		}
		claims, err := auth.ParseJWT(cookie.Value)
		if err != nil {
			return next(c)
		}
		ctx := context.WithValue(c.Request().Context(), contextKey{}, claims)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func GetJWTClaimsFromContext(ctx context.Context) (*auth.JWTClaims, bool) {
	claims, ok := ctx.Value(contextKey{}).(*auth.JWTClaims)
	return claims, ok
}
