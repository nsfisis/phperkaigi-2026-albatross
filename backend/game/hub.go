package game

import (
	"context"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"albatross-2026-backend/db"
	"albatross-2026-backend/taskqueue"
)

type TaskQueueInterface interface {
	EnqueueTaskRunTestcase(gameID, userID, submissionID, testcaseID int, language, code, stdin, stdout string) error
}

type TaskWorkerInterface interface {
	Run() error
	Results() chan taskqueue.TaskResult
}

var whitespaceRe = regexp.MustCompile(`\s+`)

type Hub struct {
	q          db.Querier
	txm        db.TxManager
	ctx        context.Context
	taskQueue  TaskQueueInterface
	taskWorker TaskWorkerInterface
}

func NewGameHub(q db.Querier, txm db.TxManager, taskQueue TaskQueueInterface, taskWorker TaskWorkerInterface) *Hub {
	return &Hub{
		q:          q,
		txm:        txm,
		ctx:        context.Background(),
		taskQueue:  taskQueue,
		taskWorker: taskWorker,
	}
}

func (hub *Hub) Run() {
	go func() {
		if err := hub.taskWorker.Run(); err != nil {
			slog.Error("task worker failed", "error", err)
			os.Exit(1)
		}
	}()

	go hub.processTaskResults()
}

func (hub *Hub) CalcCodeSize(code string, language string) int {
	trimmed := whitespaceRe.ReplaceAllString(code, "")
	if language == "php" {
		return len(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(trimmed, "<?php"), "<?"), "?>"))
	}
	return len(trimmed)
}

func (hub *Hub) EnqueueTestTasks(ctx context.Context, submissionID, gameID, userID int, language, code string) error {
	rows, err := hub.q.ListTestcasesByGameID(ctx, int32(gameID))
	if err != nil {
		return err
	}
	for _, row := range rows {
		err := hub.taskQueue.EnqueueTaskRunTestcase(
			gameID,
			userID,
			submissionID,
			int(row.TestcaseID),
			language,
			code,
			row.Stdin,
			row.Stdout,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (hub *Hub) processTaskResults() {
	for taskResult := range hub.taskWorker.Results() {
		switch taskResult := taskResult.(type) {
		case *taskqueue.TaskResultRunTestcase:
			if err := hub.processTaskResultRunTestcase(taskResult); err != nil {
				slog.Error("failed to process testcase result", "error", err, "submissionID", taskResult.TaskPayload.SubmissionID)
				continue
			}
			aggregatedStatus, err := hub.q.AggregateTestcaseResults(hub.ctx, int32(taskResult.TaskPayload.SubmissionID))
			if err != nil {
				slog.Error("failed to aggregate testcase results", "error", err, "submissionID", taskResult.TaskPayload.SubmissionID)
				continue
			}
			if aggregatedStatus == "running" {
				continue
			}

			if err := hub.updateSubmissionAndGameState(taskResult, aggregatedStatus); err != nil {
				slog.Error("failed to update submission and game state", "error", err, "submissionID", taskResult.TaskPayload.SubmissionID)
				continue
			}
		default:
			slog.Error("unexpected task result type", "type", taskResult.Type())
			continue
		}
	}
}

func (hub *Hub) updateSubmissionAndGameState(taskResult *taskqueue.TaskResultRunTestcase, aggregatedStatus string) error {
	return hub.txm.RunInTx(hub.ctx, func(qtx db.Querier) error {
		if err := qtx.UpdateSubmissionStatus(hub.ctx, db.UpdateSubmissionStatusParams{
			SubmissionID: int32(taskResult.TaskPayload.SubmissionID),
			Status:       aggregatedStatus,
		}); err != nil {
			return err
		}
		if err := qtx.UpdateGameStateStatus(hub.ctx, db.UpdateGameStateStatusParams{
			GameID: int32(taskResult.TaskPayload.GameID),
			UserID: int32(taskResult.TaskPayload.UserID),
			Status: aggregatedStatus,
		}); err != nil {
			return err
		}
		if aggregatedStatus == "success" {
			if err := qtx.SyncGameStateBestScoreSubmission(hub.ctx, db.SyncGameStateBestScoreSubmissionParams{
				GameID: int32(taskResult.TaskPayload.GameID),
				UserID: int32(taskResult.TaskPayload.UserID),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (hub *Hub) processTaskResultRunTestcase(
	taskResult *taskqueue.TaskResultRunTestcase,
) error {
	if taskResult.Err != nil {
		return taskResult.Err
	}

	if taskResult.Status != "success" {
		if err := hub.q.CreateTestcaseResult(hub.ctx, db.CreateTestcaseResultParams{
			SubmissionID: int32(taskResult.TaskPayload.SubmissionID),
			TestcaseID:   int32(taskResult.TaskPayload.TestcaseID),
			Status:       taskResult.Status,
			Stdout:       taskResult.Stdout,
			Stderr:       taskResult.Stderr,
		}); err != nil {
			return err
		}
		return nil
	}

	var status string
	if isTestcaseResultCorrect(taskResult.TaskPayload.Stdout, taskResult.Stdout) {
		status = "success"
	} else {
		status = "wrong_answer"
	}
	if err := hub.q.CreateTestcaseResult(hub.ctx, db.CreateTestcaseResultParams{
		SubmissionID: int32(taskResult.TaskPayload.SubmissionID),
		TestcaseID:   int32(taskResult.TaskPayload.TestcaseID),
		Status:       status,
		Stdout:       taskResult.Stdout,
		Stderr:       taskResult.Stderr,
	}); err != nil {
		return err
	}
	return nil
}

func normalizeTestcaseResultOutput(s string) string {
	re := regexp.MustCompile(`\r\n|\r`)
	return re.ReplaceAllString(strings.TrimSpace(s), "\n")
}

func isTestcaseResultCorrect(expectedStdout, actualStdout string) bool {
	expectedStdout = normalizeTestcaseResultOutput(expectedStdout)
	actualStdout = normalizeTestcaseResultOutput(actualStdout)
	return actualStdout == expectedStdout
}
