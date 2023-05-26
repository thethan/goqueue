package handlers

import (
	"bytes"
	"context"
	"github.com/thethan/goqueue/internal/job"
)

type BeforeRun func(ctx context.Context, j *job.Job)
type AfterRun func(ctx context.Context, j *job.Job, err error)
type Runner struct {
	stdOut bytes.Buffer
	stdErr bytes.Buffer

	before []func(ctx context.Context, j job.Job) (job.Job, error)
	after  []func(ctx context.Context, j job.Job, err error) (job.Job, error)
}

func (r *Runner) Run(ctx context.Context, j job.Job) error {
	var err error
	for idx := range r.before {
		j, err = r.before[idx](ctx, j)
	}

	//stdOut := bytes.Buffer{}
	//stdErr := bytes.Buffer{}

	err = j.Execute(ctx)

	for idx := range r.after {
		j, err = r.after[idx](ctx, j, err)
	}

	return err
}
