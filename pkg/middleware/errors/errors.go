package errors

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc"
)

// UnaryServerErrorLogInterceptor is an interceptor that prints error
func UnaryServerErrorLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			switch err := err.(type) {
			case *anzerrors.Error:
				trace := err.StackTrace()
				traceLen := len(trace)
				if traceLen > 6 {
					traceLen = 6
				}
				logf.Error(ctx, err, "%+v", trace[0:traceLen])
			default:
				// Check if err wraps a fabric error
				logf.Error(ctx, err, "%+v", err)
			}
		}

		return resp, err
	}
}

func UnaryClientErrorLogInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			if ferr, ok := anzerrors.FromStatusError(err); ok {
				return ferr
			}
		}
		return err
	}
}
