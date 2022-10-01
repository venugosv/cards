// Package servers provides a functionality to create and run grpc and http servers.
package servers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/anzx/pkg/opentelemetry"
	"github.com/pkg/errors"

	"github.com/anz-bank/pkg/health"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"

	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"
	otelHTTP "github.com/anzx/pkg/opentelemetry/instrumentation/http"
)

const (
	contentType     = "content-type"
	grpcContentType = "application/grpc"
	StartupFailure  = "startup Failure"
)

// GRPCRegistration is the signature for gRPC registrations against a server.
type GRPCRegistration func(server *grpc.Server)

// GRPCServer creates a gRPC server.
func GRPCServer(registrations []GRPCRegistration, interceptors []grpc.UnaryServerInterceptor) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...))

	reflection.Register(server)

	for _, register := range registrations {
		register(server)
	}

	return server
}

func RunOperationsServer(ctx context.Context, appName string, port int) func() error {
	return func() error {
		mux := http.NewServeMux()
		opentelemetry.Serve(mux)
		return httpServer(ctx, mux, appName, port)
	}
}

// httpServer runs a http server.
func httpServer(ctx context.Context, mux *http.ServeMux, appName string, port int) error {
	serverErr := make(chan error)
	defer close(serverErr)

	// Serve /healthz /readyz and /version from ops port
	if err := health.RegisterWithHTTP(mux); err != nil {
		logf.Err(ctx, err)
		return errors.Wrap(err, "cannot register HTTP health endpoints")
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Emulate the behaviour of the gRPC server which will not error when shutdown is called twice
	go func(server *http.Server, serverName string, p int, serverErr chan error) {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err != nil {
			logf.Err(ctx, err)
			serverErr <- anzerrors.Wrap(err, codes.Internal, StartupFailure, anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, fmt.Sprintf("Could not start %s http listener", serverName)))
			return
		}

		health.SetReady(true)

		err = server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			logf.Err(ctx, err)
			serverErr <- anzerrors.Wrap(err, codes.Internal, StartupFailure, anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, fmt.Sprintf("Could not start %s http server", serverName)))
			return
		}
		serverErr <- nil
	}(server, appName, port, serverErr)

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logf.Info(ctx, "gracefully shutting down %s http server", appName)
		_ = server.Shutdown(ctx)
	}

	return <-serverErr
}

type RestRegistration func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

// CreateRestServer stands up a server to handle rest requests for the API
func CreateRestServer(ctx context.Context, port int, serveMux *runtime.ServeMux, serviceName names.Service, registrations ...RestRegistration) *http.ServeMux {
	if serveMux == nil {
		serveMux = runtime.NewServeMux(
			runtime.WithIncomingHeaderMatcher(otelHTTP.TraceIncomingHeaderMatcher),
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
				Marshaler: &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames: true,
					},
					UnmarshalOptions: protojson.UnmarshalOptions{
						DiscardUnknown: true,
					},
				},
			}),
		)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(extractor.MonitorGRPCClientUnaryInterceptor(serviceName)),
	}
	addr := fmt.Sprintf(":%v", port)

	for _, registration := range registrations {
		if err := registration(ctx, serveMux, addr, dialOpts); err != nil {
			logf.Error(ctx, err, "failed to initialise rest server on port %v", port)
			panic(err)
		}
	}

	httpHandler := extractor.NewHTTPHandleFromGRPCGatewayV2(serveMux)

	mux := http.NewServeMux()
	mux.Handle("/", httpHandler)

	return mux
}

func Serve(ctx context.Context, appName string, port int, grpcServer *grpc.Server, restServer *http.ServeMux) error {
	serverErr := make(chan error)
	defer close(serverErr)

	if err := health.RegisterWithGRPC(grpcServer); err != nil {
		logf.Err(ctx, err)
		return anzerrors.Wrap(err, codes.Internal, StartupFailure, anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "cannot register gRPC health endpoints"))
	}

	if err := health.RegisterWithHTTP(restServer); err != nil {
		logf.Err(ctx, err)
		return anzerrors.Wrap(err, codes.Internal, StartupFailure, anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "cannot register HTTP health endpoints"))
	}

	server := &http.Server{
		Handler: grpcDispatcher(ctx, grpcServer, restServer),
	}

	logf.Info(ctx, "Starting %s API server on port %d", appName, port)

	go func(server *http.Server, p int, serverErr chan error) {
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err != nil {
			logf.Err(ctx, err)
			serverErr <- anzerrors.Wrap(err, codes.Internal, StartupFailure, anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "Could not start listener for API grpcServer"))
			return
		}

		health.SetReady(true)

		if err = server.Serve(listen); err != nil {
			logf.Err(ctx, err)
			serverErr <- anzerrors.Wrap(err, codes.Internal, StartupFailure, anzerrors.NewErrorInfo(ctx, anzcodes.StartupFailure, "Could not start API grpcServer"))
			return
		}
		serverErr <- nil
	}(server, port, serverErr)

	select {
	case err := <-serverErr:
		logf.Err(ctx, err)
		return err
	case <-ctx.Done():
		logf.Info(ctx, "gracefully shutting down %s gRPC server", appName)
		grpcServer.GracefulStop()
		logf.Info(ctx, "gracefully shutting down %s HTTP server", appName)
		_ = server.Shutdown(ctx)
	}

	return <-serverErr
}

func SignalListener(ctx context.Context) func() error {
	return func() error {
		sigC := make(chan os.Signal, 1)
		defer close(sigC)
		signal.Notify(sigC, syscall.SIGTERM)

		select {
		case <-sigC:
			return fmt.Errorf("terminated by SIGTERM")
		case <-ctx.Done():
			return nil
		}
	}
}
