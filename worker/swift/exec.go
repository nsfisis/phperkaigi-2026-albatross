package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	dataRootDir = "/app/data"

	wasmMaxMemorySize = 10 * 1024 * 1024 // 10 MiB
)

func prepareDirectories() error {
	if err := os.MkdirAll(dataRootDir, 0755); err != nil {
		return err
	}
	return nil
}

func execCommandWithTimeout(
	ctx context.Context,
	workingDir string,
	maxDuration time.Duration,
	makeCmd func(context.Context) *exec.Cmd,
) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, maxDuration)
	defer cancel()

	cmd := makeCmd(ctx)
	cmd.Dir = workingDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCh := make(chan error)
	go func() {
		exitCh <- cmd.Run()
	}()

	select {
	case <-ctx.Done():
		return stdout.String(), stderr.String(), ctx.Err()
	case err := <-exitCh:
		return stdout.String(), stderr.String(), err
	}
}

func convertCommandErrorToResultType(err error, defaultErrorStatus string) string {
	if err != nil {
		if err == context.DeadlineExceeded {
			return resultTimeout
		}
		return defaultErrorStatus
	}
	return resultSuccess
}

func prepareWorkingDir(workingDir string) execResponseData {
	err := os.MkdirAll(workingDir, 0755)
	if err != nil {
		return execResponseData{
			Status: resultInternalError,
			Stdout: "",
			Stderr: "Failed to create project directory",
		}
	}
	return execResponseData{
		Status: resultSuccess,
		Stdout: "",
		Stderr: "",
	}
}

func removeWorkingDir(workingDir string) {
	err := os.RemoveAll(workingDir)
	_ = err
}

func initSwiftProject(
	ctx context.Context,
	workingDir string,
	maxDuration time.Duration,
) execResponseData {
	stdout, stderr, err := execCommandWithTimeout(
		ctx,
		workingDir,
		maxDuration,
		func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(
				ctx,
				"swift",
				"package",
				"init",
				"--type", "executable",
			)
		},
	)
	return execResponseData{
		Status: convertCommandErrorToResultType(err, resultInternalError),
		Stdout: stdout,
		Stderr: stderr,
	}
}

func putSwiftSourceFile(
	workingDir string,
	code string,
) execResponseData {
	err := os.WriteFile(workingDir+"/Sources/main.swift", []byte(code), 0644)

	if err != nil {
		return execResponseData{
			Status: convertCommandErrorToResultType(err, resultInternalError),
			Stdout: "",
			Stderr: "Failed to copy source file",
		}
	}
	return execResponseData{
		Status: resultSuccess,
		Stdout: "",
		Stderr: "",
	}
}

func buildSwiftProject(
	ctx context.Context,
	workingDir string,
	maxDuration time.Duration,
) execResponseData {
	stdout, stderr, err := execCommandWithTimeout(
		ctx,
		workingDir,
		maxDuration,
		func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(
				ctx,
				"swift",
				"build",
				"--swift-sdk", "wasm32-unknown-wasi",
			)
		},
	)
	return execResponseData{
		Status: convertCommandErrorToResultType(err, resultCompileError),
		Stdout: stdout,
		Stderr: stderr,
	}
}

func runWasm(
	ctx context.Context,
	workingDir string,
	codeHash string,
	stdin string,
	maxDuration time.Duration,
) execResponseData {
	stdout, stderr, err := execCommandWithTimeout(
		ctx,
		workingDir,
		maxDuration,
		func(ctx context.Context) *exec.Cmd {
			cmd := exec.CommandContext(
				ctx,
				"wasmtime",
				"-W", fmt.Sprintf("max-memory-size=%d", wasmMaxMemorySize),
				".build/wasm32-unknown-wasi/debug/"+codeHash+".wasm",
			)
			cmd.Stdin = strings.NewReader(stdin)
			return cmd
		},
	)
	return execResponseData{
		Status: convertCommandErrorToResultType(err, resultRuntimeError),
		Stdout: stdout,
		Stderr: stderr,
	}
}

func doExec(
	ctx context.Context,
	code string,
	codeHash string,
	stdin string,
	maxDuration time.Duration,
) execResponseData {
	workingDir := dataRootDir + "/" + codeHash

	res := prepareWorkingDir(workingDir)
	if !res.success() {
		return res
	}
	defer removeWorkingDir(workingDir)

	res = initSwiftProject(ctx, workingDir, maxDuration)
	if !res.success() {
		return res
	}

	res = putSwiftSourceFile(workingDir, code)
	if !res.success() {
		return res
	}

	res = buildSwiftProject(ctx, workingDir, maxDuration)
	if !res.success() {
		return res
	}

	res = runWasm(ctx, workingDir, codeHash, stdin, maxDuration)
	if !res.success() {
		return res
	}

	return res
}
