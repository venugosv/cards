package requestid

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	MetaXRequestIDKey = "x-request-id"
)

type RequestIDKey struct{}

func InjectRequestID() grpc.UnaryServerInterceptor {
	// RequestID middleware does three things
	// 1. Read x-request-id from incoming context metadata, if it does not exist, create a new one
	// 2. Add the x-request-id to context for easy access, so later on it can be used to add to logs or http request header
	// 3. Append the x-request-id to outgoing context, so the x-request-id is propagated to the downstream APIs
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Reads the x-request-id from the incoming request
		id := FromIncomingContextMeta(ctx)
		if len(id) == 0 {
			id = newRequestID()
		}
		// Add a the request id to context for easy access
		ctx = AddToContext(ctx, id)
		// Append the request id to the outgoing context so it is propogated to downstream services
		ctx = appendToOutgoingContext(ctx, id)
		return handler(ctx, req)
	}
}

// FromIncomingContextMeta returns a request ID from gRPC metadata if available in ctx.
func FromIncomingContextMeta(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	header, ok := md[MetaXRequestIDKey]
	if !ok || len(header) == 0 {
		return ""
	}

	return header[0]
}

// FromContext returns a request ID from gRPC metadata if available in ctx.
func FromContext(ctx context.Context) string {
	r, ok := ctx.Value(RequestIDKey{}).(string)
	if !ok {
		r = ""
	}
	return r
}

func AddToContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey{}, id)
}

func appendToOutgoingContext(ctx context.Context, id string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, MetaXRequestIDKey, id)
}

func newRequestID() string {
	requestID, _ := uuid.NewUUID()
	return requestID.String()
}
