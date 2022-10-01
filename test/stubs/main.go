package main

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
	"github.com/anzx/pkg/log/fabriclog"

	"github.com/anzx/fabric-cards/test/stubs/grpc/visagateway/dcvv2"

	"github.com/anzx/fabric-cards/test/stubs/http/apcam"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/anzx/pkg/jwtauth"
	"github.com/anzx/pkg/jwtauth/jwtgrpc"

	"github.com/anzx/fabric-cards/test/stubs/http/forgerock"

	"github.com/anzx/fabric-cards/test/stubs/grpc/fakerock"
	frpb "github.com/anzx/fabricapis/pkg/fabric/service/fakerock/v1alpha1"

	"github.com/anzx/fabric-visa-gateway/test/stubs/http/visa/vctc"

	"github.com/anzx/fabric-cards/test/stubs/grpc/accounts"
	"github.com/anzx/fabric-cards/test/stubs/grpc/entitlements"
	"github.com/anzx/fabric-cards/test/stubs/grpc/selfservice"
	"github.com/anzx/fabric-cards/test/stubs/grpc/visagateway/customerrules"
	apb "github.com/anzx/fabricapis/pkg/fabric/service/accounts/v1alpha6"
	entpb "github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1"
	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"
	dcvv "github.com/anzx/fabricapis/pkg/gateway/visa/service/dcvv2"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	"github.com/anzx/fabric-cards/test/stubs/http/ctm"
	"github.com/anzx/fabric-cards/test/stubs/http/echidna"
	"github.com/anzx/fabric-cards/test/stubs/http/ocv"
	"github.com/anzx/fabric-cards/test/stubs/http/vault"

	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-cards/pkg/middleware/logging/loopreader"
	"github.com/anzx/fabric-cards/test/stubs/grpc/gsm"
	"github.com/anzx/fabric-cards/test/stubs/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/anzx/pkg/log"
)

const (
	grpcPort int = 9060
	httpPort int = 9070
)

type route func(ctx context.Context, router *http.ServeMux)

type runnable func(ctx context.Context) error

func gRPCServer(ctx context.Context) error {
	logDecider := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryServerAuthInterceptor(jwtauth.InsecureAuthenticator{}),
			grpclogging.PayloadUnaryServerInterceptor(logDecider),
		),
	)

	entpb.RegisterCardEntitlementsAPIServer(grpcServer, entitlements.NewStubServer(ctx))
	entpb.RegisterEntitlementsControlAPIServer(grpcServer, entitlements.NewStubControlAPIServer(ctx))
	sspb.RegisterPartyAPIServer(grpcServer, selfservice.NewStubServer())
	apb.RegisterAccountAPIServer(grpcServer, accounts.NewStubServer(ctx))
	crpb.RegisterCustomerRulesAPIServer(grpcServer, customerrules.NewStubServer(nil))
	dcvv.RegisterDCVV2APIServer(grpcServer, dcvv2.NewStubServer(nil))
	smpb.RegisterSecretManagerServiceServer(grpcServer, gsm.NewStubServer())
	credentialspb.RegisterIAMCredentialsServer(grpcServer, vault.NewIAMServer())
	frpb.RegisterFakerockAPIServer(grpcServer, fakerock.NewStubServer())

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logf.Error(ctx, err, "gRPC server failed to listen on port %d %v", grpcPort, err)
		panic(err)
	}

	logf.Info(ctx, "GRPC Server is starting... :%d", grpcPort)

	return grpcServer.Serve(listener)
}

func httpServer(ctx context.Context) error {
	router := http.NewServeMux()
	for _, appendTo := range []route{
		utils.AppendSimulateLatencyRoute,
		ctm.AppendRoutes,
		vctc.AppendRoutes,
		echidna.AppendRoutes,
		vault.AppendRoutes,
		ocv.AppendRoutes,
		forgerock.AppendRoutes,
		apcam.AppendRoutes,
	} {
		appendTo(ctx, router)
	}

	listenAddr := fmt.Sprintf(":%d", httpPort)

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      handler(ctx, router, utils.SimulateLatencies(ctx), logging(ctx)),
		ErrorLog:     stdlog.New(os.Stdout, "http:", stdlog.LstdFlags),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logf.Info(ctx, "HTTP Server is starting... %s", listenAddr)

	return server.ListenAndServe()
}

func logging(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-request-Id")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		w.Header().Set("X-request-Id", requestID)

		log.Debug(ctx, "", log.Str("requestID", requestID), log.Str("method", r.Method), log.Str("url", r.URL.Path), log.Str("query", r.URL.RawQuery))

		if r.Method != http.MethodGet {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logf.Error(ctx, err, "failed to read body")
				return
			}
			defer r.Body.Close()

			log.Debug(ctx, "", log.Bytes("body", bodyBytes))
		}
	}
}

func handler(ctx context.Context, router *http.ServeMux, handlers ...http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		r.Body, err = loopreader.New(r.Body)
		if err != nil {
			logf.Error(ctx, err, "failed to add infinite")
			return
		}
		for _, h := range handlers {
			h.ServeHTTP(w, r)
		}
		router.ServeHTTP(w, r)
	}
}

func main() {
	var err error
	errChan := make(chan error)

	fabriclog.Init(fabriclog.WithLevel(fabriclog.DebugLevel))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Set simulated latencies if env variable LATENCY_ENABLED is set to true
	latencyEnabled := os.Getenv("LATENCY_ENABLED")
	if latencyEnabled == "true" {
		utils.GetStore(ctx).SaveSimulatedLatencies(utils.SimulatedLatencies{
			"debit-card": 300, // ctm call 300ms
			"vtcc":       200, // visa call 200ms
		})
	}

	go run(ctx, gRPCServer, errChan)
	go run(ctx, httpServer, errChan)

	err = <-errChan
	if err != nil {
		panic(err)
	}
}

func run(ctx context.Context, fn runnable, errc chan error) {
	errc <- fn(ctx)
}

// UnaryServerAuthInterceptor customised to allow SecretManagerService stub through
func UnaryServerAuthInterceptor(authenticator jwtauth.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/google.cloud.secretmanager.v1.SecretManagerService/AccessSecretVersion" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "no auth in context metadata")
		}

		jwt, err := authHeader(md.Get("authorization"))
		if err != nil {
			return nil, err
		}

		claims, err := authenticator.Authenticate(ctx, jwt)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth header")
		}

		ctxWithClaims := jwtauth.AddClaimsToContext(ctx, claims)
		ctxWithJWT := context.WithValue(ctxWithClaims, jwtgrpc.JWTKey{}, jwt)
		return handler(ctxWithJWT, req)
	}
}

func authHeader(header []string) (string, error) {
	if len(header) != 1 || len(header[0]) <= 7 {
		return "", status.Errorf(codes.Unauthenticated, "auth does not contain bearer token")
	}

	return header[0][7:], nil
}
