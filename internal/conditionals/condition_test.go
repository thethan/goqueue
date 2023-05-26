package conditionals

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/thethan/goqueue/internal/redis"
	"testing"
)

func TestCondition(t *testing.T) {
	ctx := context.Background()
	t.Run("GreaterThan", func(t *testing.T) {
		t.Run("returns false if not greater than", func(t *testing.T) {
			c := NewCondition("Retries", GreaterThan, 3)

			job := &redis.RedisJob{RetryCount: 2}

			assert.False(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if greater than", func(t *testing.T) {
			c := NewCondition("Retries", GreaterThan, 3)

			job := &redis.RedisJob{RetryCount: 4}

			assert.True(t, c.Evaluate(ctx, job))
		})
		t.Run("returns false if equal ", func(t *testing.T) {
			c := NewCondition("Retries", GreaterThan, 3)

			job := &redis.RedisJob{RetryCount: 3}

			assert.False(t, c.Evaluate(ctx, job))
		})
	})
	t.Run("Contains", func(t *testing.T) {
		t.Run("returns false if not greater than", func(t *testing.T) {
			c := NewCondition("ErrorMessage", Contains, 3)

			job := &redis.RedisJob{ErrorMessage: "This is an error message"}

			assert.False(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if ErrorMessage Contains 3", func(t *testing.T) {
			c := NewCondition("ErrorMessage", Contains, "3")

			job := &redis.RedisJob{ErrorMessage: "This is an error message that has the number 3 in it"}

			assert.True(t, c.Evaluate(ctx, job))
		})
		t.Run("returns true if ErrorMessage Contains 3 and From function", func(t *testing.T) {
			c := NewCondition("GetErrorMessage", Contains, "3")

			job := &redis.RedisJob{ErrorMessage: "This is an error message that has the number 3 in it"}

			assert.True(t, c.Evaluate(ctx, job))
		})
	})
}
