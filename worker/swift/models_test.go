package main

import (
	"testing"
	"time"
)

func TestExecRequestData_Validate(t *testing.T) {
	tests := []struct {
		name          string
		maxDurationMs int
		wantErr       error
	}{
		{"positive value", 1000, nil},
		{"zero", 0, errInvalidMaxDuration},
		{"negative value", -1, errInvalidMaxDuration},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &execRequestData{MaxDurationMilliseconds: tt.maxDurationMs}
			err := req.validate()
			if err != tt.wantErr {
				t.Errorf("validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecRequestData_MaxDuration(t *testing.T) {
	tests := []struct {
		name          string
		maxDurationMs int
		want          time.Duration
	}{
		{"1000ms", 1000, 1 * time.Second},
		{"500ms", 500, 500 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &execRequestData{MaxDurationMilliseconds: tt.maxDurationMs}
			got := req.maxDuration()
			if got != tt.want {
				t.Errorf("maxDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecResponseData_Success(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"success", resultSuccess, true},
		{"compile_error", resultCompileError, false},
		{"runtime_error", resultRuntimeError, false},
		{"timeout", resultTimeout, false},
		{"internal_error", resultInternalError, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &execResponseData{Status: tt.status}
			got := res.success()
			if got != tt.want {
				t.Errorf("success() = %v, want %v", got, tt.want)
			}
		})
	}
}
