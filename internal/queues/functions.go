package queues

import (
	"context"
	"github.com/thethan/goqueue/internal/job"
	"io"
)

type Queue interface {
	GetQueue
	PushQueue
	RemoveQueue
}

type GetQueue interface {
	GetItems(ctx context.Context, jobChan chan<- job.Job) error
}
type PushQueue interface {
	PushItems(ctx context.Context, job job.Job, stdOut, stdErr io.ReadWriter, error chan error)
}

type RemoveQueue interface {
	RemoveItems(ctx context.Context, job job.Job, stdOut, stdErr io.ReadWriter, error chan error)
}

type GetItems func(ctx context.Context, jobChan chan<- job.Job) error
type PushItems func(ctx context.Context, job job.Job, stdOut, stdErr io.ReadWriter, error chan error)
type RemoveItem func(ctx context.Context, job job.Job, stdOut, stdErr io.ReadWriter, error chan error)
