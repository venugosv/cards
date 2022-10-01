package main

import (
	"context"

	"github.com/anzx/fabric-cards/internal/service/wallet"
	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
	"github.com/anzx/pkg/log/fabriclog"
	"github.com/anzx/pkg/monitoring/names"
	"github.com/anzx/pkg/opentelemetry"
	"golang.org/x/sync/errgroup"

	"github.com/anzx/fabric-cards/cmd/cards/config"
	"github.com/anzx/fabric-cards/cmd/cards/startup"
	"github.com/anzx/fabric-cards/internal/service/cards"
	"github.com/anzx/fabric-cards/internal/service/eligibility"
	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-cards/pkg/servers"
	"github.com/anzx/pkg/gsm"
	"github.com/anzx/pkg/jwtauth"
	"github.com/anzx/pkg/monitoring"
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
		fatalError(ctx, err, "Failed to initialise feature gates")
	}

	// Set features to replace reason Gate
	if err = feature.FeatureGate.Set(cfg.AppSpec.FeatureToggles.Features); err != nil {
		fatalError(ctx, err, "Failed to initialise feature gates")
	}

	// Create JWT Auth Authenticator
	authenticator, err := jwtauth.AuthFromConfig(ctx, &cfg.AppSpec.Auth, jwtauth.DefaultHTTPClientFunc)
	if err != nil {
		fatalError(ctx, err, "failed to create authenticator client")
	}

	logf.Info(ctx, "startup: initializing monitoring")
	ctx = monitoring.Init(ctx, names.FabricCards)
	err = opentelemetry.Start(ctx, cfg.OpsSpec.OpenTelemetry)
	if err != nil {
		fatalError(ctx, err, "failed to start opentelemetry")
	}

	logf.Info(ctx, "startup: creating GSM client")
	gsmClient, err := gsm.NewClient(ctx)
	if err != nil {
		fatalError(ctx, err, "could not configure GSM Client")
	}

	logf.Info(ctx, "startup: creating adapters")
	adapters, err := startup.NewAdapters(ctx, cfg.AppSpec, gsmClient)
	if err != nil {
		fatalError(ctx, err, "failed to create adapters")
	}

	logf.Info(ctx, "startup: creating servers")
	cardControlsAPI := cards.NewServer(adapters.Fabric, adapters.Internal, adapters.External)
	eligibilityAPI := eligibility.NewServer(adapters.Entitlements, adapters.CTM, adapters.Vault)
	walletAPI := wallet.NewServer(adapters.CTM, adapters.Vault, adapters.APCAM, adapters.Eligibility,
		adapters.Entitlements, adapters.AuditLog, adapters.GPay, adapters.SelfService)

	// Run servers and signal listener
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(startup.RunAPIServer(gCtx, cfg.AppSpec, serverPayloadDecider, authenticator,
		cardControlsAPI, eligibilityAPI, walletAPI))
	g.Go(servers.RunOperationsServer(gCtx, cfg.AppSpec.AppName, cfg.OpsSpec.Port))
	g.Go(servers.SignalListener(gCtx))

	logf.Info(ctx, "Cards Service terminated with error: %v", g.Wait())
}

func fatalError(ctx context.Context, err error, message string) {
	logf.Error(ctx, err, message)
	panic(err)
}
