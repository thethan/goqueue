package pipelines

import (
	"context"
	"github.com/thethan/goqueue/internal/conditionals"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"io"
)

type conditionReturnFunc executers.ExecFunc
type DecisionTree struct {
	condition conditionals.ConditionFunc

	trueMethod executers.ExecFunc

	falseMethod   executers.ExecFunc
	returnOnFalse bool
	returnOnTrue  bool
}

func NewConditionTree(condition conditionals.ConditionFunc, trueFunction, falseFunc executers.ExecFunc, returnOnTrue, returnOnFalse bool) DecisionTree {
	return DecisionTree{
		condition:     condition,
		trueMethod:    trueFunction,
		falseMethod:   falseFunc,
		returnOnFalse: returnOnFalse,
		returnOnTrue:  returnOnTrue,
	}
}

func (c *DecisionTree) Middleware() executers.FilterMiddleware {
	return func(next executers.ExecFunc) executers.ExecFunc {
		return func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
			defer func() {
				close(errChan)
			}()
			// if we hit the true method we want to hit the push before continuing in the
			// pipeline
			if c.condition(ctx, job) {
				// make the true condition
				newErrorChan := make(chan error)
				go func() {
					c.trueMethod(ctx, job, stdOut, stdErr, newErrorChan)
				}()
				for err := range newErrorChan {
					errChan <- err
					logs.Debug(ctx, "conditional filter middleware false was executed")
				}
				logs.Debug(ctx, "conditional filter middleware false was executed")

				if c.returnOnTrue {
					return
				}

			} else {
				newErrorChan := make(chan error)
				go func() {
					c.falseMethod(ctx, job, stdOut, stdErr, newErrorChan)
				}()
				for err := range newErrorChan {
					errChan <- err
					logs.Debug(ctx, "conditional filter middleware true was executed")

				}

				if c.returnOnFalse {
					return
				}

			}

			logs.Debug(ctx, "will now execute")
			newErrChan := make(chan error)

			go next(ctx, job, stdOut, stdErr, newErrChan)
			for err := range newErrChan {
				errChan <- err
			}
		}
	}
}
