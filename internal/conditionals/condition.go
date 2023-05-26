package conditionals

import (
	"context"
	"github.com/thethan/goqueue/internal/job"
	"reflect"
	"strings"
)

type condition struct {
	operator   Operator
	element    string
	comparison interface{}
}

func NewCondition(element string, operator Operator, comparison interface{}) condition {
	return condition{element: element, operator: operator, comparison: comparison}
}

func (c condition) Evaluate(ctx context.Context, job job.Job) bool {
	//t := reflect.TypeOf(job)

	r := reflect.ValueOf(job)
	f := reflect.Indirect(r).FieldByName(c.element)

	if f.Kind() == reflect.Invalid {
		return c.getMethod(job)
	}

	// this is assuming that the field is an attribute of the job
	if f.Kind() == reflect.String {
		return c.stringEval(f)
	}

	return false
}
func (c condition) getMethod(job job.Job) bool {
	f := reflect.ValueOf(job).MethodByName(c.element)

	vals := f.Call([]reflect.Value{})
	if len(vals) == 0 {
		return false
	}

	return c.eval(vals[0])
}

func (c condition) eval(v reflect.Value) bool {
	if c.operator == Contains {
		if v.Kind() == reflect.String {
			return c.stringEval(v)
		}
	}

	return c.intEval(v)
}

func (c condition) stringEval(v reflect.Value) bool {
	comp := reflect.ValueOf(c.comparison)

	switch c.operator {
	case Contains:
		mainString := v.String()
		subString := comp.String()
		return strings.Contains(mainString, subString)
	}

	return false
}
func (c condition) intEval(v reflect.Value) bool {
	comp := reflect.ValueOf(c.comparison)

	switch c.operator {
	case GreaterThan:
		return v.Int() > comp.Int()
	}

	return false
}
