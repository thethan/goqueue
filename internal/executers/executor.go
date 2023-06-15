package executers

import (
	"context"
	"fmt"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"io"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type executor struct {
	configuration *Configuration
	reg           *regexp.Regexp
}

func NewExecutor(configuration *Configuration) (*executor, error) {
	left := "{"
	right := "}"
	reg, err := regexp.Compile(`(?s)` + regexp.QuoteMeta(left) + `(.*?)` + regexp.QuoteMeta(right))
	if err != nil {
		return nil, err
	}

	return &executor{
		configuration: configuration,
		reg:           reg,
	}, nil
}

func (e *executor) Execute() ExecFunc {
	return func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
		defer func() {
			fmt.Println("closing err chan")
			close(errChan)
		}()

		var err error
		funcExecString := e.configuration.Sprintf

		returns := e.reg.FindAllString(e.configuration.Sprintf, -1)
		for _, v := range returns {
			logs.Debug(ctx, "found", logs.WithValue("value", v))

			funcExecString, err = buildString(funcExecString, v, job)
			if err != nil {
				errChan <- err
				return
			}
		}

		logs.Debug(ctx, "executing", logs.WithValue("cmd", funcExecString))

		cmd := exec.CommandContext(ctx, funcExecString)

		cmd.Stdout = stdOut
		cmd.Stderr = stdErr

		err = cmd.Run()
		if err != nil {
			errChan <- err
		}

		return
	}

}

func buildString(cmdString string, stringToReplace string, j job.Job) (string, error) {
	key := stringToReplace[1 : len(stringToReplace)-1]
	val, err := j.GetValue(key)
	if err != nil {
		return "", err
	}

	cmdString = strings.Replace(cmdString, stringToReplace, getValueFrom(val), -1)

	return cmdString, err
}

func getValueFrom(val reflect.Value) string {
	var newVal string

	switch val.Kind() {
	case reflect.String:
		newVal = val.String()
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		newVal = strconv.FormatInt(val.Int(), 10)
	case reflect.Float32, reflect.Float64:
		newVal = fmt.Sprint(val.Float())
	case reflect.Slice:
		strSlice := make([]string, val.Len())
		interfaceSLice := val.Interface().([]interface{})

		for i := range interfaceSLice {
			strSlice[i] = getValueFrom(reflect.ValueOf(interfaceSLice[i]))
		}

		newVal = strings.Join(strSlice, ",")
	}

	return newVal
}
