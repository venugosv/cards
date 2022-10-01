package jwtutil

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"google.golang.org/grpc"

	"github.com/anzx/pkg/jwtauth/jwtgrpc"
	"google.golang.org/grpc/metadata"
)

func OutgoingJWTInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(populateJWTToOutgoingContext(ctx), method, req, reply, cc, opts...)
	}
}

func populateJWTToOutgoingContext(incomingCtx context.Context) context.Context {
	token, err := jwtgrpc.GetBearerFromIncomingContext(incomingCtx)
	if err != nil {
		// The error should never happen since there is middleware guarantees the incoming requests are authenticated
		logf.Error(incomingCtx, err, "jwt: error extracting jwt from context")
		return incomingCtx
	}
	return AddJWTToOutgoingContext(incomingCtx, token)
}

func AddJWTToOutgoingContext(ctx context.Context, jwttoken string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+jwttoken)
}
