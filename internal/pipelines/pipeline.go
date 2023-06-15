package pipelines

import (
	"bytes"
	"context"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"github.com/thethan/goqueue/internal/queues"
)

type ProcessPipeline interface {
	Start(ctx context.Context) error
}

func NewPipeline(getItems queues.GetQueue, execFunc executers.ExecFunc, decisionTrees ...*DecisionTree) ProcessPipeline {
	// wrap the exec function in middleware
	for idx := len(decisionTrees) - 1; idx >= 0; idx-- {
		execFunc = decisionTrees[idx].Middleware()(execFunc)
	}

	return &pipeline{
		getItems:     getItems,
		execFunc:     execFunc,
		decisionTree: decisionTrees,
	}
}

type pipeline struct {
	getItems     queues.GetQueue
	execFunc     executers.ExecFunc
	decisionTree []*DecisionTree
}

func (p *pipeline) Start(ctx context.Context) error {
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

			go func() {
				p.execFunc(ctx, jb, stdOut, stdErr, newErrChan)
			}()
			for err := range newErrChan {
				logs.Error(ctx, "error in executing from pipeline", logs.WithError(err))
			}

		default:
			// noop
		}
	}
}
