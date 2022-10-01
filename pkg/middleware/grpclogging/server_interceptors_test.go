package grpclogging

import (
	"fmt"
	"testing"
	"time"

	"github.com/anzx/pkg/log/fabriclog"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func customCodeToLevel(c codes.Code) fabriclog.Level {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return fabriclog.ErrorLevel
	}
	return DefaultCodeToLevel(c)
}

func TestLoggingSuite(t *testing.T) {
	opts := []Option{
		WithLevels(customCodeToLevel),
	}
	b := newZRBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			UnaryServerInterceptor(opts...)),
	}
	suite.Run(t, &ServerSuite{b})
}

type ServerSuite struct {
	*BaseSuite
}

func (s *ServerSuite) TestPing_WithCustomTags() {
	fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(s.buffer))

	deadline := time.Now().Add(3 * time.Second)
	_, err := s.Client.Ping(s.DeadlineCtx(deadline), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "two log statements should be logged")

	for _, m := range msgs {
		attr := m["attributes"].(map[string]interface{})
		assert.Equal(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
		assert.Equal(s.T(), "Ping", attr["grpc.method"], "all lines must contain method name")
		assert.Equal(s.T(), "server", attr["span.kind"], "all lines must contain the kind of call (server)")
		assert.Equal(s.T(), "something", attr["custom_tags.string"], "all lines must contain `custom_tags.string`")
		assert.Equal(s.T(), "something", attr["grpc.request.value"], "all lines must contain fields extracted")
		assert.Equal(s.T(), "custom_value", attr["custom_field"], "all lines must contain `custom_field`")

		assert.Contains(s.T(), attr, "custom_tags.int", "all lines must contain `custom_tags.int`")
		require.Contains(s.T(), attr, "grpc.start_time", "all lines must contain the start time")
		_, err := time.Parse(time.RFC3339, attr["grpc.start_time"].(string))
		assert.NoError(s.T(), err, "should be able to parse start time as RFC3339")

		require.Contains(s.T(), attr, "grpc.request.deadline", "all lines must contain the deadline of the call")
		_, err = time.Parse(time.RFC3339, attr["grpc.request.deadline"].(string))
		require.NoError(s.T(), err, "should be able to parse deadline as RFC3339")
		assert.Equal(s.T(), attr["grpc.request.deadline"], deadline.Format(time.RFC3339), "should have the same deadline that was set by the caller")
	}

	assert.Equal(s.T(), "some ping", msgs[0]["body"], "handler's message must contain user message")

	assert.Equal(s.T(), "finished unary call with code OK", msgs[1]["body"], "handler's message must contain user message")
	assert.Equal(s.T(), "info", msgs[1]["severity"], "must be logged at info level")
	assert.Contains(s.T(), fmt.Sprintf("%v", msgs[1]), "grpc.time_ms", "interceptor log statement should contain execution time")
}

func (s *ServerSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code          codes.Code
		level         fabriclog.Level
		expectedLevel string
		msg           string
	}{
		{
			code:          codes.Internal,
			level:         fabriclog.ErrorLevel,
			expectedLevel: "error",
			msg:           "Internal must remap to ErrorLevel in defaultCodeToLevel",
		},
		{
			code:          codes.NotFound,
			level:         fabriclog.InfoLevel,
			expectedLevel: "info",
			msg:           "NotFound must remap to InfoLevel in defaultCodeToLevel",
		},
		{
			code:          codes.Unauthenticated,
			level:         fabriclog.ErrorLevel,
			expectedLevel: "error",
			msg:           "Unauthenticated is overwritten to ErrorLevel with customCodeToLevel override, which probably didn't work",
		},
	} {
		s.buffer.Reset()
		fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(s.buffer))

		_, err := s.Client.PingError(s.SimpleCtx(), &pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})
		require.Error(s.T(), err, "each call here must return an error")

		msgs := s.getOutputJSONs()
		require.Len(s.T(), msgs, 2, "only the interceptor log messages are printed in PingErr")

		attr := msgs[0]["attributes"].(map[string]interface{})
		assert.Equal(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
		assert.Equal(s.T(), "PingError", attr["grpc.method"], "all lines must contain method name")
		assert.Equal(s.T(), tcase.code.String(), attr["grpc.code"], "all lines have the correct gRPC code")
		assert.Equal(s.T(), tcase.expectedLevel, msgs[0]["severity"], tcase.msg)
		assert.Equal(s.T(), "finished unary call with code "+tcase.code.String(), msgs[0]["body"], "needs the correct end message")

		require.Contains(s.T(), fmt.Sprintf("%v", msgs[0]), "grpc.start_time", "all lines must contain the start time")
		_, err = time.Parse(time.RFC3339, attr["grpc.start_time"].(string))
		assert.NoError(s.T(), err, "should be able to parse start time as RFC3339")
	}
}

func TestLoggingOverrideSuite(t *testing.T) {
	opts := []Option{
		WithDurationField(DurationToDurationField),
	}
	b := newZRBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			UnaryServerInterceptor(opts...)),
	}
	suite.Run(t, &ServerOverrideSuite{b})
}

type ServerOverrideSuite struct {
	*BaseSuite
}

func (s *ServerOverrideSuite) TestPing_HasOverriddenDuration() {
	fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(s.buffer))

	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "two log statements should be logged")

	for _, m := range msgs {
		attr := m["attributes"].(map[string]interface{})
		assert.Equal(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
		assert.Equal(s.T(), "Ping", attr["grpc.method"], "all lines must contain method name")
	}

	assert.Equal(s.T(), "some ping", msgs[0]["body"], "handler's message must contain user message")
	assert.NotContains(s.T(), msgs[0], "handler's message must not contain default duration")
	assert.NotContains(s.T(), msgs[0], "grpc.duration", "handler's message must not contain overridden duration")

	assert.Equal(s.T(), "finished unary call with code OK", msgs[1]["body"], "handler's message must contain user message")
	assert.Equal(s.T(), "info", msgs[1]["severity"], "OK error codes must be logged on info level.")
	assert.NotContains(s.T(), fmt.Sprintf("%v", msgs[1]), "grpc.time_ms", "handler's message must not contain default duration")
	assert.Contains(s.T(), fmt.Sprintf("%v", msgs[1]), "grpc.duration", "handler's message must contain overridden duration")
}

func TestServerOverrideSuppressedSuite(t *testing.T) {
	opts := []Option{
		WithDecider(func(method string, err error) bool {
			if err != nil && method == "/mwitkow.testproto.TestService/PingError" {
				return true
			}
			return false
		}),
	}
	b := newZRBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			UnaryServerInterceptor(opts...)),
	}
	suite.Run(t, &ServerOverridenDeciderSuite{b})
}

type ServerOverridenDeciderSuite struct {
	*BaseSuite
}

func (s *ServerOverridenDeciderSuite) TestPing_HasOverriddenDecider() {
	fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(s.buffer))

	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "single log statements should be logged")

	attr := msgs[0]["attributes"].(map[string]interface{})
	assert.Equal(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
	assert.Equal(s.T(), "Ping", attr["grpc.method"], "all lines must contain method name")
	assert.Equal(s.T(), "some ping", msgs[0]["body"], "handler's message must contain user message")
}

func (s *ServerOverridenDeciderSuite) TestPingError_HasOverriddenDecider() {
	fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(s.buffer))

	code := codes.NotFound
	msg := "NotFound must remap to InfoLevel in DefaultCodeToLevel"

	s.buffer.Reset()
	_, err := s.Client.PingError(
		s.SimpleCtx(),
		&pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(code)})
	require.Error(s.T(), err, "each call here must return an error")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "only the interceptor log messages are printed in PingErr")
	attr := msgs[0]["attributes"].(map[string]interface{})
	assert.Equal(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
	assert.Equal(s.T(), "PingError", attr["grpc.method"], "all lines must contain method name")
	assert.Equal(s.T(), code.String(), attr["grpc.code"], "all lines must contain the correct gRPC code")
	assert.Equal(s.T(), "info", msgs[0]["severity"], msg)
}
