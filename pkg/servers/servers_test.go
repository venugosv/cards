package servers

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc"

	"github.com/anzx/pkg/monitoring/names"
)

func TestGRPCServer(t *testing.T) {
	t.Run("successfully create grpc server with registrations", func(t *testing.T) {
		registrations := []GRPCRegistration{
			func(server *grpc.Server) {},
		}
		got := GRPCServer(registrations, nil)
		assert.NotNil(t, got)
	})
}

func TestHTTPServerStartupSuccess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	mux := http.NewServeMux()

	err := httpServer(ctx, mux, "test-app", 8589)

	require.NoError(t, err)
}

func TestHTTPServerStartupError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mux := http.NewServeMux()
	portNumber := 65536 // highest possible port number + 1
	expectedErr := errors.New("Could not start test-app http listener")

	err := httpServer(ctx, mux, "test-app", portNumber)

	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestRunOperationsServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := RunOperationsServer(ctx, "test-app", 8000)
	go func() {
		time.Sleep(3 * time.Second)
		cancel()
	}()
	assert.NotPanics(t, func() {
		_ = server()
	})
}

func TestServe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	err := Serve(ctx, "test-app", 8000, GRPCServer(nil, nil), http.NewServeMux())
	assert.Error(t, err, "Could not start API grpcServer")
}

func TestCreateRestServer(t *testing.T) {
	t.Run("successfully create rest api server", func(t *testing.T) {
		registrations := func(_ context.Context, _ *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
			return nil
		}
		CreateRestServer(context.Background(), 8080, nil, names.FabricCardControls, registrations)
	})
	t.Run("successfully create rest api server for cards", func(t *testing.T) {
		registrations := func(_ context.Context, _ *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
			return nil
		}
		CreateRestServer(context.Background(), 8080, nil, names.FabricCards, registrations)
	})
	t.Run("failed create rest api server, panics", func(t *testing.T) {
		registrations := func(_ context.Context, _ *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
			return fmt.Errorf("panic")
		}
		assert.Panics(t, func() {
			CreateRestServer(context.Background(), 8080, nil, names.FabricCardControls, registrations)
		})
	})
}

func TestSignalListener(t *testing.T) {
	t.Run("", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		listener := SignalListener(ctx)
		go func() {
			_ = listener()
		}()
		defer cancel()
	})
}
