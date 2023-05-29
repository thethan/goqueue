package conditionals

import (
	"context"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"reflect"
	"strings"
)

type condition struct {
	operator   Operator
	element    string
	comparison reflect.Value
}

func NewCondition(element string, operator Operator, comparison interface{}) condition {
	compVal := reflect.ValueOf(comparison)
	switch compVal.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		compVal = reflect.ValueOf(float64(compVal.Int()))
	}

	return condition{element: element, operator: operator, comparison: compVal}
}

func (c condition) Evaluate(ctx context.Context, job job.Job) bool {
	//t := reflect.TypeOf(job)

	r, err := job.GetValue(c.element)
	if err != nil {
		logs.Warn(ctx, "error getting value", logs.WithError(err), logs.WithValue("element", c.element))
		return false
	}

	// this is assuming that the field is an attribute of the job
	return c.eval(r)
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
	switch c.operator {
	case Contains:
		mainString := v.String()
		subString := c.comparison.String()
		return strings.Contains(mainString, subString)
	}

	return false
}
func (c condition) intEval(v reflect.Value) bool {
	switch c.operator {
	case GreaterThan:
		return v.Float() > c.comparison.Float()
	}

	return false
}
