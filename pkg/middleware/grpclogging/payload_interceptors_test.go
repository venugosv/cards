package grpclogging

import (
	"context"
	"fmt"
	"testing"

	"github.com/anzx/pkg/log/fabriclog"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

func TestPayloadSuite(t *testing.T) {
	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }

	b := newZRBaseSuite(t)
	b.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(PayloadUnaryClientInterceptor(alwaysLoggingDeciderClient)),
	}
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			UnaryServerInterceptor(),
			PayloadUnaryServerInterceptor(alwaysLoggingDeciderServer)),
	}
	suite.Run(t, &PayloadSuite{b})
}

type PayloadSuite struct {
	*BaseSuite
}

func (s *PayloadSuite) getServerAndClientMessages(expectedServer int, expectedClient int) (serverMsgs []map[string]interface{}, clientMsgs []map[string]interface{}) {
	msgs := s.getOutputJSONs()
	for _, m := range msgs {
		attr := m["attributes"].(map[string]interface{})
		if attr["span.kind"] == "server" {
			serverMsgs = append(serverMsgs, m)
		} else if attr["span.kind"] == "client" {
			clientMsgs = append(clientMsgs, m)
		}
	}
	require.Len(s.T(), serverMsgs, expectedServer, "must match expected number of server log messages")
	require.Len(s.T(), clientMsgs, expectedClient, "must match expected number of client log messages")
	return serverMsgs, clientMsgs
}

func (s *PayloadSuite) TestPing_LogsBothRequestAndResponse() {
	fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(s.buffer))

	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)

	require.NoError(s.T(), err, "there must be not be an error on a successful call")
	serverMsgs, clientMsgs := s.getServerAndClientMessages(4, 2)
	for _, m := range append(serverMsgs, clientMsgs...) {
		attr := m["attributes"].(map[string]interface{})
		assert.Equal(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
		assert.Equal(s.T(), "Ping", attr["grpc.method"], "all lines must contain method name")
		assert.Equal(s.T(), "info", m["severity"], "all payloads must be logged on info level")
	}

	serverReq, serverResp := serverMsgs[0], serverMsgs[2]
	clientReq, clientResp := clientMsgs[0], clientMsgs[1]
	assert.Contains(s.T(), fmt.Sprintf("%v", clientReq), "grpc.request.content", "request payload must be logged in a structured way")
	attr := serverReq["attributes"].(map[string]interface{})
	assert.Contains(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
	assert.Contains(s.T(), fmt.Sprintf("%v", clientResp), "grpc.response.content", "response payload must be logged in a structured way")
	assert.Contains(s.T(), fmt.Sprintf("%v", serverResp), "grpc.response.content", "response payload must be logged in a structured way")
}

func (s *PayloadSuite) TestPingError_LogsOnlyRequestsOnError() {
	fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(s.buffer))

	_, err := s.Client.PingError(s.SimpleCtx(), &pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(4)})

	require.Error(s.T(), err, "there must be an error on an unsuccessful call")
	serverMsgs, clientMsgs := s.getServerAndClientMessages(1, 1)
	for _, m := range append(serverMsgs, clientMsgs...) {
		attr := m["attributes"].(map[string]interface{})
		assert.Equal(s.T(), "mwitkow.testproto.TestService", attr["grpc.service"], "all lines must contain service name")
		assert.Equal(s.T(), "PingError", attr["grpc.method"], "all lines must contain method name")
		assert.Equal(s.T(), "info", m["severity"], "must be logged at the info level")
	}

	clientLogs := fmt.Sprintf("%v", clientMsgs[0])
	assert.Contains(s.T(), clientLogs, "grpc.request.content", "request payload must be logged in a structured way")

	for _, m := range serverMsgs {
		serverLogs := fmt.Sprintf("%v", m)
		assert.Contains(s.T(), serverLogs, "grpc.request.content", "request payload must be logged in a structured way")
	}
}

func TestLogProtoMessageAsJson(t *testing.T) {
	t.Run("card numbers are not in log payload", func(t *testing.T) {
		b := newZRBaseSuite(t)
		b.SetupTest()
		redacted := "4564123412341234"

		fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(b.buffer))

		p := &pb_testproto.PingRequest{Value: redacted, ErrorCodeReturned: uint32(4)}
		logProtoMessageAsJson(context.Background(), b.logger, p, "key", "message")
		got := b.getOutputJSONs()
		require.NotNil(t, got)
		require.Len(t, got, 1)

		s := fmt.Sprintf("%v", got[0])
		assert.NotContains(t, s, redacted)
		attr := got[0]["attributes"].(map[string]interface{})["key"].(map[string]interface{})
		assert.Equal(t, "456412******1234", attr["value"])
	})
}
