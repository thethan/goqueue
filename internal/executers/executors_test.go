package executers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type mockJob struct {
	mock.Mock
	maps map[string]interface{}
}

func (m mockJob) GetValue(key string) (reflect.Value, error) {
	args := m.Called(key)
	return args.Get(0).(reflect.Value), args.Error(1)
}

func (m mockJob) Raw() []byte {
	//TODO implement me
	panic("implement me")
}

func TestExecutors(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Run("execute processor", func(t *testing.T) {
			execu, err := NewExecutor(&Configuration{
				Sprintf: "bundle exec {klass}({args})",
			})

			require.Nil(t, err)
			ctx := context.Background()
			job := &mockJob{}
			job.On("GetValue", "klass").Return(reflect.ValueOf("test"), nil)
			job.On("GetValue", "args").Return(reflect.ValueOf([]interface{}{12388, "some"}), nil)

			stdErr := bytes.NewBufferString("")
			stdOut := bytes.NewBufferString("")

			errChan := make(chan error)
			go func() {
				execu.Execute()(ctx, job, stdOut, stdErr, errChan)
			}()

			for err = range errChan {
				fmt.Println(err)
			}

			mock.AssertExpectationsForObjects(t, job)
		})
	})
	t.Run("failure", func(t *testing.T) {

	})
}
