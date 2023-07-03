package zset

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/thethan/goqueue/internal/job"
	"reflect"
	"testing"
)

const (
	envDefaultRedisHost     = "localhost:6379"
	envDefaultRedisUsername = ""
	envDefaultRedisPassword = ""
)

func Test_LRange_GetItems_Integrations(t *testing.T) {
	if testing.Short() {
		t.Skipf("skpping integration test...")
	}
	t.Run("success", func(t *testing.T) {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     envDefaultRedisHost,
			Username: envDefaultRedisUsername,
			Password: envDefaultRedisPassword,
		})

		lrange := NewZSetQueue("retry", redisClient)
		ctx := context.Background()
		jobChan := make(chan job.Job)
		go func() {
			err := lrange.GetItems(ctx, jobChan)
			require.Nil(t, err)
		}()
		for j := range jobChan {
			//val, err := j.GetValue("args")
			klass, _ := j.GetValue("class")
			errMessage, _ := j.GetValue("error_message")

			if errMessage.Kind() != reflect.Invalid && klass.String() == "WebhookWorker" {
				jid, _ := j.GetValue("jid")

				strCmd := redisClient.Get(ctx, jid.String())
				fmt.Println(strCmd.String())
			}
		}
	})
}
