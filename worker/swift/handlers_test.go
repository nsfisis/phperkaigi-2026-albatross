package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNewBadRequestError(t *testing.T) {
	httpErr := newBadRequestError(errors.New("test error"))
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
	msg, ok := httpErr.Message.(string)
	if !ok {
		t.Fatalf("expected string message, got %T", httpErr.Message)
	}
	if !strings.Contains(msg, "test error") {
		t.Errorf("message = %q, want it to contain %q", msg, "test error")
	}
}

func TestHandleExec_InvalidJSON(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/exec", strings.NewReader("not json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleExec(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
}

func TestHandleExec_ZeroMaxDuration(t *testing.T) {
	e := echo.New()
	body := `{"code":"print(1)","code_hash":"abc","stdin":"","max_duration_ms":0}`
	req := httptest.NewRequest(http.MethodPost, "/exec", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleExec(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
}
