package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
		log.Fatalf("Error loading env %v", err)
	}

	openAPISpec, err := api.GetSwaggerWithPrefix(conf.BasePath + "api")
	if err != nil {
		log.Fatalf("Error loading OpenAPI spec\n: %s", err)
	}

	ctx := context.Background()

	dbDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", conf.DBHost, conf.DBPort, conf.DBUser, conf.DBPassword, conf.DBName)
	connPool, err := connectDB(ctx, dbDSN)
	if err != nil {
		log.Fatalf("Error connecting to db %v", err)
	}
	defer connPool.Close()

	queries := db.New(connPool)

	e := echo.New()
	e.Renderer = admin.NewRenderer()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	taskQueue := taskqueue.NewQueue("task-db:6379")
	workerServer := taskqueue.NewWorkerServer("task-db:6379")

	gameHub := game.NewGameHub(queries, taskQueue, workerServer)

	loginRL := ratelimit.NewIPRateLimiter(rate.Every(time.Minute/5), 5)

	apiGroup := e.Group(conf.BasePath + "api")
	apiGroup.Use(ratelimit.LoginRateLimitMiddleware(loginRL))
	apiGroup.Use(api.SessionCookieMiddleware(queries))
	apiGroup.Use(oapimiddleware.OapiRequestValidator(openAPISpec))
	apiHandler := api.NewHandler(queries, gameHub, conf)
	api.RegisterHandlers(apiGroup, api.NewStrictHandler(apiHandler, nil))

	adminHandler := admin.NewHandler(queries, conf)
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
					log.Printf("failed to delete expired sessions: %v", err)
				}
			}
		}
	}()

	go gameHub.Run()

	if err := e.Start(":80"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
