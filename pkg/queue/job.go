package queue

import (
	"context"
	"go.uber.org/zap/buffer"
)

type Job struct {
	Retry      bool     `json:"retry"`
	Queue      string   `json:"queue"`
	Class      string   `json:"class"`
	Args       []string `json:"args"`
	Jid        string   `json:"jid"`
	CreatedAt  float64  `json:"created_at"`
	EnqueuedAt float64  `json:"enqueued_at"`
}

func (j *Job) Run(ctx context.Context, stdIn buffer.Buffer, stdOut buffer.Buffer) error {

	return nil
}
