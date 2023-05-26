package handlers

import (
	"bytes"
	"context"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"github.com/thethan/goqueue/internal/queues"
)

type Pipeline struct {
	getItems queues.GetItems
	execFunc executers.ExecFunc
}

func NewPipeline(items queues.GetItems, execFunc executers.ExecFunc) *Pipeline {
	return &Pipeline{
		getItems: items,
		execFunc: execFunc,
	}
}

func (p *Pipeline) Start(ctx context.Context) error {
	errorChan := make(chan error)
	jobChan := make(chan job.Job)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		err := p.getItems(ctx, "test", jobChan)
		if err != nil {
			errorChan <- err
			cancel()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errorChan:
			logs.Error(ctx, "retrieved an error in the pipeline", logs.WithError(err))
			return err
		case j := <-jobChan:
			stdOut := bytes.NewBuffer([]byte{})
			stdErr := bytes.NewBuffer([]byte{})

			p.execFunc(ctx, j, stdOut, stdErr, errorChan)
		}
	}

}
