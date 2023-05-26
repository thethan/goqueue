package conditionals

import (
	"context"
	"github.com/thethan/goqueue/internal/job"
)

type Operator string

const (
	GreaterThan        Operator = ">"
	GreaterThanEqualTo Operator = ">="
	LessThan           Operator = "<"
	LessThanEqualTo    Operator = "<="
	Equal              Operator = "=="

	Contains Operator = "contains"
)

type ConditionFunc func(ctx context.Context, job job.Job) bool
