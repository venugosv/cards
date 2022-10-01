package main

import (
	"context"

	"github.com/anzx/pkg/monitoring"
	"github.com/anzx/pkg/monitoring/names"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
	ncpb "github.com/anzx/fabricapis/pkg/visa/service/notificationcallback"
	"github.com/anzx/pkg/gsm"
	"github.com/anzx/pkg/jwtauth"
	"github.com/anzx/pkg/log/fabriclog"
	"github.com/anzx/pkg/opentelemetry"

	"github.com/anzx/fabric-cards/internal/service/notificationcallback"

	"github.com/anzx/fabric-cards/internal/service/enrollmentcallback"
	ecpb "github.com/anzx/fabricapis/pkg/visa/service/enrollmentcallback"

	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/cmd/callback/config"
	"github.com/anzx/fabric-cards/cmd/callback/startup"
	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"

	"golang.org/x/sync/errgroup"

	"github.com/anzx/fabric-cards/pkg/servers"
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

	logf.Info(ctx, "startup: initializing monitoring")
	ctx = monitoring.Init(ctx, names.FabricVisaCallback)
	err = opentelemetry.Start(ctx, cfg.OpsSpec.OpenTelemetry)
	if err != nil {
		fatalError(ctx, err, "failed to start opentelemetry")
	}

	logf.Info(ctx, "startup: creating servers")
	enrollmentCallbackService := enrollmentcallback.NewServer(adapters.CTM, adapters.Vault, adapters.Fakerock, adapters.Forgerock)
	notificationCallbackService := notificationcallback.NewServer(adapters.CommandCentre)

	grpcRegistrations := []servers.GRPCRegistration{
		func(server *grpc.Server) {
			ecpb.RegisterEnrollmentCallbackAPIServer(server, enrollmentCallbackService)
			ncpb.RegisterNotificationCallbackAPIServer(server, notificationCallbackService)
		},
	}

	restRegistrations := []servers.RestRegistration{
		ecpb.RegisterEnrollmentCallbackAPIHandlerFromEndpoint,
		ncpb.RegisterNotificationCallbackAPIHandlerFromEndpoint,
	}

	// Run servers and signal listener
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(startup.RunAPIServer(gCtx, cfg.AppSpec, serverPayloadDecider, grpcRegistrations, restRegistrations))
	g.Go(servers.RunOperationsServer(gCtx, cfg.AppSpec.AppName, cfg.OpsSpec.Port))
	g.Go(servers.SignalListener(gCtx))

	logf.Info(ctx, "Callback Service terminated with error: %v", g.Wait())
}

func fatalError(ctx context.Context, err error, message string) {
	logf.Error(ctx, err, message)
	panic(err)
}
