package startup

import (
	"context"
	"os"

	"github.com/anzx/fabric-cards/pkg/middleware/errors"

	"github.com/anzx/pkg/jwtauth"
	"github.com/anzx/pkg/jwtauth/jwtgrpc"

	"github.com/anzx/pkg/auditlog"
	"github.com/anzx/pkg/monitoring/extractor"

	"github.com/anzx/fabric-cards/cmd/cardcontrols/config/app"
	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-cards/pkg/middleware/requestid"
	"github.com/anzx/fabric-cards/pkg/servers"
	anzerrors "github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/monitoring/names"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
)

func RunAPIServer(ctx context.Context, cfg app.Spec, serverPayloadDecider grpclogging.ServerPayloadLoggingDecider, registrations []servers.GRPCRegistration, restRegistrations []servers.RestRegistration, auth jwtauth.Authenticator) func() error {
	return func() error {
		interceptors := []grpc.UnaryServerInterceptor{
			extractor.MonitorGRPCServerUnaryInterceptor(),
			requestid.InjectRequestID(),
			grpclogging.UnaryServerInterceptor(),
			grpcValidator.UnaryServerInterceptor(),
			anzerrors.UnaryServerInterceptor(),
			errors.UnaryServerErrorLogInterceptor(),
			feature.APIFeatureGate(),
			jwtgrpc.UnaryServerInterceptor(auth),
			auditlog.UnaryServerInterceptor(cfg.AuditLog, os.Getenv("POD_ID")),
		}

		if serverPayloadDecider != nil {
			interceptors = append(interceptors, grpclogging.PayloadUnaryServerInterceptor(serverPayloadDecider))
		}

		grpcServer := servers.GRPCServer(registrations, interceptors)

		restServer := servers.CreateRestServer(ctx, cfg.Port, nil, names.FabricCardControls, restRegistrations...)

		return servers.Serve(ctx, cfg.AppName, cfg.Port, grpcServer, restServer)
	}
}
