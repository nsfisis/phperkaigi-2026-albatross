package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"albatross-2026-backend/db"
	"albatross-2026-backend/session"
)

// mockSessionQuerier implements the subset of db.Querier needed by SessionCookieMiddleware.
type mockSessionQuerier struct {
	db.Querier
	getUserBySessionFunc func(ctx context.Context, sessionID string) (db.User, error)
}

func (m *mockSessionQuerier) GetUserBySession(ctx context.Context, sessionID string) (db.User, error) {
	if m.getUserBySessionFunc != nil {
		return m.getUserBySessionFunc(ctx, sessionID)
	}
	return db.User{}, nil
}

func TestSessionCookieMiddleware_NoCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := SessionCookieMiddleware(&mockSessionQuerier{})
	var called bool
	handler := mw(func(c echo.Context) error {
		called = true
		// User should not be set
		_, ok := session.GetUserFromContext(c.Request().Context())
		if ok {
			t.Error("expected no user in context when no cookie is present")
		}
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("next handler was not called")
	}
}

func TestSessionCookieMiddleware_ValidSession(t *testing.T) {
	expectedUser := db.User{UserID: 10, Username: "sessionuser"}
	mq := &mockSessionQuerier{
		getUserBySessionFunc: func(_ context.Context, _ string) (db.User, error) {
			return expectedUser, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "albatross_session", Value: "raw-session-id"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := SessionCookieMiddleware(mq)
	var called bool
	handler := mw(func(c echo.Context) error {
		called = true
		user, ok := session.GetUserFromContext(c.Request().Context())
		if !ok {
			t.Fatal("expected user in context")
		}
		if user.UserID != 10 {
			t.Errorf("expected user ID 10, got %d", user.UserID)
		}
		sid, ok := session.GetSessionIDFromContext(c.Request().Context())
		if !ok {
			t.Fatal("expected session ID in context")
		}
		if sid == "" {
			t.Error("expected non-empty hashed session ID")
		}
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("next handler was not called")
	}
}

func TestSessionCookieMiddleware_InvalidSession(t *testing.T) {
	mq := &mockSessionQuerier{
		getUserBySessionFunc: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, echo.ErrNotFound
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "albatross_session", Value: "invalid-session"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := SessionCookieMiddleware(mq)
	var called bool
	handler := mw(func(c echo.Context) error {
		called = true
		_, ok := session.GetUserFromContext(c.Request().Context())
		if ok {
			t.Error("expected no user in context for invalid session")
		}
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("next handler was not called")
	}
}
