package startup

import (
	"context"
	"fmt"

	"github.com/anzx/pkg/gsm"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/anzx/fabric-cards/pkg/middleware/errors"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"

	"github.com/anzx/fabric-cards/pkg/util/jwtutil"
	"google.golang.org/grpc"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway"

	"github.com/anzx/fabric-cards/internal/service/controls/v1beta2"

	"github.com/anzx/fabric-cards/internal/service/controls/v1beta1"

	"github.com/anzx/fabric-cards/pkg/integration/ocv"

	anzerrors "github.com/anzx/pkg/errors"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"

	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/vault"

	"github.com/anzx/fabric-cards/cmd/cardcontrols/config/app"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/visa"
)

type Adapters struct {
	V1beta1 Beta1
	V1beta2 Beta2
}

type Beta1 struct {
	v1beta1.Fabric
	v1beta1.Internal
	v1beta1.External
}

type Beta2 struct {
	v1beta2.Fabric
	v1beta2.Internal
	v1beta2.External
}

func NewAdapters(ctx context.Context, config app.Spec, gsmClient *gsm.Client) (*Adapters, error) {
	var adapters Adapters

	decider := grpclogging.ClientPayloadDecider(config.Log.PayloadLoggingDecider)
	opts := getDialOptions(grpclogging.PayloadUnaryClientInterceptor(decider))

	// Fabric Adapters
	commandCentreClient, err := commandcentre.NewClient(ctx, config.CommandCentre)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure CommandCentre client with config %+v", config.CommandCentre))
	}
	adapters.V1beta1.CommandCentre = commandCentreClient
	adapters.V1beta2.CommandCentre = commandCentreClient

	eligibilityClient, err := eligibility.NewClient(ctx, config.Eligibility, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Eligibility Client with config %+v", config.Eligibility))
	}
	adapters.V1beta1.Eligibility = eligibilityClient
	adapters.V1beta2.Eligibility = eligibilityClient

	entitlementsClient, err := entitlements.NewClient(ctx, config.Entitlements, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Entitlements client with config %+v", config.Entitlements))
	}
	adapters.V1beta1.Entitlements = entitlementsClient
	adapters.V1beta2.Entitlements = entitlementsClient

	ctmClient, err := ctm.ClientFromConfig(ctx, nil, config.CTM, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure CTM Client with config %+v", config.CTM))
	}

	adapters.V1beta1.CTM = ctmClient
	adapters.V1beta2.CTM = ctmClient

	visaClient, err := visa.ClientFromConfig(ctx, nil, config.Visa, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Visa Client with config %+v", config.Visa))
	}
	adapters.V1beta1.Visa = visaClient

	forgerockClient, err := forgerock.ClientFromConfig(ctx, nil, config.Forgerock, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Forgerock Client with config %+v", config.Forgerock))
	}
	adapters.V1beta2.Forgerock = forgerockClient

	visaGatewayClient, err := visagateway.NewClient(ctx, config.VisaGateway, opts...)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Visa Gateway Clienter with config %+v", config.VisaGateway))
	} else if visaGatewayClient != nil {
		adapters.V1beta2.Visa = visaGatewayClient.CustomerRules
	}

	vaultClient, err := vault.NewClient(ctx, nil, config.Vault)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Vault Client with config %+v", config.Vault))
	}
	adapters.V1beta1.Vault = vaultClient
	adapters.V1beta2.Vault = vaultClient

	auditLogClient, err := auditlogger.NewClient(ctx, config.AuditLog)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure Auditlog Client with config %+v", config.AuditLog))
	}
	adapters.V1beta1.AuditLog = auditLogClient
	adapters.V1beta2.AuditLog = auditLogClient

	ocvClient, err := ocv.ClientFromConfig(ctx, nil, config.OCV, gsmClient)
	if err != nil {
		return nil, anzErr(err, fmt.Sprintf("could not configure OCV Client with config %+v", config.OCV))
	}
	adapters.V1beta1.OCV = ocvClient
	adapters.V1beta2.OCV = ocvClient

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
