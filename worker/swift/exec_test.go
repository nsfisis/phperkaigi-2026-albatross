package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestConvertCommandErrorToResultType(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		defaultStatus string
		want          string
	}{
		{"nil error returns success", nil, resultRuntimeError, resultSuccess},
		{"DeadlineExceeded returns timeout", context.DeadlineExceeded, resultRuntimeError, resultTimeout},
		{"other error returns default status", os.ErrNotExist, resultCompileError, resultCompileError},
		{"other error returns runtime_error default", os.ErrPermission, resultRuntimeError, resultRuntimeError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertCommandErrorToResultType(tt.err, tt.defaultStatus)
			if got != tt.want {
				t.Errorf("convertCommandErrorToResultType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecCommandWithTimeout_Success(t *testing.T) {
	stdout, stderr, err := execCommandWithTimeout(
		context.Background(),
		t.TempDir(),
		5*time.Second,
		func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "echo", "hello")
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "hello\n" {
		t.Errorf("stdout = %q, want %q", stdout, "hello\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestExecCommandWithTimeout_Failure(t *testing.T) {
	_, _, err := execCommandWithTimeout(
		context.Background(),
		t.TempDir(),
		5*time.Second,
		func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "false")
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecCommandWithTimeout_Timeout(t *testing.T) {
	_, _, err := execCommandWithTimeout(
		context.Background(),
		t.TempDir(),
		50*time.Millisecond,
		func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "sleep", "10")
		},
	)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestExecCommandWithTimeout_Stderr(t *testing.T) {
	_, stderr, _ := execCommandWithTimeout(
		context.Background(),
		t.TempDir(),
		5*time.Second,
		func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "sh", "-c", "echo errmsg >&2")
		},
	)
	if stderr != "errmsg\n" {
		t.Errorf("stderr = %q, want %q", stderr, "errmsg\n")
	}
}

func TestPrepareWorkingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "subdir")
	res := prepareWorkingDir(dir)
	if res.Status != resultSuccess {
		t.Fatalf("prepareWorkingDir() status = %q, want %q", res.Status, resultSuccess)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory, got file")
	}
}

func TestPutSwiftSourceFile(t *testing.T) {
	t.Run("writes file when Sources/ exists", func(t *testing.T) {
		dir := t.TempDir()
		sourcesDir := filepath.Join(dir, "Sources")
		if err := os.MkdirAll(sourcesDir, 0755); err != nil {
			t.Fatalf("failed to create Sources dir: %v", err)
		}
		res := putSwiftSourceFile(dir, "print(\"hello\")")
		if res.Status != resultSuccess {
			t.Fatalf("putSwiftSourceFile() status = %q, want %q", res.Status, resultSuccess)
		}
		content, err := os.ReadFile(filepath.Join(sourcesDir, "main.swift"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != "print(\"hello\")" {
			t.Errorf("file content = %q, want %q", string(content), "print(\"hello\")")
		}
	})

	t.Run("returns error when Sources/ does not exist", func(t *testing.T) {
		dir := t.TempDir()
		res := putSwiftSourceFile(dir, "print(\"hello\")")
		if res.Status == resultSuccess {
			t.Fatal("expected error status, got success")
		}
	})
}

func TestRemoveWorkingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "toremove")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	removeWorkingDir(dir)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("directory still exists after removeWorkingDir")
	}
}
