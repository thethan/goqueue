package conditionals

import (
	"context"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"reflect"
	"regexp"
	"strconv"
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
	element := c.element
	if n := strings.IndexByte(c.element, '['); n >= 0 {
		element = element[:n]
	}

	r, err := job.GetValue(element)
	if err != nil {
		logs.Warn(ctx, "error getting value", logs.WithError(err), logs.WithValue("element", c.element))
		return false
	}

	// this is assuming that the field is an attribute of the job
	return c.eval(r)
}

/**
getMethod is not used
*/
//func (c condition) getMethod(job job.Job) bool {
//	f := reflect.ValueOf(job).MethodByName(c.element)
//
//	vals := f.Call([]reflect.Value{})
//	if len(vals) == 0 {
//		return false
//	}
//
//	return c.eval(vals[0])
//}

func (c condition) eval(v reflect.Value) bool {

	if v.Kind() == reflect.String {
		return c.stringEval(v)
	}

	if v.Kind() == reflect.Slice {
		return c.sliceEval(v)
	}

	return c.intEval(v)
}

func (c condition) stringEval(v reflect.Value) bool {
	mainString := v.String()
	subString := c.comparison.String()

	switch c.operator {
	case Contains:
		return strings.Contains(mainString, subString)
	case Equal:
		return mainString == subString
	}

	return false
}
func (c condition) intEval(v reflect.Value) bool {
	switch c.operator {
	case GreaterThan:
		return v.Float() > c.comparison.Float()
	case Equal:
		floatVal := v.Float()
		compVal := c.comparison.Float()
		_ = floatVal
		_ = compVal

		return v.Float() == c.comparison.Float()
	}

	return false
}

func (c condition) sliceEval(v reflect.Value) bool {

	reg, err := regexp.Compile("\\[(.*?)]")
	if err != nil {
		logs.Warn(context.Background(), "could not get regex", logs.WithError(err))
		return false
	}

	stringsts := reg.FindAllString(c.element, 1)

	if len(stringsts) == 0 {
		logs.Warn(context.Background(), "length of slice is greater than 1", logs.WithError(err))
		return false
	}

	strIdx := strings.Replace(strings.Replace(stringsts[0], "]", "", -1), "[", "", -1)
	idx, err := strconv.Atoi(strIdx)
	if err != nil {
		logs.Warn(context.Background(), "string inside of [] was not an int", logs.WithError(err))
		return false
	}

	slice, ok := v.Interface().([]interface{})
	if !ok {
		logs.Warn(context.Background(), "slice as not of []interface")
		return false
	}
	if idx > len(slice)-1 {
		logs.Warn(context.Background(), "string inside of [] was not an int", logs.WithError(err))

		return false
	}

	value := reflect.ValueOf(slice[idx])
	if value.Kind() != c.comparison.Kind() {
		//logs.Warn(context.Background(), "not comparing kind",
		//	logs.WithValue("slice[idx]", slice[idx]),
		//	logs.WithValue("comp", c.comparison.Kind()),
		//	logs.WithValue("value", value.Kind()))

		return false
	}

	return c.eval(value)
}
