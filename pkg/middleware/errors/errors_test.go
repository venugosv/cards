package errors

import (
	"context"
	"fmt"
	"testing"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/stretchr/testify/require"
)

func TestUnaryServerErrorLogInterceptor(t *testing.T) {
	tests := []struct {
		name    string
		handler grpc.UnaryHandler
	}{
		{
			name: "anzerror",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, anzerrors.New(codes.Unimplemented, "message",
					anzerrors.NewErrorInfo(context.Background(), anzcodes.Unknown, "reason"))
			},
		},
		{
			name: "status error",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, status.Error(codes.Unknown, "message")
			},
		},
		{
			name: "std error",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, fmt.Errorf("message")
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			interceptor := UnaryServerErrorLogInterceptor()
			got, err := interceptor(context.Background(), nil, nil, test.handler)
			require.Error(t, err)
			assert.Nil(t, got)
		})
	}
}

func TestUnaryClientErrorLogInterceptor(t *testing.T) {
	t.Run("FabricError", func(t *testing.T) {
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return anzerrors.New(codes.Internal, "message", anzerrors.NewErrorInfo(context.Background(), anzcodes.Unknown, "reason"))
		}
		err := UnaryClientErrorLogInterceptor()(context.Background(), "method", nil, nil, nil, invoker)
		require.Error(t, err)
		assert.IsType(t, &anzerrors.Error{}, err)
		assert.Equal(t, anzerrors.GetMessage(err), "message")
	})
	t.Run("NotAFabricError", func(t *testing.T) {
		msg := "not a fabric error"
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return fmt.Errorf(msg)
		}
		err := UnaryClientErrorLogInterceptor()(context.Background(), "method", nil, nil, nil, invoker)
		require.Error(t, err)
		require.EqualError(t, err, msg)
	})
}
