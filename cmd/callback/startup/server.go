package startup

import (
	"context"

	"github.com/anzx/fabric-cards/pkg/middleware/errors"
	anzerrors "github.com/anzx/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/middleware/certvalidator"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/anzx/fabric-cards/cmd/callback/config/app"
	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-cards/pkg/middleware/requestid"
	"github.com/anzx/fabric-cards/pkg/servers"
	"github.com/anzx/pkg/monitoring/extractor"
	"github.com/anzx/pkg/monitoring/names"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
)

func RunAPIServer(ctx context.Context, cfg app.Spec, serverPayloadDecider grpclogging.ServerPayloadLoggingDecider, grpcRegistrations []servers.GRPCRegistration, restRegistrations []servers.RestRegistration) func() error {
	return func() error {
		interceptors := []grpc.UnaryServerInterceptor{
			extractor.MonitorGRPCServerUnaryInterceptor(),
			requestid.InjectRequestID(),
			grpclogging.UnaryServerInterceptor(),
			anzerrors.UnaryServerInterceptor(),
			errors.UnaryServerErrorLogInterceptor(),
			grpcValidator.UnaryServerInterceptor(),
			feature.APIFeatureGate(),
		}

		if serverPayloadDecider != nil {
			interceptors = append(interceptors, grpclogging.PayloadUnaryServerInterceptor(serverPayloadDecider))
		}

		if cfg.Certificates != nil {
			certValidator, err := certvalidator.UnaryServerInterceptor(ctx, cfg.Certificates)
			if err != nil {
				return err
			}
			interceptors = append(interceptors, certValidator)
		}

		grpcServer := servers.GRPCServer(grpcRegistrations, interceptors)

		serveMux := runtime.NewServeMux(
			runtime.WithIncomingHeaderMatcher(certvalidator.IncomingHeaderMatcher),
		)

		restServer := servers.CreateRestServer(ctx, cfg.Port, serveMux, names.FabricVisaCallback, restRegistrations...)

		return servers.Serve(ctx, cfg.AppName, cfg.Port, grpcServer, restServer)
	}
}
