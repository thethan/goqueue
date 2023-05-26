package queues

import (
	"context"
	"github.com/thethan/goqueue/internal/job"
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
	PushItems(ctx context.Context, job job.Job) error
}

type RemoveQueue interface {
	RemoveItems(ctx context.Context, job job.Job) error
}

type GetItems func(ctx context.Context, key string, jobChan chan<- job.Job) error
type PushItems func(ctx context.Context, key string, job job.Job) error
type RemoveItem func(ctx context.Context, key string, job job.Job) error
