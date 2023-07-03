package queue

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/thethan/goqueue/internal/conditionals"
	"github.com/thethan/goqueue/internal/executers"
	"github.com/thethan/goqueue/internal/job"
	"github.com/thethan/goqueue/internal/logs"
	"github.com/thethan/goqueue/internal/pipelines"
	"github.com/thethan/goqueue/internal/queues"
	"github.com/thethan/goqueue/pkg/redis/redis/lrange"
	"github.com/thethan/goqueue/pkg/redis/redis/zset"
	"go.opentelemetry.io/otel"
	metric2 "go.opentelemetry.io/otel/metric"
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

func BuildPipeline(ctx context.Context, configFileLocation string) (pipelines.ProcessPipeline, error) {
	file, err := os.Open(configFileLocation)
	if err != nil {
		logs.Error(ctx, "could not open config file", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	decoder := yaml.NewDecoder(file)
	configuration := Configuration{}
	//exp, err := otlpmetricgrpc.New(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//
	//meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
	//defer func() {
	//	if err := meterProvider.Shutdown(ctx); err != nil {
	//		panic(err)
	//	}
	//}()
	//otel.SetMeterProvider(meterProvider)
	//exporter, err := prometheus.New()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//provider := metric.NewMeterProvider(metric.WithReader(exporter))
	//meter := provider.Meter("github.com/open-telemetry/opentelemetry-go/example/prometheus")

	meter := otel.GetMeterProvider().Meter("github.com/open-telemetry/opentelemetry-go/example/prometheus")
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

	executors, err := makeExecutors(configuration)
	if err != nil {
		logs.Error(ctx, "could not get executors ", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	pipeline, err := makePipeline(configuration, queuesMap, conditionalMap, executors, meter)
	if err != nil {
		logs.Error(ctx, "could not open config file", logs.WithError(err), logs.WithValue("configFileLocation", configFileLocation))
		return nil, err
	}

	return pipeline, nil
}
func makeQueues(configuration Configuration) (map[string]queues.Queue, error) {
	dataSourceNames := map[string]interface{}{}
	queueMap := map[string]queues.Queue{}
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

				queueMap[queueConfiguration.Name] = zsetQueue
			case LRange:
				client, ok := dataSourceNames[queueConfiguration.RedisConfiguration.Datasource]
				if !ok {
					logs.Fatal(context.Background(), "could not find redisclient", logs.WithValue("datasource", queueConfiguration.RedisConfiguration.Datasource))
				}

				redisClient, ok := client.(*redis.Client)
				if !ok {
					logs.Fatal(context.Background(), "could not find redisclient", logs.WithValue("datasource", queueConfiguration.RedisConfiguration.Datasource))
				}

				larangeQueue := lrange.NewLRangeQueue(queueConfiguration.RedisConfiguration.Key, redisClient)

				queueMap[queueConfiguration.Name] = larangeQueue
			}

		}
	}

	return queueMap, nil
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
		case "==":
			conditional := conditionals.NewCondition(configCondition.Element, conditionals.Equal, configCondition.Comparison)
			conditionalMap[configCondition.Name] = conditional.Evaluate
		}
	}

	return conditionalMap, nil
}

func makeExecutors(configuration Configuration) (map[string]executers.ExecFunc, error) {
	execMap := make(map[string]executers.ExecFunc)
	for _, executorConfiguration := range configuration.Executors {
		// todo move to its own function
		execFunc, err := executers.NewExecutor(&executers.Configuration{
			Sprintf: executorConfiguration.SprintfCMD,
		})

		if err != nil {
			return nil, err
		}

		execMap[executorConfiguration.Name] = execFunc.Execute()
	}

	return execMap, nil
}

func makePipeline(configuration Configuration, queues map[string]queues.Queue, conditionalMap map[string]conditionals.ConditionFunc, executorsMap map[string]executers.ExecFunc, meter metric2.Meter) (pipelines.ProcessPipeline, error) {
	// todo check length
	queueGetItems, ok := queues[configuration.Pipelines.GetItems[0].Name]
	if !ok {
		logs.Error(context.Background(), "could not find get items queue", logs.WithValue("queueName", configuration.Pipelines.GetItems[0].Name))

		return nil, errors.New("could not find get items queue")
	}

	decisionTrees := make([]*pipelines.DecisionTree, 0)
	execFunc := func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
		close(errChan)
	}
	if configuration.Pipelines.Executor != nil {
		execFunc = executorsMap[configuration.Pipelines.Executor.Name]
	}
	for _, configConditional := range configuration.Pipelines.DecisionTree {
		if conditional, ok := conditionalMap[configConditional.Name]; ok {

			// get queue
			successQueue, err := getQueueForQueueFunc(queues, configConditional.Success)
			if err != nil {
				logs.Error(context.Background(), "could not find queue", logs.WithValue("queueName", configConditional.Success.Name))

				return nil, err
			}

			failureQueue, err := getQueueForQueueFunc(queues, configConditional.Failure)
			if err != nil && !configConditional.Failure.Return {
				logs.Error(context.Background(), "could not find queue", logs.WithValue("queueName", configConditional.Failure.Name))

				return nil, err
			}

			successFunc, successReturn := getQueueFunc(successQueue, configConditional.Success)
			failureFunc, falseReturn := getQueueFunc(failureQueue, configConditional.Failure)
			if err != nil {
				return nil, err
			}

			// then return function
			decisionTree := pipelines.NewConditionTree(configConditional.Name, conditional, successFunc, failureFunc, successReturn, falseReturn)
			decisionTrees = append(decisionTrees, &decisionTree)
		} else {
			logs.Error(context.Background(), "could not find conditional", logs.WithValue("conditionalName", configConditional.Name))
			return nil, errors.New("could not find conditional")
		}
	}

	pipeline := pipelines.NewPipeline(configuration.Name, meter, queueGetItems, execFunc, decisionTrees...)

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

	return &noopQueue{}, nil
}

func getExecutorFromMap(queues map[string]executers.ExecFunc, queueFunc *PipelineConditionTreeFunc) (executers.ExecFunc, error) {
	if queueFunc == nil {
		return nil, nil
	}

	if queueFunc.Executors != nil {
		name := queueFunc.Executors[0].Name
		executor, ok := queues[name]
		if !ok {
			logs.Error(context.Background(), "could not find get executorfunc", logs.WithValue("exec Func", name))

			return nil, errors.New("could not find pipeline func")
		}

		return executor, nil
	}

	return nil, errors.New("could not find executor")
}
func getQueueFunc(queue queues.Queue, condFunc *PipelineConditionTreeFunc) (executers.ExecFunc, bool) {
	if condFunc == nil {
		return func(ctx context.Context, job job.Job, stdOut io.ReadWriter, stdErr io.ReadWriter, errChan chan error) {
			defer close(errChan)
			logs.Debug(ctx, "noop queue func", logs.WithValue("job", job))
		}, false
	}

	if condFunc.PushItem != nil {
		return queue.PushItems, condFunc.Return
	}

	if condFunc.RemoveItem != nil {
		return queue.RemoveItems, condFunc.Return
	}

	return func(ctx context.Context, job job.Job, stdOut, stdErr io.ReadWriter, errChan chan error) {
		defer close(errChan)
		logs.Debug(ctx, "noop queue func", logs.WithValue("job", job))

	}, condFunc.Return
}

// make conditional middleware

type noopQueue struct {
}

func (queue *noopQueue) GetItems(ctx context.Context, jobChan chan<- job.Job) error {
	//TODO implement me
	return nil
}

func (queue *noopQueue) PushItems(ctx context.Context, job job.Job, stdOut, stdErr io.ReadWriter, error chan error) {
	close(error)
}
func (queue *noopQueue) RemoveItems(ctx context.Context, job job.Job, stdOut, stdErr io.ReadWriter, error chan error) {
	close(error)
}
