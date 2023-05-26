package zset

import (
	"context"
	"fmt"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	goqueueRedis "github.com/thethan/goqueue/internal/redis"
	"io"
)

func (r *ZSetQueue) RemoveFilterMiddleWare(key string) executers.FilterMiddleware {
	return func(next executers.ExecFunc) executers.ExecFunc {
		return func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
			defer func() {
				close(errChan)
			}()

			logs.Debug(ctx, "remove from retry")

			sideKiqJob, ok := job.(*goqueueRedis.RedisJob)
			if !ok {
				errChan <- fmt.Errorf("could not convert job to sidekiq job")
				return
			}

			// conditionals

			if sideKiqJob.Retries() > 2 {
				// remove from retry queue
				newErrorChan := make(chan error)

				err := r.ZRangeRemove(ctx, newErrorChan)(ctx, key, job)
				if err != nil {
					errChan <- fmt.Errorf("could not remove from retry queue")
				}

				for err = range newErrorChan {
					errChan <- err
				}
				return
			}

			nextErrChan := make(chan error)
			go func() {
				next(ctx, job, stdOut, stdErr, nextErrChan)
			}()

			for err := range nextErrChan {
				errChan <- err
			}
		}

	}
}
