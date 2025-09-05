package game

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/nsfisis/iosdc-japan-2025-albatross/backend/db"
	"github.com/nsfisis/iosdc-japan-2025-albatross/backend/taskqueue"
)

type Hub struct {
	q          *db.Queries
	ctx        context.Context
	taskQueue  *taskqueue.Queue
	taskWorker *taskqueue.WorkerServer
}

func NewGameHub(q *db.Queries, taskQueue *taskqueue.Queue, taskWorker *taskqueue.WorkerServer) *Hub {
	return &Hub{
		q:          q,
		ctx:        context.Background(),
		taskQueue:  taskQueue,
		taskWorker: taskWorker,
	}
}

func (hub *Hub) Run() {
	go func() {
		if err := hub.taskWorker.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	go hub.processTaskResults()
}

func (hub *Hub) CalcCodeSize(code string, language string) int {
	re := regexp.MustCompile(`\s+`)
	trimmed := re.ReplaceAllString(code, "")
	if language == "php" {
		return len(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(trimmed, "<?php"), "<?"), "?>"))
	} else {
		return len(trimmed)
	}
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
			// TODO: error handling
			_ = hub.processTaskResultRunTestcase(taskResult)
			aggregatedStatus, _ := hub.q.AggregateTestcaseResults(hub.ctx, int32(taskResult.TaskPayload.SubmissionID))
			if aggregatedStatus == "running" {
				continue
			}

			// TODO: error handling
			// TODO: transaction
			_ = hub.q.UpdateSubmissionStatus(hub.ctx, db.UpdateSubmissionStatusParams{
				SubmissionID: int32(taskResult.TaskPayload.SubmissionID),
				Status:       aggregatedStatus,
			})
			_ = hub.q.UpdateGameStateStatus(hub.ctx, db.UpdateGameStateStatusParams{
				GameID: int32(taskResult.TaskPayload.GameID),
				UserID: int32(taskResult.TaskPayload.UserID),
				Status: aggregatedStatus,
			})
			if aggregatedStatus != "success" {
				continue
			}
			_ = hub.q.SyncGameStateBestScoreSubmission(hub.ctx, db.SyncGameStateBestScoreSubmissionParams{
				GameID: int32(taskResult.TaskPayload.GameID),
				UserID: int32(taskResult.TaskPayload.UserID),
			})
		default:
			panic("unexpected task result type")
		}
	}
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

func isTestcaseResultCorrect(expectedStdout, actualStdout string) bool {
	expectedStdout = strings.TrimSpace(expectedStdout)
	actualStdout = strings.TrimSpace(actualStdout)
	return actualStdout == expectedStdout
}
