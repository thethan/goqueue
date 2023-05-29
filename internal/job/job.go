package job

import (
	"errors"
	"io"
	"reflect"
)

type Job interface {
	Data
}

type Data interface {
	GetValue(key string) (reflect.Value, error)
	Raw() []byte
}

type JobWrapper struct {
	Job

	raw    []byte
	StdErr io.ReadWriter
	StdOut io.ReadWriter
}

type Builder struct {
	config *Configuration
}

func NewBuilder(configuration *Configuration) *Builder {
	return &Builder{
		config: configuration,
	}
}

func (jF *Builder) MakeJob(bts []byte) (Data, error) {
	if jF.config == nil {
		return nil, errors.New("invalid configuration")
	}

	if jF.config.Type == "json" {
		return makeJsonJob(jF.config, bts)
	}

	return nil, errors.New("invalid job type")
}
