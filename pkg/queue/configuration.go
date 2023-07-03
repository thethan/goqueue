package queue

type Configuration struct {
	Name        string       `yaml:"name"`
	DataSources []DataSource `yaml:"dataSources"`
	// Queues have the ability
	Queues       []QueueConfiguration    `yaml:"queues"`
	Conditionals []Conditional           `yaml:"conditionals"`
	Pipelines    Pipeline                `yaml:"pipeline"`
	Executors    []ExecutorConfiguration `yaml:"executors"`
}

type QueueConfiguration struct {
	Name               string              `yaml:"name"`
	RedisConfiguration *RedisConfiguration `yaml:"redis,omitempty"`
}

type RedisQueueType string

const (
	ZType  RedisQueueType = "zset"
	LRange RedisQueueType = "lrange"
)

type RedisConfiguration struct {
	Key        string         `yaml:"key"`
	Datasource string         `yaml:"dataSource"`
	Type       RedisQueueType `yaml:"type"`
}

type DataSource struct {
	Name               string       `yaml:"name"`
	RedisConfiguration *RedisClient `yaml:"redis,omitempty"`
}

type RedisClient struct {
	Host     string `yaml:"host"`
	Auth     bool   `yaml:"auth,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Conditional struct {
	Name       string      `yaml:"name"`
	Operator   string      `yaml:"operator"`
	Element    string      `yaml:"element"`
	Comparison interface{} `yaml:"comparison"`
}

type Pipeline struct {
	GetItems     []PipelineGetItems      `yaml:"getItems"`
	DecisionTree []PipelineConditionTree `yaml:"decisionTree"`
	Executor     *ExecutorConfiguration  `yaml:"executor,omitempty"`
}

type PipelineGetItems struct {
	Name string `yaml:"name"`
}

type PipelineConditionTree struct {
	Name     string                     `yaml:"name"`
	Success  *PipelineConditionTreeFunc `yaml:"success,omitempty"`
	Failure  *PipelineConditionTreeFunc `yaml:"failure,omitempty"`
	Executor *PipelineConditionTreeFunc `yaml:"executor,omitempty"`
	Return   bool                       `yaml:"return"`
}

type PipelineConditionTreeFunc struct {
	Name       string                                `yaml:"name"`
	PushItem   []*PipelineConditionTreeFuncQueueName `yaml:"pushItems,omitempty"`
	RemoveItem []*PipelineConditionTreeFuncQueueName `yaml:"removeItems,omitempty"`
	Executors  []*PipelineConditionTreeFuncQueueName `yaml:"executors,omitempty"`
	Return     bool                                  `yaml:"return"`
}

type PipelineConditionTreeFuncQueueName struct {
	Name string `yaml:"name"`
}

type ExecutorConfiguration struct {
	Name       string `yaml:"name"`
	SprintfCMD string `yaml:"command"`
}
