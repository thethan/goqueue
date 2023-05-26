package redis

import (
	"context"
	"encoding/json"
	"github.com/thethan/goqueue/internal/logs"
)

type RedisJob struct {
	raw          string
	Retry        interface{}   `json:"retry"`
	Queue        string        `json:"queue"`
	Klass        string        `json:"class"`
	Args         []interface{} `json:"args"`
	Job          string        `json:"job,omitempty"`
	Jid          string        `json:"jid"`
	RetryCount   int32         `json:"retry_count,omitempty"`
	CreatedAt    float64       `json:"created_at"`
	EnqueuedAt   float64       `json:"enqueued_at"`
	ErrorMessage string        `json:"error_message,omitempty"`
	ErrorClass   string        `json:"error_class,omitempty"`
}

func (i *RedisJob) Class() string {
	//TODO implement me
	return i.Klass
}

func (i *RedisJob) Execute(ctx context.Context) error {
	//TODO implement me
	return nil
}

func (i *RedisJob) SetRaw(s string) {
	i.raw = s
}

func (i *RedisJob) GetErrorMessage() string {
	//TODO implement me
	return i.ErrorMessage
}

func (i *RedisJob) Raw() []byte {
	if i.raw != "" {
		return []byte(i.raw)
	}

	data, err := json.Marshal(i)
	if err != nil {
		logs.Error(context.Background(), "data was nil", logs.WithError(err), logs.WithValue("job", i))
		return data
	}

	return data
}

func (i *RedisJob) Retries() int32 {
	return i.RetryCount
}

func (i *RedisJob) MarshalBinary() (data []byte, err error) {
	i.raw = string(data)
	return json.Marshal(i)
}

func (i *RedisJob) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i)
}
