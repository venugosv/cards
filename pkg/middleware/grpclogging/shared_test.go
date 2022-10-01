package grpclogging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/anzx/pkg/log"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging/ctxlg"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
)

var goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}

type loggingPingService struct {
	pb_testproto.TestServiceServer
}

func (s *loggingPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctxlg.AddFields(ctx, map[string]interface{}{"custom_field": "custom_value"})
	ctxLog := ctxlg.Extract(ctx)
	log.Info(ctx, "some ping", fromMapToAttributes(ctxLog.Fields)...)

	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *loggingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *loggingPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

type BaseSuite struct {
	*grpc_testing.InterceptorTestSuite
	buffer *bytes.Buffer
	logger *ctxlg.CtxLogger
}

func newZRBaseSuite(t *testing.T) *BaseSuite {
	b := &bytes.Buffer{}

	return &BaseSuite{
		logger: &ctxlg.CtxLogger{Fields: make(map[string]interface{})},
		buffer: b,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &loggingPingService{&grpc_testing.TestPingService{T: t}},
		},
	}
}

func (s *BaseSuite) SetupTest() {
	s.buffer.Reset()
}

func (s *BaseSuite) getOutputJSONs() []map[string]interface{} {
	ret := make([]map[string]interface{}, 0)
	dec := json.NewDecoder(s.buffer)

	for {
		var val map[string]interface{}
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.T().Fatalf("failed decoding output from Zerolog JSON: %v", err)
		}

		ret = append(ret, val)
	}

	return ret
}
