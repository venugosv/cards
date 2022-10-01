package requestid

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

const defaultRequestID = "cddf15f6-d2ff-11ea-8179-acde48001122"

func TestFromIncomingContextMeta(t *testing.T) {
	t.Run("successfully read x-request-id from meta", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", defaultRequestID))
		id := FromIncomingContextMeta(ctx)
		assert.Equal(t, defaultRequestID, id)
	})

	t.Run("return empty string if no x-request-id from meta", func(t *testing.T) {
		ctx := context.Background()
		id := FromIncomingContextMeta(ctx)
		assert.Equal(t, "", id)
	})
}

func TestFromContext(t *testing.T) {
	t.Run("successfully read x-request-id from meta", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", defaultRequestID))
		ctx = AddToContext(ctx, defaultRequestID)
		id := FromContext(ctx)
		assert.Equal(t, defaultRequestID, id)
	})

	t.Run("return empty string if no x-request-id", func(t *testing.T) {
		ctx := context.Background()
		id := FromContext(ctx)
		assert.Equal(t, id, "")
	})

	t.Run("return empty string if x-request-id is an empty string", func(t *testing.T) {
		ctx := context.Background()
		mdmap := map[string]string{
			MetaXRequestIDKey: "",
		}
		md := metadata.New(mdmap)
		metadata.NewIncomingContext(ctx, md)
		id := FromContext(ctx)
		assert.Equal(t, id, "")
	})
}

func TestAppendToOutgoingContext(t *testing.T) {
	t.Run("successfully add x-request-id to outgoing meta", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", defaultRequestID))
		ctx = appendToOutgoingContext(ctx, defaultRequestID)
		md, _ := metadata.FromOutgoingContext(ctx)
		assert.Equal(t, defaultRequestID, md[MetaXRequestIDKey][0])
	})
}

func TestInjectRequestID(t *testing.T) {
	t.Run("successfully add x-request-id to context", func(t *testing.T) {
		fn := InjectRequestID()
		called := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			id := FromContext(ctx)
			assert.Equal(t, defaultRequestID, id)
			called = true
			return "", nil
		}
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", defaultRequestID))
		_, _ = fn(ctx, nil, nil, handler)
		assert.Equal(t, true, called)
	})

	t.Run("successfully add x-request-id to context if x-request-id is empty", func(t *testing.T) {
		fn := InjectRequestID()
		called := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			id := FromContext(ctx)
			assert.NotEmpty(t, id)
			called = true
			return "", nil
		}
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", ""))
		_, _ = fn(ctx, nil, nil, handler)
		assert.Equal(t, true, called)
	})

	t.Run("successfully add x-request-id to context if no x-request-id in incoming context", func(t *testing.T) {
		fn := InjectRequestID()
		called := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			id := FromContext(ctx)
			assert.NotEmpty(t, id)
			called = true
			return "", nil
		}
		ctx := context.Background()
		_, _ = fn(ctx, nil, nil, handler)
		assert.Equal(t, true, called)
	})
}
