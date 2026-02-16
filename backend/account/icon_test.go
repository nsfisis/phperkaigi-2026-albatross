package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFile_Success(t *testing.T) {
	expectedContent := "file content here"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "test.png")

	err := downloadFile(context.Background(), server.URL+"/icon.png", filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, string(data))
	}
}

func TestDownloadFile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")

	err := downloadFile(context.Background(), server.URL+"/missing.png", filePath)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestDownloadFile_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")

	err := downloadFile(context.Background(), server.URL+"/error.png", filePath)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestDownloadFile_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")

	err := downloadFile(context.Background(), "http://localhost:1/unreachable", filePath)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestDownloadFile_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data"))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")

	err := downloadFile(ctx, server.URL+"/icon.png", filePath)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}
