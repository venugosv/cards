package grpclogging

import (
	"bytes"
	"context"
	"fmt"
	"path"

	"github.com/anzx/fabric-cards/pkg/middleware/requestid"
	"github.com/anzx/pkg/log"
	"github.com/anzx/pkg/opentelemetry"

	"github.com/anzx/fabric-cards/pkg/sanitize"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging/ctxlg"
	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/golang/protobuf/proto"  //nolint:staticcheck
	"google.golang.org/grpc"
)

// JsonPbMarshaller is the marshaller used for serializing protobuf messages.
var JsonPbMarshaller = &jsonpb.Marshaler{}

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpclogging.UnaryServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadUnaryServerInterceptor(decider ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !decider(ctx, info.FullMethod, info.Server) {
			return handler(ctx, req)
		}
		// Use the provided log.Logger for logging but use the fields from context.
		resLogger := ctxlg.CtxLogger{
			Fields: ctxlg.MergeFields(
				serverCallFields(info.FullMethod), ctxlg.TagsToFields(ctx)),
		}

		logProtoMessageAsJson(ctx, &resLogger, req, "grpc.request.content", "server request payload logged as grpc.request.content field")
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJson(ctx, &resLogger, resp, "grpc.response.content", "server response payload logged as grpc.response.content field")
		}
		return resp, err
	}
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the paylods of requests and responses.
func PayloadUnaryClientInterceptor(decider ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		logEntry := ctxlg.CtxLogger{
			Fields: newClientLoggerFields(ctx, method),
		}
		logProtoMessageAsJson(ctx, &logEntry, req, "grpc.request.content", "client request payload logged as grpc.request.content")
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJson(ctx, &logEntry, reply, "grpc.response.content", "client response payload logged as grpc.response.content")
		}
		return err
	}
}

func logProtoMessageAsJson(ctx context.Context, logger *ctxlg.CtxLogger, pbMsg interface{}, key string, msg string) {
	p, ok := pbMsg.(proto.Message)
	if !ok {
		return
	}
	payload, err := (&jsonpbObjectMarshaler{pb: p}).MarshalJSON()
	if err != nil {
		fields := append(fromMapToAttributes(logger.Fields), log.Bytes(key, payload))
		log.Error(ctx, err, msg, fields...)
	} else {
		sanitizedPayload := sanitize.ConvertToLoggableFieldValue(payload)
		fields := append(fromMapToAttributes(logger.Fields), log.Any(key, sanitizedPayload))
		log.Info(ctx, msg, fields...)
	}
}

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}

	return b.Bytes(), nil
}

func newClientLoggerFields(ctx context.Context, fullMethodString string) map[string]interface{} {
	spanDetails := opentelemetry.GetSpanDetails(ctx)

	return map[string]interface{}{
		"system":       "grpc",
		"span.kind":    "client",
		"grpc.service": path.Dir(fullMethodString)[1:],
		"grpc.method":  path.Base(fullMethodString),
		"x-request-id": requestid.FromContext(ctx),
		"x-b3-traceid": spanDetails.GetTraceID(),
	}
}
