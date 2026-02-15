package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"
	"golang.org/x/time/rate"

	"albatross-2026-backend/admin"
	"albatross-2026-backend/api"
	"albatross-2026-backend/config"
	"albatross-2026-backend/db"
	"albatross-2026-backend/game"
	"albatross-2026-backend/ratelimit"
	"albatross-2026-backend/taskqueue"
)

func connectDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func main() {
	var err error
	conf, err := config.NewConfigFromEnv()
	if err != nil {
		slog.Error("failed to load env", "error", err)
		os.Exit(1)
	}

	openAPISpec, err := api.GetSwaggerWithPrefix(conf.BasePath + "api")
	if err != nil {
		slog.Error("failed to load OpenAPI spec", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	dbDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", conf.DBHost, conf.DBPort, conf.DBUser, conf.DBPassword, conf.DBName)
	connPool, err := connectDB(ctx, dbDSN)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer connPool.Close()

	queries := db.New(connPool)

	e := echo.New()
	e.Renderer = admin.NewRenderer()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
			}
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
			}
			slog.LogAttrs(context.Background(), slog.LevelInfo, "request", attrs...)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	taskQueue := taskqueue.NewQueue("task-db:6379")
	workerServer := taskqueue.NewWorkerServer("task-db:6379")

	gameHub := game.NewGameHub(queries, connPool, taskQueue, workerServer)

	loginRL := ratelimit.NewIPRateLimiter(rate.Every(time.Minute/5), 5)

	apiGroup := e.Group(conf.BasePath + "api")
	apiGroup.Use(ratelimit.LoginRateLimitMiddleware(loginRL))
	apiGroup.Use(api.SessionCookieMiddleware(queries))
	apiGroup.Use(oapimiddleware.OapiRequestValidator(openAPISpec))
	apiHandler := api.NewHandler(queries, connPool, gameHub, conf)
	api.RegisterHandlers(apiGroup, api.NewStrictHandler(apiHandler, nil))

	adminHandler := admin.NewHandler(queries, connPool, conf)
	adminGroup := e.Group(conf.BasePath + "admin")
	adminGroup.Use(api.SessionCookieMiddleware(queries))
	adminHandler.RegisterHandlers(adminGroup)

	if conf.IsLocal {
		filesGroup := e.Group(conf.BasePath + "files")
		filesGroup.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:       "/",
			Filesystem: http.Dir("/data/files"),
			IgnoreBase: true,
		}))

		e.GET(conf.BasePath+"*", func(c echo.Context) error {
			return c.Redirect(http.StatusPermanentRedirect, "http://localhost:5173"+c.Request().URL.Path)
		})
		e.POST(conf.BasePath+"*", func(c echo.Context) error {
			return c.Redirect(http.StatusPermanentRedirect, "http://localhost:5173"+c.Request().URL.Path)
		})

		// Allow access from dev server.
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"http://localhost:5173"},
			AllowCredentials: true,
		}))
	}

	sessionCleanupCtx, cancelSessionCleanup := context.WithCancel(context.Background())
	defer cancelSessionCleanup()
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-sessionCleanupCtx.Done():
				return
			case <-ticker.C:
				if err := queries.DeleteExpiredSessions(sessionCleanupCtx); err != nil {
					slog.Error("failed to delete expired sessions", "error", err)
				}
			}
		}
	}()

	go gameHub.Run()

	if err := e.Start(":80"); err != http.ErrServerClosed {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
