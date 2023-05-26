package queue

type Configuration struct {
	DataSources []DataSource `yaml:"dataSources"`
	// Queues have the ability
	Queues       []QueueConfiguration `yaml:"queues"`
	Conditionals []Conditional        `yaml:"conditionals"`
	Pipelines    Pipeline             `yaml:"pipeline"`
}

type QueueConfiguration struct {
	Name               string              `yaml:"name"`
	RedisConfiguration *RedisConfiguration `yaml:"redis,omitempty"`
}

type RedisQueueType string

const (
	ZType RedisQueueType = "zset"
	LType RedisQueueType = "list"
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
}

type PipelineGetItems struct {
	Name string `yaml:"name"`
}

type PipelineConditionTree struct {
	Name    string                     `yaml:"name"`
	Success *PipelineConditionTreeFunc `yaml:"success,omitempty"`
	Failure *PipelineConditionTreeFunc `yaml:"failure,omitempty"`
}

type PipelineConditionTreeFunc struct {
	Name       string                                `yaml:"name"`
	PushItem   []*PipelineConditionTreeFuncQueueName `yaml:"pushItems,omitempty"`
	RemoveItem []*PipelineConditionTreeFuncQueueName `yaml:"removeItems,omitempty"`
}

type PipelineConditionTreeFuncQueueName struct {
	Name string `yaml:"name"`
}
