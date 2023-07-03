package conditionals

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	job2 "github.com/thethan/goqueue/internal/job"
	"testing"
)

func TestCondition(t *testing.T) {
	ctx := context.Background()
	t.Run("GreaterThan", func(t *testing.T) {
		t.Run("returns false if not greater than", func(t *testing.T) {
			c := NewCondition("retry_count", GreaterThan, float64(3))

			jsonString := `{"retry_count": 2}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.False(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if greater than", func(t *testing.T) {
			c := NewCondition("retry_count", GreaterThan, float64(3))

			jsonString := `{"retry_count": 4}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
		t.Run("returns false if equal ", func(t *testing.T) {
			c := NewCondition("retry_count", GreaterThan, float64(3))

			jsonString := `{"retry_count": 3}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.False(t, c.Evaluate(ctx, job))
		})

		t.Run("returns true if comparison is int64 ", func(t *testing.T) {
			c := NewCondition("retry_count", GreaterThan, int64(2))

			jsonString := `{"retry_count": 3}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if comparison is int ", func(t *testing.T) {
			c := NewCondition("retry_count", GreaterThan, int(2))

			jsonString := `{"retry_count": 3}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if comparison is int32 ", func(t *testing.T) {
			c := NewCondition("retry_count", GreaterThan, int32(2))

			jsonString := `{"retry_count": 3}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
	})
	t.Run("Contains", func(t *testing.T) {
		t.Run("returns false if error message does not contain 3", func(t *testing.T) {
			c := NewCondition("error_message", Contains, "Nil Error")

			jsonString := `{"error_message": "this is an error"}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.False(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if ErrorMessage Contains 3", func(t *testing.T) {
			c := NewCondition("error_message", Contains, "Nil Error")

			jsonString := `{"error_message": "Error runnning: Nil Error"}`
			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})

			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if ErrorMessage Contains 3 and From function", func(t *testing.T) {
			c := NewCondition("error_message", Contains, "3")

			jsonString := `{"error_message": "This is an error message that has the number 3 in it"}`
			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})

			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
	})

	t.Run("Equal", func(t *testing.T) {
		t.Run("returns false if not equal", func(t *testing.T) {
			c := NewCondition("retry_count", Equal, float64(3))

			jsonString := `{"retry_count": 2}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.False(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if equal", func(t *testing.T) {
			c := NewCondition("retry_count", Equal, float64(3))

			jsonString := `{"retry_count": 3}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})

		t.Run("returns false if not equal - string", func(t *testing.T) {
			c := NewCondition("retry_count", Equal, "not a test string")

			jsonString := `{"retry_count": "test string"}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.False(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if equal - string", func(t *testing.T) {
			c := NewCondition("retry_count", Equal, "test string")

			jsonString := `{"retry_count": "test string"}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
	})
}

func TestCondition_FromSlice(t *testing.T) {
	ctx := context.Background()
	t.Run("Success", func(t *testing.T) {
		t.Run("returns false if not equal", func(t *testing.T) {
			c := NewCondition("args[0]", Equal, 6)

			jsonString := `{"args": [5, "some message"]}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.False(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if equal - int", func(t *testing.T) {
			c := NewCondition("args[0]", Equal, 5)

			jsonString := `{"args": [5, "some message"]}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.True(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if equal - int", func(t *testing.T) {
			c := NewCondition("args[2]", Equal, "some message")

			jsonString := `{"args": [5, "some message"]}`

			builder := job2.NewBuilder(&job2.Configuration{Type: "json"})
			job, err := builder.MakeJob([]byte(jsonString))
			require.Nil(t, err)

			assert.False(t, c.Evaluate(ctx, job))
		})
	})
}
