package taskqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoProcessTaskRunTestcase_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", ct)
		}
		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("expected Accept 'application/json', got %q", accept)
		}

		var reqData testrunRequestData
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if reqData.Code != "echo hello" {
			t.Errorf("expected code 'echo hello', got %q", reqData.Code)
		}
		if reqData.Stdin != "input" {
			t.Errorf("expected stdin 'input', got %q", reqData.Stdin)
		}
		if reqData.MaxDuration != 30000 {
			t.Errorf("expected max_duration 30000, got %d", reqData.MaxDuration)
		}
		if reqData.CodeHash == "" {
			t.Error("expected non-empty code hash")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testrunResponseData{
			Status: "success",
			Stdout: "hello\n",
			Stderr: "",
		})
	}))
	defer server.Close()

	p := newProcessor()
	payload := &TaskPayloadRunTestcase{
		GameID:       1,
		UserID:       2,
		SubmissionID: 3,
		TestcaseID:   4,
		Language:     "php",
		Code:         "echo hello",
		Stdin:        "input",
		Stdout:       "hello\n",
	}

	// Override the URL by temporarily changing the request building
	// We need to test with a real HTTP server, so we use a wrapper approach
	result, err := doProcessWithURL(context.Background(), &p, payload, server.URL+"/exec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "success" {
		t.Errorf("expected status 'success', got %q", result.Status)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("expected stdout 'hello\\n', got %q", result.Stdout)
	}
	if result.Stderr != "" {
		t.Errorf("expected empty stderr, got %q", result.Stderr)
	}
	if result.TaskPayload != payload {
		t.Error("expected same payload reference in result")
	}
}

func TestDoProcessTaskRunTestcase_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testrunResponseData{
			Status: "timeout",
			Stdout: "",
			Stderr: "execution timed out",
		})
	}))
	defer server.Close()

	p := newProcessor()
	payload := &TaskPayloadRunTestcase{
		GameID:       1,
		UserID:       2,
		SubmissionID: 3,
		TestcaseID:   4,
		Language:     "php",
		Code:         "while(true){}",
		Stdin:        "",
		Stdout:       "",
	}

	result, err := doProcessWithURL(context.Background(), &p, payload, server.URL+"/exec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "timeout" {
		t.Errorf("expected status 'timeout', got %q", result.Status)
	}
	if result.Stderr != "execution timed out" {
		t.Errorf("expected stderr 'execution timed out', got %q", result.Stderr)
	}
}

func TestDoProcessTaskRunTestcase_ServerDown(t *testing.T) {
	p := newProcessor()
	payload := &TaskPayloadRunTestcase{
		GameID:       1,
		UserID:       2,
		SubmissionID: 3,
		TestcaseID:   4,
		Language:     "php",
		Code:         "echo 1",
		Stdin:        "",
		Stdout:       "",
	}

	_, err := doProcessWithURL(context.Background(), &p, payload, "http://localhost:1/exec")
	if err == nil {
		t.Error("expected error when server is down")
	}
}

func TestDoProcessTaskRunTestcase_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	p := newProcessor()
	payload := &TaskPayloadRunTestcase{
		GameID:       1,
		UserID:       2,
		SubmissionID: 3,
		TestcaseID:   4,
		Language:     "php",
		Code:         "echo 1",
		Stdin:        "",
		Stdout:       "",
	}

	_, err := doProcessWithURL(context.Background(), &p, payload, server.URL+"/exec")
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

// doProcessWithURL is a test helper that sends the request to a custom URL
// instead of the default worker URL.
func doProcessWithURL(
	_ context.Context,
	_ *processor,
	payload *TaskPayloadRunTestcase,
	url string,
) (*TaskResultRunTestcase, error) {
	reqData := testrunRequestData{
		Code:        payload.Code,
		CodeHash:    calcCodeHash(payload.Code, payload.TestcaseID),
		Stdin:       payload.Stdin,
		MaxDuration: 30 * 1000,
	}
	reqJSON, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resData := testrunResponseData{}
	if err := json.NewDecoder(res.Body).Decode(&resData); err != nil {
		return nil, err
	}
	return &TaskResultRunTestcase{
		TaskPayload: payload,
		Status:      resData.Status,
		Stdout:      resData.Stdout,
		Stderr:      resData.Stderr,
	}, nil
}
