package taskqueue

import (
	"encoding/json"
	"testing"
)

func TestNewTaskRunTestcase(t *testing.T) {
	task, err := newTaskRunTestcase(1, 2, 3, 4, "php", "<?php echo 1;", "input", "output")
	if err != nil {
		t.Fatalf("newTaskRunTestcase returned error: %v", err)
	}
	if task.Type() != string(TaskTypeRunTestcase) {
		t.Errorf("task type = %q, want %q", task.Type(), TaskTypeRunTestcase)
	}

	var payload TaskPayloadRunTestcase
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.GameID != 1 {
		t.Errorf("GameID = %d, want 1", payload.GameID)
	}
	if payload.UserID != 2 {
		t.Errorf("UserID = %d, want 2", payload.UserID)
	}
	if payload.SubmissionID != 3 {
		t.Errorf("SubmissionID = %d, want 3", payload.SubmissionID)
	}
	if payload.TestcaseID != 4 {
		t.Errorf("TestcaseID = %d, want 4", payload.TestcaseID)
	}
	if payload.Language != "php" {
		t.Errorf("Language = %q, want %q", payload.Language, "php")
	}
	if payload.Code != "<?php echo 1;" {
		t.Errorf("Code = %q, want %q", payload.Code, "<?php echo 1;")
	}
	if payload.Stdin != "input" {
		t.Errorf("Stdin = %q, want %q", payload.Stdin, "input")
	}
	if payload.Stdout != "output" {
		t.Errorf("Stdout = %q, want %q", payload.Stdout, "output")
	}
}

func TestTaskResultRunTestcase_Interface(t *testing.T) {
	result := &TaskResultRunTestcase{
		TaskPayload: &TaskPayloadRunTestcase{GameID: 42},
		Status:      "pass",
		Stdout:      "hello",
		Stderr:      "",
	}

	var _ TaskResult = result

	if result.Type() != TaskTypeRunTestcase {
		t.Errorf("Type() = %q, want %q", result.Type(), TaskTypeRunTestcase)
	}
	if result.GameID() != 42 {
		t.Errorf("GameID() = %d, want 42", result.GameID())
	}
}
