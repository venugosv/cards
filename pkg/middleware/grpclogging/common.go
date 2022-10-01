package grpclogging

import (
	"context"
	"strings"

	"github.com/anzx/pkg/log"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/codes"
)

type PayloadLoggingDecider struct {
	Server map[string]bool `json:"server" yaml:"server" mapstructure:"server"`
	Client map[string]bool `json:"client" yaml:"client" mapstructure:"client"`
}

// ErrorToCode function determines the error code of an error
// This makes using custom errors with grpc middleware easier
type ErrorToCode func(err error) codes.Code

func DefaultErrorToCode(err error) codes.Code {
	return status.Code(err)
}

// Decider function defines rules for suppressing any interceptor logs
type Decider func(fullMethodName string, err error) bool

// DefaultDeciderMethod is the default implementation of decider to see if you should log the call
// by default this if always true so all calls are logged
func ServerDecider(cfg PayloadLoggingDecider) Decider {
	return func(fullMethodName string, err error) bool {
		if decision, ok := cfg.Server[strings.ToLower(fullMethodName)]; ok {
			return decision
		}
		return false
	}
}

func DefaultDeciderMethod(_ string, _ error) bool {
	return true
}

// ServerPayloadLoggingDecider is a user-provided function for deciding whether to log the server-side
// request/response payloads
type ServerPayloadLoggingDecider func(ctx context.Context, fullMethodName string, servingObject interface{}) bool

func ServerPayloadDecider(cfg PayloadLoggingDecider) ServerPayloadLoggingDecider {
	return func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
		if decision, ok := cfg.Server[strings.ToLower(fullMethodName)]; ok {
			return decision
		}
		return false
	}
}

// ClientPayloadLoggingDecider is a user-provided function for deciding whether to log the client-side
// request/response payloads
type ClientPayloadLoggingDecider func(ctx context.Context, fullMethodName string) bool

func ClientPayloadDecider(cfg PayloadLoggingDecider) ClientPayloadLoggingDecider {
	return func(ctx context.Context, fullMethodName string) bool {
		if decision, ok := cfg.Client[strings.ToLower(fullMethodName)]; ok {
			return decision
		}
		return false
	}
}

func fromMapToAttributes(fields map[string]interface{}) []log.Attribute {
	attr := []log.Attribute{}

	for i, v := range fields {
		attr = append(attr, log.Any(i, v))
	}

	return attr
}
