package executers

import (
	"bytes"
	"context"
	"github.com/thethan/goqueue/internal/job"
	"io"
	"strings"
)

type RetryExecuter struct {
}

func (r *RetryExecuter) ExecFunc(next ExecFunc) FilterMiddleware {
	return func(next ExecFunc) ExecFunc {
		return func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {

			// check if the job is in the retry queue
			bytesOut := make([]byte, 0, 1024)
			bytesIn := make([]byte, 0, 1024)

			newStdErr := bytes.NewBuffer(bytesOut)
			newStdOut := bytes.NewBuffer(bytesIn)

			next(ctx, job, newStdOut, newStdErr, errChan)

			_, _ = stdOut.Write(newStdOut.Bytes())
			_, _ = stdErr.Write(newStdErr.Bytes())

			if len(newStdErr.Bytes()) > 0 {
				strings.Contains(newStdErr.String(), "asdfjiajsfijaifjai")
				// todo delete z function
			}
		}
	}
}
