package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func newBadRequestError(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid request: %s", err.Error()))
}

func handleExec(c echo.Context) error {
	var req execRequestData
	if err := c.Bind(&req); err != nil {
		return newBadRequestError(err)
	}
	if err := req.validate(); err != nil {
		return newBadRequestError(err)
	}

	res := doExec(
		c.Request().Context(),
		req.Code,
		req.CodeHash,
		req.Stdin,
		req.maxDuration(),
	)

	return c.JSON(http.StatusOK, res)
}
