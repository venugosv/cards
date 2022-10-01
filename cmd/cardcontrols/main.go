package main

import (
	"context"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
	"github.com/anzx/pkg/gsm"
	"github.com/anzx/pkg/log/fabriclog"

	"github.com/anzx/fabric-cards/internal/service/controls/v1beta1"
	"github.com/anzx/fabric-cards/internal/service/controls/v1beta2"
	v1beta1pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	v1beta2pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/cmd/cardcontrols/startup"
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/pkg/monitoring"
	"github.com/anzx/pkg/monitoring/names"
	"github.com/anzx/pkg/opentelemetry"

	"golang.org/x/sync/errgroup"

	"github.com/anzx/fabric-cards/cmd/cardcontrols/config"
	"github.com/anzx/fabric-cards/pkg/servers"
	"github.com/anzx/pkg/jwtauth"
)

func main() {
	ctx := context.Background()

	// Load Config
	cfg, err := config.Load()
	if err != nil {
		fatalError(ctx, err, "failed to load app config")
	}

	// Once config is passed we can set the correct log level
	fabriclog.Init(fabriclog.WithLevelStr(cfg.AppSpec.Log.Level))

	logf.Info(ctx, "Config Loaded: %v", cfg.String())

	// Set third party packages to use the updated logger functions
	jwtauth.SetLogFuncs(func(ctx context.Context, args ...interface{}) {
		logf.Info(ctx, "%v", args)
	}, func(ctx context.Context, format string, args ...interface{}) {
		logf.Info(ctx, format, args)
	})

	serverPayloadDecider := grpclogging.ServerPayloadDecider(cfg.AppSpec.Log.PayloadLoggingDecider)

	// Set features to default Gate
	if err = feature.RPCGate.Set(cfg.AppSpec.FeatureToggles.RPCs); err != nil {
		fatalError(ctx, err, "failed to initialise feature gates")
	}

	// Set features to controls Gate
	if err = feature.FeatureGate.Set(cfg.AppSpec.FeatureToggles.Features); err != nil {
		fatalError(ctx, err, "failed to initialise feature gates")
	}

	// Create JWT Auth Authenticator
	authenticator, err := jwtauth.AuthFromConfig(ctx, &cfg.AppSpec.Auth, jwtauth.DefaultHTTPClientFunc)
	if err != nil {
		fatalError(ctx, err, "failed to create authenticator client")
	}

	logf.Info(ctx, "startup: initializing monitoring")
	ctx = monitoring.Init(ctx, names.FabricCardControls)
	err = opentelemetry.Start(ctx, cfg.OpsSpec.OpenTelemetry)
	if err != nil {
		fatalError(ctx, err, "failed to start opentelemetry")
	}

	logf.Info(ctx, "startup: creating GSM client")
	gsmClient, err := gsm.NewClient(ctx)
	if err != nil {
		fatalError(ctx, err, "failed to create gsm client")
	}

	logf.Info(ctx, "startup: creating adapters")
	adapters, err := startup.NewAdapters(ctx, cfg.AppSpec, gsmClient)
	if err != nil {
		fatalError(ctx, err, "failed to create adapters")
	}

	logf.Info(ctx, "startup: creating servers")
	cardControlsV1Beta1API := v1beta1.NewServer(adapters.V1beta1.Fabric, adapters.V1beta1.Internal, adapters.V1beta1.External)
	cardControlsV1Beta2API := v1beta2.NewServer(adapters.V1beta2.Fabric, adapters.V1beta2.Internal, adapters.V1beta2.External)

	registrations := []servers.GRPCRegistration{
		func(server *grpc.Server) {
			v1beta1pb.RegisterCardControlsAPIServer(server, cardControlsV1Beta1API)
			v1beta2pb.RegisterCardControlsAPIServer(server, cardControlsV1Beta2API)
		},
	}

	restRegistrations := []servers.RestRegistration{
		v1beta1pb.RegisterCardControlsAPIHandlerFromEndpoint,
		v1beta2pb.RegisterCardControlsAPIHandlerFromEndpoint,
	}

	// Run servers and signal listener
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(startup.RunAPIServer(gCtx, cfg.AppSpec, serverPayloadDecider, registrations, restRegistrations, authenticator))
	g.Go(servers.RunOperationsServer(gCtx, cfg.AppSpec.AppName, cfg.OpsSpec.Port))
	g.Go(servers.SignalListener(gCtx))

	logf.Info(ctx, "Card Features Service terminated with error: %v", g.Wait())
}

func fatalError(ctx context.Context, err error, message string) {
	logf.Error(ctx, err, message)
	panic(err)
}
