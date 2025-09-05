package taskqueue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

type TaskType string

const (
	TaskTypeRunTestcase TaskType = "run_testcase"
)

type TaskPayloadRunTestcase struct {
	GameID       int
	UserID       int
	SubmissionID int
	TestcaseID   int
	Language     string
	Code         string
	Stdin        string
	Stdout       string
}

func newTaskRunTestcase(
	gameID int,
	userID int,
	submissionID int,
	testcaseID int,
	language string,
	code string,
	stdin string,
	stdout string,
) (*asynq.Task, error) {
	payload, err := json.Marshal(TaskPayloadRunTestcase{
		GameID:       gameID,
		UserID:       userID,
		SubmissionID: submissionID,
		TestcaseID:   testcaseID,
		Language:     language,
		Code:         code,
		Stdin:        stdin,
		Stdout:       stdout,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(
		string(TaskTypeRunTestcase),
		payload,
		asynq.MaxRetry(3),
	), nil
}

type TaskResult interface {
	Type() TaskType
	GameID() int
}

type TaskResultRunTestcase struct {
	TaskPayload *TaskPayloadRunTestcase
	Status      string
	Stdout      string
	Stderr      string
	Err         error
}

func (r *TaskResultRunTestcase) Type() TaskType { return TaskTypeRunTestcase }
func (r *TaskResultRunTestcase) GameID() int    { return r.TaskPayload.GameID }
