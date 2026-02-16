package fortee

import (
	"context"
	"net/http"
	"testing"
)

func TestAddAcceptHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if err := addAcceptHeader(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := req.Header.Get("Accept")
	if got != "application/json" {
		t.Errorf("expected Accept header 'application/json', got %q", got)
	}
}

func TestEndpoint(t *testing.T) {
	if Endpoint != "https://fortee.jp" {
		t.Errorf("expected endpoint 'https://fortee.jp', got %q", Endpoint)
	}
}

func TestErrorValues(t *testing.T) {
	if ErrLoginFailed == nil {
		t.Error("ErrLoginFailed should not be nil")
	}
	if ErrUserNotFound == nil {
		t.Error("ErrUserNotFound should not be nil")
	}
	if ErrLoginFailed.Error() != "fortee login failed" {
		t.Errorf("unexpected error message: %q", ErrLoginFailed.Error())
	}
	if ErrUserNotFound.Error() != "fortee user not found" {
		t.Errorf("unexpected error message: %q", ErrUserNotFound.Error())
	}
}
