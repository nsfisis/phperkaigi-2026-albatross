package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

func TestGetLimiter_SameIP(t *testing.T) {
	rl := &IPRateLimiter{
		rate:  rate.Limit(10),
		burst: 1,
	}

	l1 := rl.getLimiter("192.168.1.1")
	l2 := rl.getLimiter("192.168.1.1")
	if l1 != l2 {
		t.Error("expected same limiter for same IP")
	}
}

func TestGetLimiter_DifferentIP(t *testing.T) {
	rl := &IPRateLimiter{
		rate:  rate.Limit(10),
		burst: 1,
	}

	l1 := rl.getLimiter("192.168.1.1")
	l2 := rl.getLimiter("192.168.1.2")
	if l1 == l2 {
		t.Error("expected different limiters for different IPs")
	}
}

func TestLoginRateLimitMiddleware_AllowsNonLogin(t *testing.T) {
	rl := &IPRateLimiter{
		rate:  rate.Limit(0), // zero rate = deny all
		burst: 0,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/games")

	handler := LoginRateLimitMiddleware(rl)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestLoginRateLimitMiddleware_BlocksExcessiveLogin(t *testing.T) {
	rl := &IPRateLimiter{
		rate:  rate.Limit(0.001), // very low rate
		burst: 1,
	}

	e := echo.New()

	// First request should succeed (burst = 1)
	req1 := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	c1.SetPath("/api/login")

	handler := LoginRateLimitMiddleware(rl)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec1.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rec1.Code)
	}

	// Second request should be rate limited
	req2 := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetPath("/api/login")

	if err := handler(c2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rec2.Code)
	}
}

func TestLoginRateLimitMiddleware_AllowsNonPostLogin(t *testing.T) {
	rl := &IPRateLimiter{
		rate:  rate.Limit(0),
		burst: 0,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/login")

	handler := LoginRateLimitMiddleware(rl)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for GET /login, got %d", rec.Code)
	}
}
