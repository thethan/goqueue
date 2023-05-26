package logs

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	loggerSettingsJSON = `{
		"level": "debug",
		"encoding": "console",
		"outputPaths": ["stderr"],
	  	"errorOutputPaths": ["stderr"],
		"development": false,
		"disableStacktrace": true,
	  	"encoderConfig": {
	    	"messageKey": "message",
	    	"levelKey": "level",
	    	"levelEncoder": "capital",
			"timeKey": "timestamp",
			"timeEncoder": "iso8601",
			"callerKey": "caller",
			"callerEncoder": "short",
			"stacktraceKey": "stack"
	  	}
	}`

	correlationIDKey   = "correlation_id"
	requestIDKey       = "request_id"
	userAgentKey       = "user_agent"
	ipAddressKey       = "ip_address"
	originUserAgentKey = "origin_user_agent"
	pathKey            = "pathKey"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
	INFO  LogLevel = "INFO"
)

type LogFormat string

const (
	CONSOLE LogFormat = "console"
	JSON    LogFormat = "json"
)

type Log struct {
	Level      LogLevel
	Time       time.Time
	LoggerName string
	Message    string
	Caller     string
	Stack      string
}

// nolint: gochecknoglobals
var l *zap.Logger

// nolint: gochecknoinits
func init() {
	l = New(
		LogFormat(os.Getenv("LOG_FORMAT")),
		LogLevel(os.Getenv("LOG_LEVEL")),
	)
}

func withGrpcContext(ctx context.Context) []zapcore.Field {
	fields := make([]zapcore.Field, 0)

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return fields
	}

	if correlationIDHeader := md.Get("x-skuiq-correlation-id"); len(correlationIDHeader) == 1 {
		fields = append(fields, zap.String(correlationIDKey, correlationIDHeader[0]))
	}

	if requestIDHeader := md.Get("x-request-id"); len(requestIDHeader) == 1 {
		fields = append(fields, zap.String(requestIDKey, requestIDHeader[0]))
	}

	if UAHeader := md.Get("user-agent"); len(UAHeader) == 1 {
		fields = append(fields, zap.String(userAgentKey, UAHeader[0]))
	}

	if XFFHeader := md.Get("x-forwarded-for"); len(XFFHeader) > 0 {
		fields = append(fields, zap.String(ipAddressKey, XFFHeader[0]))
	}

	if XFUAHeader := md.Get("x-forwarded-user-agent"); len(XFUAHeader) > 0 {
		fields = append(fields, zap.String(originUserAgentKey, XFUAHeader[0]))
	}

	method, ok := grpc.Method(ctx)

	if ok {
		// Method looks something like - skuiq.proto.Api/Rpc
		splitMethod := strings.Split(method, ".")

		// Get the last value - Api/Rpc
		method = splitMethod[len(splitMethod)-1]
		// Do a char replacement to . to properly index - Api.Rpc
		method = strings.ReplaceAll(method, "/", ".")

		fields = append(fields, zap.String(pathKey, method))
	}

	return fields
}

func withContext(ctx context.Context) []zapcore.Field {
	fields := make([]zapcore.Field, 0)

	fields = append(fields, withRequestID(ctx)...)
	fields = append(fields, withGrpcContext(ctx)...)

	return fields
}

func withRequestID(ctx context.Context) []zapcore.Field {
	fields := make([]zapcore.Field, 0)

	requestIDInterface := ctx.Value(requestIDKey)
	if requestIDInterface != nil {
		switch value := requestIDInterface.(type) {
		case uint8:
			stringValue := strconv.Itoa(int(value))
			fields = append(fields, zap.String(requestIDKey, stringValue))
		case []uint8:
			stringValue := string(value)
			fields = append(fields, zap.String(requestIDKey, stringValue))
		case string:
			fields = append(fields, zap.String(requestIDKey, value))
		}
	}

	return fields
}

func withStackTrace(fields []zapcore.Field) []zapcore.Field {
	return append(fields, zap.Stack("stacktrace"))
}

type LogField struct {
	Key   string
	Value interface{}
}

type LogOptions struct {
	Err    error
	Fields []LogField
}

type LogOption func(*LogOptions)

func newOptionFields(ctx context.Context, errOpts ...LogOption) []zapcore.Field {
	fields := withContext(ctx)
	opts := &LogOptions{}

	for _, option := range errOpts {
		option(opts)
	}

	for _, field := range opts.Fields {
		fields = append(fields, zap.Reflect(field.Key, field.Value))
	}

	if opts.Err != nil {
		fields = append(fields, zap.Error(opts.Err))
	}

	return fields
}

func WithValue(key string, value interface{}) LogOption {
	return func(opts *LogOptions) {
		opts.Fields = append(opts.Fields, LogField{
			Key:   key,
			Value: value,
		})
	}
}

func WithValues(m map[string]any) LogOption {
	return func(opts *LogOptions) {
		for key, value := range m {
			opts.Fields = append(opts.Fields, LogField{
				Key:   key,
				Value: value,
			})
		}
	}
}

func WithError(err error) LogOption {
	return func(opts *LogOptions) {
		opts.Err = err
	}
}

func New(env LogFormat, logLevel LogLevel) *zap.Logger {
	var cfg zap.Config
	if err := json.Unmarshal([]byte(loggerSettingsJSON), &cfg); err != nil {
		panic(err)
	}

	if env == JSON {
		cfg.Encoding = "json"
	}

	switch logLevel {
	case DEBUG:
		cfg.Level.SetLevel(zap.DebugLevel)
	case ERROR:
		cfg.Level.SetLevel(zap.ErrorLevel)
	case WARN:
		cfg.Level.SetLevel(zap.WarnLevel)
	case INFO:
		cfg.Level.SetLevel(zap.InfoLevel)
	default:
		cfg.Level.SetLevel(zap.InfoLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	l = logger.WithOptions(zap.AddCallerSkip(1))

	return l
}

func Debug(ctx context.Context, msg string, opts ...LogOption) {
	l.Debug(msg, newOptionFields(ctx, opts...)...)
}

func Info(ctx context.Context, msg string, opts ...LogOption) {
	l.Info(msg, newOptionFields(ctx, opts...)...)
}

func Warn(ctx context.Context, msg string, opts ...LogOption) {
	l.Warn(msg, newOptionFields(ctx, opts...)...)
}

func Error(ctx context.Context, msg string, opts ...LogOption) {
	l.Error(msg, withStackTrace(newOptionFields(ctx, opts...))...)
}

func Fatal(ctx context.Context, msg string, opts ...LogOption) {
	l.Fatal(msg, withStackTrace(newOptionFields(ctx, opts...))...)
}

func Panic(ctx context.Context, msg string, opts ...LogOption) {
	l.Panic(msg, withStackTrace(newOptionFields(ctx, opts...))...)
}

func DPanic(ctx context.Context, msg string, opts ...LogOption) {
	l.DPanic(msg, withStackTrace(newOptionFields(ctx, opts...))...)
}

func Sync() {
	_ = l.Sync()
}
