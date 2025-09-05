package taskqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type processor struct{}

func newProcessor() processor {
	return processor{}
}

type testrunRequestData struct {
	Code        string `json:"code"`
	Stdin       string `json:"stdin"`
	MaxDuration int    `json:"max_duration_ms"`
}

type testrunResponseData struct {
	Status string `json:"status"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (p *processor) doProcessTaskRunTestcase(
	_ context.Context,
	payload *TaskPayloadRunTestcase,
) (*TaskResultRunTestcase, error) {
	reqData := testrunRequestData{
		Code:        payload.Code,
		Stdin:       payload.Stdin,
		MaxDuration: 5000,
	}
	reqJSON, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %v", err)
	}
	req, err := http.NewRequest("POST", "http://worker-php:80/exec", bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do failed: %v", err)
	}
	defer res.Body.Close()

	resData := testrunResponseData{}
	if err := json.NewDecoder(res.Body).Decode(&resData); err != nil {
		return nil, fmt.Errorf("json.Decode failed: %v", err)
	}
	return &TaskResultRunTestcase{
		TaskPayload: payload,
		Status:      resData.Status,
		Stdout:      resData.Stdout,
		Stderr:      resData.Stderr,
	}, nil
}
