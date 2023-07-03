package pipelines

import (
	"bytes"
	"context"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"github.com/thethan/goqueue/internal/queues"
	"go.opentelemetry.io/otel/attribute"
	metric2 "go.opentelemetry.io/otel/metric"
)

type ProcessPipeline interface {
	Start(ctx context.Context) error
}

func NewPipeline(name string, meter metric2.Meter, getItems queues.GetQueue, execFunc executers.ExecFunc, decisionTrees ...*DecisionTree) ProcessPipeline {
	// wrap the exec function in middleware
	for idx := len(decisionTrees) - 1; idx >= 0; idx-- {
		execFunc = decisionTrees[idx].Middleware(meter)(execFunc)
	}

	return &pipeline{
		meter:        meter,
		name:         name,
		getItems:     getItems,
		execFunc:     execFunc,
		decisionTree: decisionTrees,
	}
}

type pipeline struct {
	name         string
	getItems     queues.GetQueue
	execFunc     executers.ExecFunc
	decisionTree []*DecisionTree

	meter metric2.Meter
}

func (p *pipeline) Start(ctx context.Context) error {
	jobChan := make(chan job.Job)
	errorChan := make(chan error)

	counter, err := p.meter.Int64Counter("job_processed", metric2.WithDescription("a gauge to determine the amount of jobs processed"))
	if err != nil {
		logs.Fatal(ctx, "could not initialize counter")
	}
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

			success := true
			for err := range newErrChan {
				logs.Error(ctx, "error in executing from pipeline", logs.WithError(err))
				success = false
			}

			opt := metric2.WithAttributes(
				attribute.Key("pipeline").String(p.name),
				attribute.Key("pipeline").String(p.name),
				attribute.Key("success").Bool(success),
			)

			counter.Add(ctx, 1, opt)
		default:
			// noop
		}
	}
}
