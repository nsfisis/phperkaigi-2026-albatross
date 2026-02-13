package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := prepareDirectories(); err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.POST("/exec", handleExec)

	if err := e.Start(":80"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
