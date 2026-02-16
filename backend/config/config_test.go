package config

import (
	"testing"
)

func TestNewConfigFromEnv_AllSet(t *testing.T) {
	t.Setenv("ALBATROSS_DB_HOST", "localhost")
	t.Setenv("ALBATROSS_DB_PORT", "5432")
	t.Setenv("ALBATROSS_DB_USER", "user")
	t.Setenv("ALBATROSS_DB_PASSWORD", "pass")
	t.Setenv("ALBATROSS_DB_NAME", "testdb")
	t.Setenv("ALBATROSS_BASE_PATH", "/app")
	t.Setenv("ALBATROSS_IS_LOCAL", "1")

	conf, err := NewConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.DBHost != "localhost" {
		t.Errorf("expected DBHost 'localhost', got %q", conf.DBHost)
	}
	if conf.DBPort != "5432" {
		t.Errorf("expected DBPort '5432', got %q", conf.DBPort)
	}
	if conf.DBUser != "user" {
		t.Errorf("expected DBUser 'user', got %q", conf.DBUser)
	}
	if conf.DBPassword != "pass" {
		t.Errorf("expected DBPassword 'pass', got %q", conf.DBPassword)
	}
	if conf.DBName != "testdb" {
		t.Errorf("expected DBName 'testdb', got %q", conf.DBName)
	}
	if conf.BasePath != "/app" {
		t.Errorf("expected BasePath '/app', got %q", conf.BasePath)
	}
	if !conf.IsLocal {
		t.Error("expected IsLocal true")
	}
}

func TestNewConfigFromEnv_IsLocalFalse(t *testing.T) {
	t.Setenv("ALBATROSS_DB_HOST", "localhost")
	t.Setenv("ALBATROSS_DB_PORT", "5432")
	t.Setenv("ALBATROSS_DB_USER", "user")
	t.Setenv("ALBATROSS_DB_PASSWORD", "pass")
	t.Setenv("ALBATROSS_DB_NAME", "testdb")
	t.Setenv("ALBATROSS_BASE_PATH", "/app")
	// ALBATROSS_IS_LOCAL not set

	conf, err := NewConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.IsLocal {
		t.Error("expected IsLocal false when env not set")
	}
}

func TestNewConfigFromEnv_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "missing DB_HOST",
			envVars: map[string]string{
				"ALBATROSS_DB_PORT":     "5432",
				"ALBATROSS_DB_USER":     "user",
				"ALBATROSS_DB_PASSWORD": "pass",
				"ALBATROSS_DB_NAME":     "testdb",
				"ALBATROSS_BASE_PATH":   "/app",
			},
		},
		{
			name: "missing DB_PORT",
			envVars: map[string]string{
				"ALBATROSS_DB_HOST":     "localhost",
				"ALBATROSS_DB_USER":     "user",
				"ALBATROSS_DB_PASSWORD": "pass",
				"ALBATROSS_DB_NAME":     "testdb",
				"ALBATROSS_BASE_PATH":   "/app",
			},
		},
		{
			name: "missing DB_USER",
			envVars: map[string]string{
				"ALBATROSS_DB_HOST":     "localhost",
				"ALBATROSS_DB_PORT":     "5432",
				"ALBATROSS_DB_PASSWORD": "pass",
				"ALBATROSS_DB_NAME":     "testdb",
				"ALBATROSS_BASE_PATH":   "/app",
			},
		},
		{
			name: "missing DB_PASSWORD",
			envVars: map[string]string{
				"ALBATROSS_DB_HOST":   "localhost",
				"ALBATROSS_DB_PORT":   "5432",
				"ALBATROSS_DB_USER":   "user",
				"ALBATROSS_DB_NAME":   "testdb",
				"ALBATROSS_BASE_PATH": "/app",
			},
		},
		{
			name: "missing DB_NAME",
			envVars: map[string]string{
				"ALBATROSS_DB_HOST":     "localhost",
				"ALBATROSS_DB_PORT":     "5432",
				"ALBATROSS_DB_USER":     "user",
				"ALBATROSS_DB_PASSWORD": "pass",
				"ALBATROSS_BASE_PATH":   "/app",
			},
		},
		{
			name: "missing BASE_PATH",
			envVars: map[string]string{
				"ALBATROSS_DB_HOST":     "localhost",
				"ALBATROSS_DB_PORT":     "5432",
				"ALBATROSS_DB_USER":     "user",
				"ALBATROSS_DB_PASSWORD": "pass",
				"ALBATROSS_DB_NAME":     "testdb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			_, err := NewConfigFromEnv()
			if err == nil {
				t.Error("expected error for missing env var, got nil")
			}
		})
	}
}
