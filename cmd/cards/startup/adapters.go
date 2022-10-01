package startup

import (
	"context"
	"fmt"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"
	"github.com/anzx/fabric-cards/pkg/integration/gpay"
	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
	"github.com/anzx/pkg/log"

	"github.com/anzx/pkg/gsm"

	"github.com/anzx/fabric-cards/pkg/integration/apcam"

	"github.com/anzx/fabric-cards/pkg/middleware/errors"

	"github.com/anzx/fabric-cards/pkg/sanitize"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/anzx/fabric-cards/pkg/integration/cardcontrols"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway"
	"github.com/anzx/fabric-cards/pkg/util/jwtutil"
	anzerrors "github.com/anzx/pkg/errors"
	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/cmd/cards/config/app"
	"github.com/anzx/fabric-cards/internal/service/cards"
	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"
	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/echidna"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/ocv"
	"github.com/anzx/fabric-cards/pkg/integration/selfservice"
	"github.com/anzx/fabric-cards/pkg/integration/vault"
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
	"github.com/anzx/fabric-cards/pkg/ratelimit"
)

type Adapters struct {
	cards.Fabric
	cards.Internal
	cards.External
	APCAM apcam.Client
	GPay  gpay.Client
}

func NewAdapters(ctx context.Context, config app.Spec, gsmClient *gsm.Client) (*Adapters, error) {
	var adapters Adapters

	decider := grpclogging.ClientPayloadDecider(config.Log.PayloadLoggingDecider)
	opts := getDialOptions(grpclogging.PayloadUnaryClientInterceptor(decider))

	// Fabric Adapters
	cardControlsCLient, err := cardcontrols.NewClient(ctx, config.CardControls, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Card Controls client with env %+v", config.CardControls))
	}
	adapters.CardControls = cardControlsCLient

	commandCentreClient, err := commandcentre.NewClient(ctx, config.CommandCentre)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure CommandCentre client with config %+v", config.CommandCentre))
	}
	adapters.CommandCentre = commandCentreClient

	eligibilityClient, err := eligibility.NewClient(ctx, config.Eligibility, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Eligibility Client with config %+v", config.Eligibility))
	}
	adapters.Eligibility = eligibilityClient

	entitlementsClient, err := entitlements.NewClient(ctx, config.Entitlements, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Entitlements client with config %+v", config.Entitlements))
	}
	adapters.Entitlements = entitlementsClient

	selfServiceClient, err := selfservice.NewClient(ctx, config.SelfService, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Self Service client with config %+v", config.SelfService))
	}
	adapters.SelfService = selfServiceClient

	visaGatewayClient, err := visagateway.NewClient(ctx, config.VisaGateway, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Visa Gateway Client with config %+v", config.VisaGateway))
	} else if visaGatewayClient != nil {
		adapters.DCVV2 = visaGatewayClient.DCVV2
	}

	// Internal Adapters
	rateLimitClient, err := ratelimit.NewClient(ctx, config.RateLimit, gsmClient)
	if err != nil {
		log.Error(ctx, err, "", log.Any("config", sanitize.ConvertToLoggableFieldValue(config.RateLimit.Byte())))
		return nil, anzErr(err, "could not configure Rate Limit client")
	}
	adapters.RateLimit = rateLimitClient

	// External Adapters
	ctmClient, err := ctm.ClientFromConfig(ctx, nil, config.CTM, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure CTM Client with config %+v", config.CTM))
	}
	adapters.CTM = ctmClient

	echidnaClient, err := echidna.ClientFromConfig(ctx, nil, config.Echidna, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Echidna Client with config %+v", config.Echidna))
	}
	adapters.Echidna = echidnaClient

	vaultClient, err := vault.NewClient(ctx, nil, config.Vault)
	if err != nil {
		return nil, anzErr(err, "unable to create vault adapter")
	}
	adapters.Vault = vaultClient

	auditlogPublisher, err := auditlogger.NewClient(ctx, config.AuditLog)
	if err != nil {
		logf.Error(ctx, err, "unable to create auditlog adapter")
		return nil, anzErr(err, fmt.Sprintf("could not configure auditlog Client with config %+v", config.AuditLog))
	}
	adapters.AuditLog = auditlogPublisher

	ocvClient, err := ocv.ClientFromConfig(ctx, nil, config.OCV, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure OCV Client with config %+v", config.OCV))
	}
	adapters.OCV = ocvClient

	apcamClient, err := apcam.ClientFromConfig(ctx, nil, config.APCAM, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure APCAM Client with config %+v", config.APCAM))
	}
	adapters.APCAM = apcamClient

	forgerockClient, err := forgerock.ClientFromConfig(ctx, nil, config.Forgerock, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Forgerock Client with config %+v", config.Forgerock))
	}
	adapters.Forgerock = forgerockClient

	gPayClient, err := gpay.NewClientFromConfig(ctx, config.GPay, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure GPay Client with config %+v", config.GPay))
	}
	adapters.GPay = gPayClient

	return &adapters, nil
}

func getDialOptions(interceptors ...grpc.UnaryClientInterceptor) []grpc.DialOption {
	baseInterceptors := []grpc.UnaryClientInterceptor{
		jwtutil.OutgoingJWTInterceptor(),
		errors.UnaryClientErrorLogInterceptor(),
	}

	interceptors = append(baseInterceptors, interceptors...)

	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(interceptors...),
	}
}

func anzErr(err error, msg string) error {
	return anzerrors.Wrap(err, anzerrors.GetStatusCode(err), msg, anzerrors.GetErrorInfo(err))
}
