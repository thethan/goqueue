package pipelines

import (
	"context"
	"github.com/thethan/goqueue/internal/conditionals"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"io"
)

type conditionReturnFunc func(ctx context.Context, job job.Job) error
type DecisionTree struct {
	condition conditionals.ConditionFunc
	success   Pipeline

	trueMethod conditionReturnFunc

	executeAfterConditional bool
	falseMethod             conditionReturnFunc
}

func NewConditionTree(condition conditionals.ConditionFunc, trueFunction, falseFunc conditionReturnFunc) DecisionTree {
	return DecisionTree{
		condition:               condition,
		executeAfterConditional: true,
		trueMethod:              trueFunction,
		falseMethod:             falseFunc,
	}
}

func (c DecisionTree) run(ctx context.Context, job2 job.Job) error {

	if c.condition(context.Background(), job2) {
		return c.trueMethod(ctx, job2)
	}

	return c.falseMethod(ctx, job2)
}

func (c *DecisionTree) Middleware() executers.FilterMiddleware {
	return func(next executers.ExecFunc) executers.ExecFunc {
		return func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
			// if we hit the true method we want to hit the push before continueing in the
			// pipeline
			if !c.condition(context.Background(), job) {
				err := c.falseMethod(ctx, job)
				if err != nil {
					logs.Error(ctx, "error in running conditional filter middleware failure ", logs.WithError(err))
					return
				}

				logs.Debug(ctx, "conditional filter middleware false was executed")
			} else {
				err := c.trueMethod(ctx, job)
				if err != nil {
					logs.Error(ctx, "error in running conditional filter middleware success ", logs.WithError(err))
					return
				}
			}

			logs.Debug(ctx, "will now execute")

			if c.executeAfterConditional {
				next(ctx, job, stdOut, stdErr, errChan)
			}
		}
	}
}
