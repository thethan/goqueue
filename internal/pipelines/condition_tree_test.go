package pipelines_test

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/thethan/goqueue/internal/job"
	"testing"
)

type conditionMock struct {
	mock.Mock
}

func (m *conditionMock) call(ctx context.Context, job job.Job) bool {
	args := m.Called(ctx, job)
	return args.Bool(0)
}

func TestDecisionTree_Middleware(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cMock := &conditionMock{}
		cMock.On("call", mock.Anything)
		type ConditionFunc func(ctx context.Context, job job.Job) bool
		//pipelines.NewConditionTree(cMock.call)
	})
}
