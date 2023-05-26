package zset

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/queues"

	"github.com/thethan/goqueue/internal/logs"
	goqueueRedis "github.com/thethan/goqueue/internal/redis"
	"math"
	"time"
)

const count = 100

func NewZSetQueue(key string, client *redis.Client) *ZSetQueue {
	return &ZSetQueue{key: key, client: client}
}

func (r *ZSetQueue) GetItems(ctx context.Context, jobChan chan<- job.Job) error {
	defer func() {
		close(jobChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			stop := fmt.Sprintf("%d.%d", time.Now().Add(time.Hour*48).Unix(), time.Now().Nanosecond())

			z := redis.ZRangeArgs{
				Key:     r.key,
				Offset:  0,
				Count:   count,
				Start:   "-inf",
				ByScore: true,
				Stop:    stop,
			}

			res, err := r.client.ZRangeArgsWithScores(ctx, z).Result()
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

				j := goqueueRedis.RedisJob{}

				err = json.Unmarshal([]byte(jobStr), &j)
				if err != nil {
					return fmt.Errorf("could not convert job to string")
				}

				j.SetRaw(jobStr)

				jobChan <- &j
			}
		}
	}
}

func (r *ZSetQueue) PushItems(ctx context.Context, job job.Job) error {
	//TODO implement me
	panic("implement me")
}

func (r *ZSetQueue) RemoveItems(ctx context.Context, job job.Job) error {
	sideJob, ok := job.(*goqueueRedis.RedisJob)
	if !ok {
		return fmt.Errorf("could not convert job to sidekiq job")
	}

	retry := sideJob.RetryCount
	if retry < 3 {
		return nil
	}

	intCmd := r.client.ZRem(ctx, r.key, sideJob.Raw())
	if intCmd.Err() != nil {
		return intCmd.Err()
	}

	logs.Info(ctx, "removed from retry queue", logs.WithValue("jid", sideJob.Jid), logs.WithValue("int", intCmd.Val()))

	return nil
}

type ZSetQueue struct {
	key    string
	client *redis.Client
}

func (r *ZSetQueue) ZRangeRemove(ctx context.Context, errorChan chan error) queues.RemoveItem {
	return func(ctx context.Context, key string, jobJob job.Job) error {
		defer func() {
			close(errorChan)
		}()

		err := r.RemoveItems(ctx, jobJob)
		if err != nil {
			errorChan <- err
		}

		redisJob := jobJob.(*goqueueRedis.RedisJob)

		logs.Info(ctx, "removed from retry queue", logs.WithValue("jid", redisJob.Jid))

		return nil
	}
}

func (r *ZSetQueue) ZRangePushItems(ctx context.Context, errChan chan<- error) queues.PushItems {
	return func(ctx context.Context, key string, jobJob job.Job) error {
		defer func() {
			close(errChan)
		}()

		j := jobJob.(*goqueueRedis.RedisJob)

		z := &redis.Z{Member: j.Raw(), Score: getDelay(j)}
		intCmd := r.client.ZAdd(ctx, key, z)
		if intCmd.Err() != nil {

			errChan <- intCmd.Err()
			return nil
		}

		return nil
	}
}

func getDelay(sidekiqJob *goqueueRedis.RedisJob) float64 {
	if sidekiqJob.Retry == 0 || sidekiqJob.Retry == true {
		return 0
	}

	retry, ok := sidekiqJob.Retry.(float64)
	if !ok {
		retry = 1
	}

	delay := math.Pow(2, retry) + 15 + float64(time.Now().Unix())

	return delay

}

func (r *ZSetQueue) ZRangeGetItems(ctx context.Context, key string, errChan chan<- error, count int64) queues.GetItems {
	defer func() {
		close(errChan)
	}()

	jobChan := make(chan job.Job, 1)

	err := r.GetItems(ctx, jobChan)
	if err != nil {
		errChan <- err
		return nil
	}

	return nil
}
