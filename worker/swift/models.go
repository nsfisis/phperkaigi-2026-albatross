package main

import (
	"errors"
	"time"
)

const (
	resultSuccess       = "success"
	resultCompileError  = "compile_error"
	resultRuntimeError  = "runtime_error"
	resultTimeout       = "timeout"
	resultInternalError = "internal_error"
)

var (
	errInvalidMaxDuration = errors.New("'max_duration_ms' must be positive")
)

type execRequestData struct {
	Code                    string `json:"code"`
	CodeHash                string `json:"code_hash"`
	Stdin                   string `json:"stdin"`
	MaxDurationMilliseconds int    `json:"max_duration_ms"`
}

func (req *execRequestData) maxDuration() time.Duration {
	return time.Duration(req.MaxDurationMilliseconds) * time.Millisecond
}

func (req *execRequestData) validate() error {
	if req.MaxDurationMilliseconds <= 0 {
		return errInvalidMaxDuration
	}
	return nil
}

type execResponseData struct {
	Status string `json:"status"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (res *execResponseData) success() bool {
	return res.Status == resultSuccess
}
