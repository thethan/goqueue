package pipelines

import (
	"bytes"
	"context"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"github.com/thethan/goqueue/internal/queues"
)

func NewPipeline(getItems queues.GetQueue, execFunc executers.ExecFunc, decisionTrees ...*DecisionTree) *Pipeline {
	// wrap the exec function in middleware
	for idx := range decisionTrees {
		execFunc = decisionTrees[idx].Middleware()(execFunc)
	}

	return &Pipeline{
		getItems:     getItems,
		execFunc:     execFunc,
		decisionTree: decisionTrees,
	}
}

type Pipeline struct {
	getItems     queues.GetQueue
	execFunc     executers.ExecFunc
	decisionTree []*DecisionTree
}

func (p *Pipeline) Start(ctx context.Context) error {
	jobChan := make(chan job.Job)
	errorChan := make(chan error)

	go func() {
		err := p.getItems.GetItems(ctx, jobChan)
		if err != nil {
			errorChan <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errorChan:
			logs.Error(ctx, "error in pipeline", logs.WithError(err))
		case jb := <-jobChan:
			newErrChan := make(chan error)
			stdOut := bytes.NewBuffer([]byte{})
			stdErr := bytes.NewBuffer([]byte{})

			p.execFunc(ctx, jb, stdOut, stdErr, newErrChan)
			for err := range newErrChan {
				logs.Error(ctx, "error in executing from pipeline", logs.WithError(err))
			}

		default:
			// noop
		}
	}
}
