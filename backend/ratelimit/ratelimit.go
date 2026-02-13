package ratelimit

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	entries sync.Map
	rate    rate.Limit
	burst   int
}

func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		rate:  r,
		burst: burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	now := time.Now()
	if v, ok := rl.entries.Load(ip); ok {
		e := v.(*entry)
		e.lastSeen = now
		return e.limiter
	}
	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.entries.Store(ip, &entry{limiter: limiter, lastSeen: now})
	return limiter
}

func (rl *IPRateLimiter) cleanup() {
	for {
		time.Sleep(10 * time.Minute)
		rl.entries.Range(func(key, value any) bool {
			e := value.(*entry)
			if time.Since(e.lastSeen) > 10*time.Minute {
				rl.entries.Delete(key)
			}
			return true
		})
	}
}

func LoginRateLimitMiddleware(rl *IPRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method != http.MethodPost || !strings.HasSuffix(c.Path(), "/login") {
				return next(c)
			}
			ip := c.RealIP()
			if !rl.getLimiter(ip).Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"message": "Too many login attempts. Please try again later.",
				})
			}
			return next(c)
		}
	}
}
