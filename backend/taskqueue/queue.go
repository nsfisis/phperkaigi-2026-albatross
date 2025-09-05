package taskqueue

import (
	"github.com/hibiken/asynq"
)

type Queue struct {
	client *asynq.Client
}

func NewQueue(redisAddr string) *Queue {
	return &Queue{
		client: asynq.NewClient(asynq.RedisClientOpt{
			Addr: redisAddr,
		}),
	}
}

func (q *Queue) Close() {
	q.client.Close()
}

func (q *Queue) EnqueueTaskRunTestcase(
	gameID int,
	userID int,
	submissionID int,
	testcaseID int,
	language string,
	code string,
	stdin string,
	stdout string,
) error {
	task, err := newTaskRunTestcase(
		gameID,
		userID,
		submissionID,
		testcaseID,
		language,
		code,
		stdin,
		stdout,
	)
	if err != nil {
		return err
	}
	_, err = q.client.Enqueue(task)
	return err
}
