package job

import (
	"context"
	"io"
)

type Job interface {
	Class() string
	Execute(ctx context.Context) error

	//Retries() int32
	//bytes() []byte

	//Raw() []byte
}

type JobWrapper struct {
	Job

	raw    []byte
	StdErr io.ReadWriter
	StdOut io.ReadWriter
}
