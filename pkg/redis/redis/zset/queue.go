package zset

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"github.com/thethan/goqueue/internal/queues"
	"io"
	"math"
	"time"
)

const count = 100

func NewZSetQueue(key string, client *redis.Client) *ZSetQueue {
	jobJuilder := job.NewBuilder(&job.Configuration{Type: "json"})

	return &ZSetQueue{key: key, client: client, jobbuilder: jobJuilder}
}

func (z *ZSetQueue) GetItems(ctx context.Context, jobChan chan<- job.Job) error {
	defer func() {
		close(jobChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			stop := fmt.Sprintf("%d.%d", time.Now().Add(time.Hour*48).Unix(), time.Now().Nanosecond())

			zRangeArgs := redis.ZRangeArgs{
				Key:     z.key,
				Offset:  0,
				Count:   count,
				Start:   "-inf",
				ByScore: true,
				Stop:    stop,
			}

			res, err := z.client.ZRangeArgsWithScores(ctx, zRangeArgs).Result()
			if err != nil {
				return err
			}

			logs.Info(ctx, "got items from queue", logs.WithValue("count", len(res)))

			for idx := range res {
				r := res[idx]
				jobStr, ok := r.Member.(string)
				if !ok {
					jobMap, ok := r.Member.(map[string]interface{})
					if !ok {
						return fmt.Errorf("could not convert job to string")
					}

					byts, err := json.Marshal(jobMap)
					if err != nil {
						return fmt.Errorf("could not convert job to string")
					}

					jobStr = string(byts)
				}

				j, err := z.jobbuilder.MakeJob([]byte(jobStr))
				if err != nil {
					return fmt.Errorf("could not convert job to string")
				}

				jobChan <- j
			}
		}
	}
}

func (z *ZSetQueue) PushItems(context.Context, job.Job, io.ReadWriter, io.ReadWriter, chan error) {
	//TODO implement me
	panic("implement me")
}

func (z *ZSetQueue) RemoveItems(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
	close(errChan)
	intCmd := z.client.ZRem(ctx, z.key, job.Raw())
	if intCmd.Err() != nil {
		errChan <- intCmd.Err()
	}

	jidVal, err := job.GetValue("jid")
	if err != nil {
		errChan <- intCmd.Err()
	}

	logs.Info(ctx, "removed from retry queue", logs.WithValue("id", jidVal.String()), logs.WithValue("int", intCmd.Val()))

}

type ZSetQueue struct {
	jobbuilder *job.Builder
	key        string
	client     *redis.Client
}

func (z *ZSetQueue) ZRangeRemove(ctx context.Context, key string) queues.RemoveItem {
	return func(ctx context.Context, jobJob job.Job, stdOut, stdErr io.ReadWriter, errChan chan error) {
		defer func() {
			close(errChan)
		}()

		newErrChan := make(chan error)
		go z.RemoveItems(ctx, jobJob, stdOut, stdErr, newErrChan)
		for err := range newErrChan {
			errChan <- err
		}

		jidVal, err := jobJob.GetValue("jid")
		if err != nil {

		}

		logs.Info(ctx, "removed from retry queue", logs.WithValue("jid", jidVal.String()))

		return
	}
}

func (z *ZSetQueue) ZRangePushItems(ctx context.Context, key string) queues.PushItems {
	return func(ctx context.Context, jobJob job.Job, stdOut io.ReadWriter, writer io.ReadWriter, errChan chan error) {
		defer func() {
			close(errChan)
		}()

		zQuery := &redis.Z{Member: jobJob.Raw(), Score: getDelay(jobJob)}
		intCmd := z.client.ZAdd(ctx, key, zQuery)
		if intCmd.Err() != nil {

			errChan <- intCmd.Err()
		}

	}
}

func getDelay(jobJob job.Job) float64 {
	retry, err := jobJob.GetValue("retry")
	if err != nil {
		logs.Warn(context.Background(), "could not get retry value", logs.WithError(err))
	}

	switch retry.Interface().(type) {
	case int, int32, int64:
		return float64(retry.Int())
	case bool:
		return 0
	case float32, float64:
		break
	}

	retryFloat := retry.Float()

	delay := math.Pow(2, retryFloat) + 15 + float64(time.Now().Unix())

	return delay

}

func (z *ZSetQueue) ZRangeGetItems(ctx context.Context, key string, errChan chan<- error, count int64) queues.GetItems {
	defer func() {
		close(errChan)
	}()

	jobChan := make(chan job.Job, 1)

	err := z.GetItems(ctx, jobChan)
	if err != nil {
		errChan <- err
		return nil
	}

	return nil
}
