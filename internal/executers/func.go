package executers

import (
	"context"
	"github.com/thethan/goqueue/internal/job"
	"io"
)

type ExecFunc func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error)

type FilterMiddleware func(next ExecFunc) ExecFunc

func (mw FilterMiddleware) Middleware(execFunc ExecFunc) ExecFunc {
	return mw(execFunc)
}
