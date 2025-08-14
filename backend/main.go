package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"

	"github.com/nsfisis/phperkaigi-2025-albatross/backend/admin"
	"github.com/nsfisis/phperkaigi-2025-albatross/backend/api"
	"github.com/nsfisis/phperkaigi-2025-albatross/backend/config"
	"github.com/nsfisis/phperkaigi-2025-albatross/backend/db"
	"github.com/nsfisis/phperkaigi-2025-albatross/backend/game"
	"github.com/nsfisis/phperkaigi-2025-albatross/backend/taskqueue"
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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	taskQueue := taskqueue.NewQueue("task-db:6379")
	workerServer := taskqueue.NewWorkerServer("task-db:6379")

	gameHub := game.NewGameHub(queries, taskQueue, workerServer)

	apiGroup := e.Group(conf.BasePath + "api")
	apiGroup.Use(oapimiddleware.OapiRequestValidator(openAPISpec))
	apiHandler := api.NewHandler(queries, gameHub)
	api.RegisterHandlers(apiGroup, api.NewStrictHandler(apiHandler, nil))

	adminHandler := admin.NewHandler(queries, conf)
	adminGroup := e.Group(conf.BasePath + "admin")
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
		e.Use(middleware.CORS())
	}

	go gameHub.Run()

	if err := e.Start(":80"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
