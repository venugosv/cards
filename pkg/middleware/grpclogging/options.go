package grpclogging

import (
	"time"

	"github.com/anzx/pkg/log"
	"github.com/anzx/pkg/log/fabriclog"

	"google.golang.org/grpc/codes"
)

var defaultOptions = &options{
	levelFunc:    DefaultCodeToLevel,
	shouldLog:    DefaultDeciderMethod,
	codeFunc:     DefaultErrorToCode,
	durationFunc: DefaultDurationToField,
}

type options struct {
	levelFunc    CodeToLevel
	shouldLog    Decider
	codeFunc     ErrorToCode
	durationFunc DurationToField
}

type Option func(*options)

type (
	CodeToLevel     func(code codes.Code) fabriclog.Level
	DurationToField func(startTime time.Time) log.Attribute
)

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log.
func WithDecider(f Decider) Option {
	return func(o *options) {
		o.shouldLog = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// WithDurationField customizes the function for mapping request durations to log fields.
func WithDurationField(f DurationToField) Option {
	return func(o *options) {
		o.durationFunc = f
	}
}

// DefaultCodeToLevel is the default implementation of gRPC return codes and interceptor log level for server side.
func DefaultCodeToLevel(code codes.Code) fabriclog.Level {
	switch code {
	case codes.OK, codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.Unauthenticated:
		return fabriclog.InfoLevel
	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unavailable:
		return fabriclog.DebugLevel
	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return fabriclog.ErrorLevel
	default:
		return fabriclog.InfoLevel
	}
}

var DefaultDurationToField = DurationToTimeMillisField

// DurationToTimeMillisField converts the duration to milliseconds and uses the key `grpc.time_ms`.
func DurationToTimeMillisField(startTime time.Time) log.Attribute {
	return log.Dur("grpc.time_ms", time.Since(startTime))
}

// DurationToDurationField uses a Duration field to log the request duration
// and leaves it up to Zap's encoder settings to determine how that is output.
func DurationToDurationField(startTime time.Time) log.Attribute {
	return log.Dur("grpc.duration", time.Since(startTime))
}
