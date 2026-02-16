package game

import (
	"context"
	"errors"
	"testing"

	"albatross-2026-backend/db"
	"albatross-2026-backend/taskqueue"
)

// mockTaskQueue implements TaskQueueInterface for testing.
type mockTaskQueue struct {
	enqueued []taskqueue.TaskPayloadRunTestcase
	err      error
}

func (m *mockTaskQueue) EnqueueTaskRunTestcase(gameID, userID, submissionID, testcaseID int, language, code, stdin, stdout string) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, taskqueue.TaskPayloadRunTestcase{
		GameID:       gameID,
		UserID:       userID,
		SubmissionID: submissionID,
		TestcaseID:   testcaseID,
		Language:     language,
		Code:         code,
		Stdin:        stdin,
		Stdout:       stdout,
	})
	return nil
}

// mockQuerier implements db.Querier for testing.
type mockQuerier struct {
	db.Querier
	listTestcasesByGameIDFunc func(ctx context.Context, gameID int32) ([]db.Testcase, error)
	createTestcaseResultFunc  func(ctx context.Context, arg db.CreateTestcaseResultParams) error
	createTestcaseResultCalls []db.CreateTestcaseResultParams
}

func (m *mockQuerier) ListTestcasesByGameID(ctx context.Context, gameID int32) ([]db.Testcase, error) {
	if m.listTestcasesByGameIDFunc != nil {
		return m.listTestcasesByGameIDFunc(ctx, gameID)
	}
	return nil, nil
}

func (m *mockQuerier) CreateTestcaseResult(_ context.Context, arg db.CreateTestcaseResultParams) error {
	m.createTestcaseResultCalls = append(m.createTestcaseResultCalls, arg)
	if m.createTestcaseResultFunc != nil {
		return m.createTestcaseResultFunc(context.Background(), arg)
	}
	return nil
}

func TestEnqueueTestTasks(t *testing.T) {
	testcases := []db.Testcase{
		{TestcaseID: 1, ProblemID: 10, Stdin: "input1", Stdout: "output1"},
		{TestcaseID: 2, ProblemID: 10, Stdin: "input2", Stdout: "output2"},
	}

	tq := &mockTaskQueue{}
	mq := &mockQuerier{
		listTestcasesByGameIDFunc: func(_ context.Context, _ int32) ([]db.Testcase, error) {
			return testcases, nil
		},
	}

	hub := &Hub{q: mq, taskQueue: tq, ctx: context.Background()}

	err := hub.EnqueueTestTasks(context.Background(), 100, 1, 42, "php", "<?php echo 1;")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tq.enqueued) != 2 {
		t.Fatalf("expected 2 enqueued tasks, got %d", len(tq.enqueued))
	}
	if tq.enqueued[0].TestcaseID != 1 || tq.enqueued[1].TestcaseID != 2 {
		t.Errorf("unexpected testcase IDs: %d, %d", tq.enqueued[0].TestcaseID, tq.enqueued[1].TestcaseID)
	}
	if tq.enqueued[0].SubmissionID != 100 {
		t.Errorf("expected submission ID 100, got %d", tq.enqueued[0].SubmissionID)
	}
}

func TestEnqueueTestTasks_QueueError(t *testing.T) {
	queueErr := errors.New("queue full")
	tq := &mockTaskQueue{err: queueErr}
	mq := &mockQuerier{
		listTestcasesByGameIDFunc: func(_ context.Context, _ int32) ([]db.Testcase, error) {
			return []db.Testcase{{TestcaseID: 1, Stdin: "in", Stdout: "out"}}, nil
		},
	}

	hub := &Hub{q: mq, taskQueue: tq, ctx: context.Background()}

	err := hub.EnqueueTestTasks(context.Background(), 100, 1, 42, "php", "code")
	if !errors.Is(err, queueErr) {
		t.Errorf("expected queue error, got: %v", err)
	}
}

func TestEnqueueTestTasks_DBError(t *testing.T) {
	dbErr := errors.New("db error")
	tq := &mockTaskQueue{}
	mq := &mockQuerier{
		listTestcasesByGameIDFunc: func(_ context.Context, _ int32) ([]db.Testcase, error) {
			return nil, dbErr
		},
	}

	hub := &Hub{q: mq, taskQueue: tq, ctx: context.Background()}

	err := hub.EnqueueTestTasks(context.Background(), 100, 1, 42, "php", "code")
	if !errors.Is(err, dbErr) {
		t.Errorf("expected db error, got: %v", err)
	}
}

func TestProcessTaskResultRunTestcase_Success_Correct(t *testing.T) {
	mq := &mockQuerier{}
	hub := &Hub{q: mq, ctx: context.Background()}

	result := &taskqueue.TaskResultRunTestcase{
		TaskPayload: &taskqueue.TaskPayloadRunTestcase{
			SubmissionID: 1,
			TestcaseID:   2,
			Stdout:       "expected output",
		},
		Status: "success",
		Stdout: "expected output",
		Stderr: "",
	}

	err := hub.processTaskResultRunTestcase(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mq.createTestcaseResultCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mq.createTestcaseResultCalls))
	}
	if mq.createTestcaseResultCalls[0].Status != "success" {
		t.Errorf("expected status 'success', got %q", mq.createTestcaseResultCalls[0].Status)
	}
}

func TestProcessTaskResultRunTestcase_Success_WrongAnswer(t *testing.T) {
	mq := &mockQuerier{}
	hub := &Hub{q: mq, ctx: context.Background()}

	result := &taskqueue.TaskResultRunTestcase{
		TaskPayload: &taskqueue.TaskPayloadRunTestcase{
			SubmissionID: 1,
			TestcaseID:   2,
			Stdout:       "expected",
		},
		Status: "success",
		Stdout: "wrong",
		Stderr: "",
	}

	err := hub.processTaskResultRunTestcase(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mq.createTestcaseResultCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mq.createTestcaseResultCalls))
	}
	if mq.createTestcaseResultCalls[0].Status != "wrong_answer" {
		t.Errorf("expected status 'wrong_answer', got %q", mq.createTestcaseResultCalls[0].Status)
	}
}

func TestProcessTaskResultRunTestcase_NonSuccess(t *testing.T) {
	mq := &mockQuerier{}
	hub := &Hub{q: mq, ctx: context.Background()}

	result := &taskqueue.TaskResultRunTestcase{
		TaskPayload: &taskqueue.TaskPayloadRunTestcase{
			SubmissionID: 1,
			TestcaseID:   2,
			Stdout:       "expected",
		},
		Status: "timeout",
		Stdout: "",
		Stderr: "execution timed out",
	}

	err := hub.processTaskResultRunTestcase(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mq.createTestcaseResultCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mq.createTestcaseResultCalls))
	}
	if mq.createTestcaseResultCalls[0].Status != "timeout" {
		t.Errorf("expected status 'timeout', got %q", mq.createTestcaseResultCalls[0].Status)
	}
}

func TestProcessTaskResultRunTestcase_TaskError(t *testing.T) {
	mq := &mockQuerier{}
	hub := &Hub{q: mq, ctx: context.Background()}

	taskErr := errors.New("worker crashed")
	result := &taskqueue.TaskResultRunTestcase{
		TaskPayload: &taskqueue.TaskPayloadRunTestcase{
			SubmissionID: 1,
			TestcaseID:   2,
		},
		Err: taskErr,
	}

	err := hub.processTaskResultRunTestcase(result)
	if !errors.Is(err, taskErr) {
		t.Errorf("expected task error, got: %v", err)
	}
	if len(mq.createTestcaseResultCalls) != 0 {
		t.Error("expected no DB calls when task has error")
	}
}

func TestNormalizeTestcaseResultOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no changes needed", "hello", "hello"},
		{"trim spaces", "  hello  ", "hello"},
		{"CRLF to LF", "line1\r\nline2", "line1\nline2"},
		{"CR to LF", "line1\rline2", "line1\nline2"},
		{"mixed", "  line1\r\nline2\r  ", "line1\nline2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTestcaseResultOutput(tt.input)
			if got != tt.want {
				t.Errorf("normalizeTestcaseResultOutput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCalcCodeSize_PHP(t *testing.T) {
	hub := &Hub{}
	tests := []struct {
		name     string
		code     string
		language string
		want     int
	}{
		{
			name:     "simple php code",
			code:     "<?php echo 1;",
			language: "php",
			want:     6, // "echo1;" after stripping whitespace and "<?php"
		},
		{
			name:     "php with short open tag",
			code:     "<? echo 1;",
			language: "php",
			want:     6, // "echo1;" after stripping whitespace and "<?"
		},
		{
			name:     "php with closing tag",
			code:     "<?php echo 1; ?>",
			language: "php",
			want:     6, // "echo1;" after stripping whitespace, "<?php", and "?>"
		},
		{
			name:     "php with whitespace",
			code:     "<?php echo   1 ;  ?>",
			language: "php",
			want:     6,
		},
		{
			name:     "non-php language",
			code:     "print(1)",
			language: "swift",
			want:     8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hub.CalcCodeSize(tt.code, tt.language)
			if got != tt.want {
				t.Errorf("CalcCodeSize(%q, %q) = %d, want %d", tt.code, tt.language, got, tt.want)
			}
		})
	}
}

func TestIsTestcaseResultCorrect(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		want     bool
	}{
		{
			name:     "exact match",
			expected: "hello",
			actual:   "hello",
			want:     true,
		},
		{
			name:     "trailing newline ignored",
			expected: "hello\n",
			actual:   "hello",
			want:     true,
		},
		{
			name:     "CRLF normalized",
			expected: "hello\r\n",
			actual:   "hello\n",
			want:     true,
		},
		{
			name:     "mismatch",
			expected: "hello",
			actual:   "world",
			want:     false,
		},
		{
			name:     "multiline match",
			expected: "line1\nline2",
			actual:   "line1\nline2\n",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTestcaseResultCorrect(tt.expected, tt.actual)
			if got != tt.want {
				t.Errorf("isTestcaseResultCorrect(%q, %q) = %v, want %v", tt.expected, tt.actual, got, tt.want)
			}
		})
	}
}
