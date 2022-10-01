package startup

import (
	"context"
	"os"

	"github.com/anzx/fabric-cards/pkg/middleware/errors"

	"github.com/anzx/pkg/jwtauth/jwtgrpc"

	"github.com/anzx/fabric-cards/cmd/cards/config/app"
	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-cards/pkg/middleware/requestid"
	"github.com/anzx/fabric-cards/pkg/servers"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epb "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"github.com/anzx/pkg/auditlog"
	anzerrors "github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/jwtauth"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
)

func RunAPIServer(ctx context.Context, cfg app.Spec,
	serverPayloadDecider grpclogging.ServerPayloadLoggingDecider, authenticator jwtauth.Authenticator,
	cardsAPI cpb.CardAPIServer, eligibilityAPI epb.CardEligibilityAPIServer, walletAPI cpb.WalletAPIServer,
) func() error {
	return func() error {
		interceptors := []grpc.UnaryServerInterceptor{
			extractor.MonitorGRPCServerUnaryInterceptor(),
			requestid.InjectRequestID(),
			grpclogging.UnaryServerInterceptor(),
			grpcValidator.UnaryServerInterceptor(),
			anzerrors.UnaryServerInterceptor(),
			errors.UnaryServerErrorLogInterceptor(),
			feature.APIFeatureGate(),
			jwtgrpc.UnaryServerInterceptor(authenticator),
			auditlog.UnaryServerInterceptor(cfg.AuditLog, os.Getenv("POD_ID")),
		}

		if serverPayloadDecider != nil {
			interceptors = append(interceptors, grpclogging.PayloadUnaryServerInterceptor(serverPayloadDecider))
		}

		registrations := []servers.GRPCRegistration{
			func(server *grpc.Server) {
				cpb.RegisterCardAPIServer(server, cardsAPI)
				epb.RegisterCardEligibilityAPIServer(server, eligibilityAPI)
				cpb.RegisterWalletAPIServer(server, walletAPI)
			},
		}

		grpcServer := servers.GRPCServer(registrations, interceptors)

		restServer := servers.CreateRestServer(ctx, cfg.Port, nil, names.FabricCards, cpb.RegisterCardAPIHandlerFromEndpoint)

		return servers.Serve(ctx, cfg.AppName, cfg.Port, grpcServer, restServer)
	}
}
