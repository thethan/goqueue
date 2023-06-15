package zset

import (
	"bytes"
	"context"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"io"
)

func (z *ZSetQueue) RemoveFilterMiddleWare(key string) executers.FilterMiddleware {
	return func(next executers.ExecFunc) executers.ExecFunc {
		return func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
			defer func() {
				close(errChan)
			}()

			logs.Debug(ctx, "remove from retry")

			val, err := job.GetValue("retry_count")
			if err != nil {
				errChan <- err
				return
			}

			// conditionals

			if float64(val.Int()) > float64(2) {
				// remove from retry queue
				newErrorChan := make(chan error)
				stdOut := bytes.NewBuffer([]byte{})
				stdErr := bytes.NewBuffer([]byte{})
				go z.ZRangeRemove(ctx, key)(ctx, job, stdOut, stdErr, newErrorChan)
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
