package queue

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/thethan/goqueue/internal/conditionals"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"github.com/thethan/goqueue/internal/pipelines"
	"github.com/thethan/goqueue/internal/queues"
	"github.com/thethan/goqueue/internal/redis/zset"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

const (
	envRedisHost = "REDIS_HOST"

	envRedisUsername = "REDIS_USERNAME"

	envRedisPassword = "REDIS_PASSWORD"

	envRedisAuth = "REDIS_AUTH"

	defaultRedisHost     = "localhost:6379"
	defaultRedisUsername = ""

	defaultRedisPassword = ""

	defaultRedisAuth = false
)

func init() {
	_ = viper.BindEnv(envRedisAuth, envRedisUsername, envRedisHost, envRedisPassword, envRedisHost)

	viper.SetDefault(envRedisHost, defaultRedisHost)
	viper.SetDefault(envRedisUsername, defaultRedisUsername)
	viper.SetDefault(envRedisPassword, defaultRedisPassword)
	viper.SetDefault(envRedisAuth, defaultRedisAuth)
}

func BuildPipeline(ctx context.Context, configFileLocation string) (*pipelines.Pipeline, error) {
	file, err := os.Open(configFileLocation)
	if err != nil {
		logs.Error(ctx, "could not open config file", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	decoder := yaml.NewDecoder(file)
	configuration := Configuration{}

	err = decoder.Decode(&configuration)
	if err != nil {
		logs.Fatal(ctx, "could not open config file", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	queuesMap, err := makeQueues(configuration)
	if err != nil {
		logs.Error(ctx, "could not open config file", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	conditionalMap, err := makeConditionals(configuration)
	if err != nil {
		logs.Error(ctx, "could not open config file", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	pipeline, err := makePipeline(configuration, queuesMap, conditionalMap)
	if err != nil {
		logs.Error(ctx, "could not open config file", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	return pipeline, nil
}
func makeQueues(configuration Configuration) (map[string]queues.Queue, error) {
	dataSourceNames := map[string]interface{}{}
	queues := map[string]queues.Queue{}
	for _, queueConfiguration := range configuration.DataSources {
		// todo move to its won function
		if queueConfiguration.RedisConfiguration != nil {
			redisClient := redis.NewClient(&redis.Options{
				Addr:     queueConfiguration.RedisConfiguration.Host,
				Username: queueConfiguration.RedisConfiguration.Username,
				Password: queueConfiguration.RedisConfiguration.Password,
			})
			dataSourceNames[queueConfiguration.Name] = redisClient
		}
	}
	// todo move to its own function
	for _, queueConfiguration := range configuration.Queues {
		if queueConfiguration.RedisConfiguration != nil {
			// todo move to its own function
			switch queueConfiguration.RedisConfiguration.Type {
			case ZType:
				client, ok := dataSourceNames[queueConfiguration.RedisConfiguration.Datasource]
				if !ok {
					logs.Fatal(context.Background(), "could not find redisclient", logs.WithValue("datasource", queueConfiguration.RedisConfiguration.Datasource))
				}

				redisClient, ok := client.(*redis.Client)
				if !ok {
					logs.Fatal(context.Background(), "could not find redisclient", logs.WithValue("datasource", queueConfiguration.RedisConfiguration.Datasource))
				}

				zsetQueue := zset.NewZSetQueue(queueConfiguration.RedisConfiguration.Key, redisClient)
				_ = zsetQueue

				queues[queueConfiguration.Name] = zsetQueue
			}
		}
	}

	return queues, nil
}

func makeConditionals(configuration Configuration) (map[string]conditionals.ConditionFunc, error) {
	conditionalMap := make(map[string]conditionals.ConditionFunc)
	for _, configCondition := range configuration.Conditionals {
		// todo move to its own function
		switch configCondition.Operator {
		case ">":
			conditional := conditionals.NewCondition(configCondition.Element, conditionals.GreaterThan, configCondition.Comparison)
			conditionalMap[configCondition.Name] = conditional.Evaluate
		case "contains":
			conditional := conditionals.NewCondition(configCondition.Element, conditionals.Contains, configCondition.Comparison)
			conditionalMap[configCondition.Name] = conditional.Evaluate
		}
	}

	return conditionalMap, nil
}

func makePipeline(configuration Configuration, queues map[string]queues.Queue, conditionalMap map[string]conditionals.ConditionFunc) (*pipelines.Pipeline, error) {
	// todo check length
	queueGetItems, ok := queues[configuration.Pipelines.GetItems[0].Name]
	if !ok {
		logs.Error(context.Background(), "could not find get items queue", logs.WithValue("queueName", configuration.Pipelines.GetItems[0].Name))

		return nil, errors.New("could not find get items queue")
	}

	decisionTrees := make([]*pipelines.DecisionTree, 0)
	for _, configConditional := range configuration.Pipelines.DecisionTree {
		if conditional, ok := conditionalMap[configConditional.Name]; ok {

			// get queue
			successQueue, err := getQueueForQueueFunc(queues, configConditional.Success)
			if err != nil {
				logs.Error(context.Background(), "could not find queue", logs.WithValue("queueName", configConditional.Success.Name))

				return nil, err
			}

			failureQueue, err := getQueueForQueueFunc(queues, configConditional.Failure)
			if err != nil {
				logs.Error(context.Background(), "could not find queue", logs.WithValue("queueName", configConditional.Success.Name))

				return nil, err
			}

			successFunc := getQueueFunc(successQueue, configConditional.Success)
			failureFunc := getQueueFunc(failureQueue, configConditional.Failure)

			// then return function
			decisionTree := pipelines.NewConditionTree(conditional, successFunc, failureFunc)
			decisionTrees = append(decisionTrees, &decisionTree)
		} else {
			logs.Error(context.Background(), "could not find conditional", logs.WithValue("conditionalName", configConditional.Name))
			return nil, errors.New("could not find conditional")
		}
	}

	execFunc := func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
		close(errChan)
	}

	pipeline := pipelines.NewPipeline(queueGetItems, execFunc, decisionTrees...)

	return pipeline, nil
}

func getQueue(queues map[string]queues.Queue, name string) (queues.Queue, error) {
	queue, ok := queues[name]
	if !ok {
		logs.Error(context.Background(), "could not find get items queue", logs.WithValue("queueName", name))

		return nil, errors.New("could not find pipeline func")
	}

	return queue, nil
}

func getQueueForQueueFunc(queues map[string]queues.Queue, queueFunc *PipelineConditionTreeFunc) (queues.Queue, error) {
	if queueFunc == nil {
		return nil, nil
	}

	if queueFunc.PushItem != nil {
		return getQueue(queues, queueFunc.PushItem[0].Name)
	}

	if queueFunc.RemoveItem != nil {
		return getQueue(queues, queueFunc.RemoveItem[0].Name)
	}

	return nil, errors.New("could not find queue")
}
func getQueueFunc(queue queues.Queue, condFunc *PipelineConditionTreeFunc) func(ctx context.Context, job job.Job) error {
	if condFunc == nil {
		return func(ctx context.Context, job job.Job) error {
			logs.Debug(ctx, "noop queue func", logs.WithValue("job", job))

			return nil
		}
	}

	if condFunc.PushItem != nil {
		return queue.PushItems
	}

	if condFunc.RemoveItem != nil {
		return queue.RemoveItems
	}

	return func(ctx context.Context, job job.Job) error {
		logs.Info(ctx, "noop queue func", logs.WithValue("job", job))

		return nil
	}
}

// make conditional middleware
