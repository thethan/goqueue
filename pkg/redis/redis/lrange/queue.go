package lrange

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"io"
)

type LRangeQueue struct {
	jobbuilder *job.Builder
	key        string
	client     *redis.Client
}

func NewLRangeQueue(key string, client *redis.Client) *LRangeQueue {
	jobJuilder := job.NewBuilder(&job.Configuration{Type: "json"})

	return &LRangeQueue{key: key, client: client, jobbuilder: jobJuilder}
}

func (l *LRangeQueue) GetItems(ctx context.Context, jobChan chan<- job.Job) error {
	defer func() {
		close(jobChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			stop := int64(1000)

			start := int64(0)
			jobsString := make([]string, 0)
			err := l.client.LRange(ctx, l.key, start, stop).ScanSlice(&jobsString)
			if err != nil {
				return err
			}

			logs.Info(ctx, "got items from queue", logs.WithValue("count", len(jobsString)))

			for idx := range jobsString {
				jobStr := jobsString[idx]

				j, err := l.jobbuilder.MakeJob([]byte(jobStr))
				if err != nil {
					return fmt.Errorf("could not convert job to string")
				}

				jobChan <- j
			}
		}
	}
}

func (z *LRangeQueue) PushItems(context.Context, job.Job, io.ReadWriter, io.ReadWriter, chan error) {
	//TODO implement me
	panic("implement me")
}

func (l *LRangeQueue) RemoveItems(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
	close(errChan)
	intCmd := l.client.LRem(ctx, l.key, 1, string(job.Raw()))
	if intCmd.Err() != nil {
		errChan <- intCmd.Err()
	}

	jidVal, err := job.GetValue("jid")
	if err != nil {
		errChan <- intCmd.Err()
	}

	logs.Info(ctx, "removed from lrange queue", logs.WithValue("key", l.key), logs.WithValue("id", jidVal.String()), logs.WithValue("int", intCmd.Val()))
}
