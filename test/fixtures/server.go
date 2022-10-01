package fixtures

import (
	"context"
	"time"

	"github.com/anzx/pkg/log/fabriclog"

	"github.com/anzx/fabric-cards/test/stubs/pkg/gpay"

	"github.com/anzx/fabric-cards/test/stubs/http/apcam"

	crpb "github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules"

	"github.com/anzx/fabric-cards/pkg/integration/ocv"

	"github.com/anzx/fabric-cards/test/stubs/grpc/cardcontrols"
	"github.com/anzx/fabric-cards/test/stubs/grpc/fakerock"

	"github.com/anzx/fabric-cards/test/stubs/grpc/visagateway/dcvv2"

	"github.com/anzx/fabric-cards/test/stubs/grpc/visagateway/customerrules"

	"github.com/anzx/fabric-cards/pkg/integration/echidna"
	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"google.golang.org/grpc/codes"

	"github.com/anzx/fabric-cards/test/stubs/http/forgerock"

	"github.com/anzx/fabric-cards/test/util"
	"github.com/anzx/pkg/auditlog/auditlogtest"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/stubs/grpc/accounts"
	auditLogStub "github.com/anzx/fabric-cards/test/stubs/grpc/auditlog"
	commandCentreStub "github.com/anzx/fabric-cards/test/stubs/grpc/commandcentre"
	"github.com/anzx/fabric-cards/test/stubs/grpc/eligibility"
	"github.com/anzx/fabric-cards/test/stubs/grpc/entitlements"
	"github.com/anzx/fabric-cards/test/stubs/grpc/selfservice"
	ctmStub "github.com/anzx/fabric-cards/test/stubs/http/ctm"
	echidnaStub "github.com/anzx/fabric-cards/test/stubs/http/echidna"
	ocvStub "github.com/anzx/fabric-cards/test/stubs/http/ocv"
	vaultStub "github.com/anzx/fabric-cards/test/stubs/http/vault"
	visaStub "github.com/anzx/fabric-cards/test/stubs/http/visa"
	rateLimitStub "github.com/anzx/fabric-cards/test/stubs/pkg/ratelimit"

	"github.com/google/uuid"
	"gopkg.in/square/go-jose.v2/jwt"

	ccv1beta1pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
	ccv1beta2pb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
	sspb "github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2"
	"github.com/anzx/pkg/jwtauth"
	"google.golang.org/grpc/metadata"
)

const Issuer = "fakerock.sit.fabric.gcpnp.anz"

type ServerBuilder struct {
	CardEntitlementsAPIClient    entitlements.StubClient
	EntitlementsControlAPIClient entitlements.StubControlAPIClient
	CardEligibilityAPIClient     eligibility.StubClient
	CTMClient                    ctmStub.StubClient
	VaultClient                  vaultStub.StubClient
	VisaClient                   visaStub.StubClient
	CustomerRulesClient          customerrules.StubClient
	DCVV2Client                  dcvv2.StubClient
	CommandCentreEnv             commandCentreStub.StubClient
	EchidnaClient                echidnaStub.StubClient
	RateLimit                    rateLimitStub.StubClient
	SelfServiceClient            selfservice.StubClient
	AuditLogPublisher            auditLogStub.StubClient
	AccountsClient               accounts.StubClient
	OCVClient                    ocvStub.StubClient
	ForgerockClient              forgerock.StubClient
	FakerockClient               fakerock.StubClient
	CardControlsClient           cardcontrols.StubClient
	APCAMClient                  apcam.StubClient
	GPayClient                   gpay.StubClient
}

func AServer() *ServerBuilder {
	return &ServerBuilder{}
}

func (c *ServerBuilder) WithData(users ...*data.User) *ServerBuilder {
	testData := &data.Data{Users: users}
	return &ServerBuilder{
		CardEntitlementsAPIClient:    entitlements.NewStubClient(testData),
		EntitlementsControlAPIClient: entitlements.NewControlAPIStubClient(),
		CardEligibilityAPIClient:     eligibility.NewStubClient(testData),
		CTMClient:                    ctmStub.NewStubClient(testData),
		VaultClient:                  vaultStub.NewVaultClient(testData),
		VisaClient:                   visaStub.NewStubClient(testData),
		CustomerRulesClient:          customerrules.NewStubClient(testData),
		DCVV2Client:                  dcvv2.NewStubClient(testData),
		EchidnaClient:                echidnaStub.NewStubClient(testData),
		RateLimit:                    rateLimitStub.NewStubClient(),
		AuditLogPublisher:            auditLogStub.NewStubClient(),
		OCVClient:                    ocvStub.NewStubClient(),
		ForgerockClient:              forgerock.NewStubClient(),
		FakerockClient:               fakerock.NewStubClient(),
		CardControlsClient:           cardcontrols.NewStubClient(),
		APCAMClient:                  apcam.NewStubClient(),
		GPayClient:                   gpay.NewStubClient(),
	}
}

func (c *ServerBuilder) WithEntMayError(err error) *ServerBuilder {
	c.CardEntitlementsAPIClient.GetEntitledCardErr = err
	return c
}

func (c *ServerBuilder) WithEntListError(err error) *ServerBuilder {
	c.CardEntitlementsAPIClient.ListEntitledCardsErr = err
	return c
}

func (c *ServerBuilder) WithEligibilityError() *ServerBuilder {
	c.CardEligibilityAPIClient.CanErr = anzerrors.New(codes.PermissionDenied, "eligibility failed", anzerrors.NewErrorInfo(context.Background(), anzcodes.CardIneligible, "card not eligible"))
	return c
}

func (c *ServerBuilder) WithCtmActivateError(err error) *ServerBuilder {
	c.CTMClient.ActivateError = err
	return c
}

func (c *ServerBuilder) WithCtmInquiryError(err error) *ServerBuilder {
	c.CTMClient.InquiryError = err
	return c
}

func (c *ServerBuilder) WithCtmInquiryErrorFunc(f func() error) *ServerBuilder {
	c.CTMClient.InquiryErrorFunc = f
	return c
}

func (c *ServerBuilder) WithCtmReplaceError(err error) *ServerBuilder {
	c.CTMClient.ReplaceError = err
	return c
}

func (c *ServerBuilder) WithCtmUpdateError(err error) *ServerBuilder {
	c.CTMClient.UpdateError = err
	return c
}

func (c *ServerBuilder) WithCtmSetPreferenceError(err error) *ServerBuilder {
	c.CTMClient.SetPreferenceError = err
	return c
}

func (c *ServerBuilder) WithCtmPINInfoUpdateError(err error) *ServerBuilder {
	c.CTMClient.PINInfoUpdateError = err
	return c
}

func (c *ServerBuilder) WithVaultError(err error) *ServerBuilder {
	c.VaultClient.Err = err
	return c
}

func (c *ServerBuilder) WithVtcQueryError(err error) *ServerBuilder {
	c.VisaClient.QueryError = err
	return c
}

func (c *ServerBuilder) WithVtcCreateError(err error) *ServerBuilder {
	c.VisaClient.CreateError = err
	return c
}

func (c *ServerBuilder) WithVtcEnrolError(err error) *ServerBuilder {
	c.VisaClient.EnrolError = err
	return c
}

func (c *ServerBuilder) WithVtcUpdateError(err error) *ServerBuilder {
	c.VisaClient.UpdateError = err
	return c
}

func (c *ServerBuilder) WithVtcControls(controlTypes ...ccv1beta1pb.ControlType) *ServerBuilder {
	c.VisaClient.Controls = controlTypes
	return c
}

func (c *ServerBuilder) WithVtcReplaceError(err error) *ServerBuilder {
	c.VisaClient.ReplaceError = err
	return c
}

func (c *ServerBuilder) WithEchidnaErrorCode(e int) *ServerBuilder {
	c.EchidnaClient.Err = anzerrors.New(echidna.GetGRPCError(e), "failed request",
		anzerrors.NewErrorInfo(context.Background(), echidna.GetANZError(e), echidna.GetErrorMsg(e)))
	return c
}

func (c *ServerBuilder) WithRateLimitError(err error) *ServerBuilder {
	c.RateLimit.Err = anzerrors.Wrap(err, codes.ResourceExhausted, "rate limit failed",
		anzerrors.NewErrorInfo(context.Background(), anzcodes.RateLimitExhausted, err.Error()))
	return c
}

func (c *ServerBuilder) WithSelfServiceError(err error) *ServerBuilder {
	c.SelfServiceClient.GetPartyError = err
	return c
}

func (c *ServerBuilder) WithSelfServiceResponse(response *sspb.GetPartyResponse) *ServerBuilder {
	c.SelfServiceClient.GetPartyResponse = response
	return c
}

func (c *ServerBuilder) WithAuditLogError(err error) *ServerBuilder {
	c.AuditLogPublisher.Err = err
	return c
}

func (c *ServerBuilder) WithAuditLogHook(hook auditLogStub.AuditLogHook) *ServerBuilder {
	c.AuditLogPublisher.Hook = hook
	return c
}

func (c *ServerBuilder) WithOCVRetrievePartyError(err error) *ServerBuilder {
	c.OCVClient.RetrievePartyErr = err
	return c
}

func (c *ServerBuilder) WithOCVRetrievePartyResp(fn ...func([]*ocv.RetrievePartyRs)) *ServerBuilder {
	c.OCVClient.RetrievePartyRespFunc = fn
	return c
}

func (c *ServerBuilder) WithOCVMaintainContractError(err error) *ServerBuilder {
	c.OCVClient.MaintainContractErr = err
	return c
}

func (c *ServerBuilder) WithEntitlementsRegisterCardToPersonaErr(err error) *ServerBuilder {
	c.EntitlementsControlAPIClient.RegisterCardToPersonaErr = err
	return c
}

func (c *ServerBuilder) WithEntitlementsForcePartyToLatestErr(err error) *ServerBuilder {
	c.EntitlementsControlAPIClient.ForcePartyToLatestErr = err
	return c
}

func (c *ServerBuilder) WithVisaGatewayListError(err error) *ServerBuilder {
	c.CustomerRulesClient.ListError = err
	return c
}

func (c *ServerBuilder) WithVisaGatewayCreateError(err error) *ServerBuilder {
	c.CustomerRulesClient.CreateError = err
	return c
}

func (c *ServerBuilder) WithVisaGatewayRegistrationError(err error) *ServerBuilder {
	c.CustomerRulesClient.RegistrationError = err
	return c
}

func (c *ServerBuilder) WithVisaGatewayUpdateError(err error) *ServerBuilder {
	c.CustomerRulesClient.UpdateError = err
	return c
}

func (c *ServerBuilder) WithVisaGatewayDeleteError(err error) *ServerBuilder {
	c.CustomerRulesClient.DeleteError = err
	return c
}

func (c *ServerBuilder) WithVisaGatewayResource(resource *crpb.Resource) *ServerBuilder {
	c.CustomerRulesClient.FixedResponse = resource
	return c
}

func (c *ServerBuilder) WithVisaGatewayControls(controlTypes ...ccv1beta2pb.ControlType) *ServerBuilder {
	c.CustomerRulesClient.CustomerRulesAPIServer.Controls = controlTypes
	return c
}

func (c *ServerBuilder) WithVisaGatewayGamblingImpulse(start, remaining string) *ServerBuilder {
	c.CustomerRulesClient.CustomerRulesAPIServer.GamblingImpulseDelayStart = start
	c.CustomerRulesClient.CustomerRulesAPIServer.GamblingImpulseDelayRemaining = remaining
	return c
}

func (c *ServerBuilder) WithVisaGatewayReplaceError(err error) *ServerBuilder {
	c.CustomerRulesClient.ReplaceError = err
	return c
}

func (c *ServerBuilder) WithDCVV2GenerateError(err error) *ServerBuilder {
	c.DCVV2Client.GenerateErr = err
	return c
}

func (c *ServerBuilder) WithForgerockError(err error) *ServerBuilder {
	c.ForgerockClient.Err = err
	return c
}

func (c *ServerBuilder) WithFakerockError(err error) *ServerBuilder {
	c.FakerockClient.Err = err
	return c
}

func (c *ServerBuilder) WithCardControlsTransferControlsError(err error) *ServerBuilder {
	c.CardControlsClient.TransferControlsErr = err
	return c
}

func (c *ServerBuilder) WithPushProvisionError(err error) *ServerBuilder {
	c.APCAMClient.Err = err
	return c
}

func (c *ServerBuilder) WithGPayError(err error) *ServerBuilder {
	c.GPayClient.Err = err
	return c
}

func GetTestContext() context.Context {
	return GetTestContextWithJWT(data.DefaultUser().PersonaID)
}

func GetTestContextWithJWT(personaID string) context.Context {
	const staticJWT = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImFwaXNpdHRva2VuLmNvcnAuZGV2LmFueiJ9.eyJpc3MiOiJodHRwczovL2RhdGFwb3dlci1zdHMuYW56LmNvbSIsImF1ZCI6ImF1ZHBjbGllbnQwMi5kZXYuYW56Iiwic3ViIjoiYXVkcGNsaWVudDAyLmRldi5hbnoiLCJleHAiOjE1NzgyNTIwMzcuNzUzLCJzY29wZXMiOlsiQVUuUkVUQUlMLkFDQ09VTlQuUFJPRklMRS5SRUFEIl0sImFtciI6WyJwb3AiXSwiYWNyIjoiSUFMMi5BQUwxLkZBTDEifQ.HiSM1dlHwJWpb4sPE7hSriX8nekh8lNV-MnaDE4RL3mrXGHyOBrlQfa3D13Rb_PDBNdbfqzm79E6ajVVIz5U-2G2CCy1CzT1TuiVlBcyd25HJl4JhiBAKcn4aOAwRbnMp88KLYjVbGdEg4egWhfsaPdBBTEX1M5G0KWfBHAfDA5Lesq5dkSTVRGlun0Q9MhpaZSmEI6FYKt-YDEe7wMifjsEFeDF9a_H8qyyYazopFMv0XM6aIjW000nk-XFzRhBYvznwm_LzafQCVGF5tULOp5jYVnv4d7W1GnH2THMnLtC9WtgQYdQOX1eZlK4QrqsLBXrWotM9v4fy8KP06V5lg"

	ctx := context.Background()
	claims := jwtauth.NewClaims(
		jwtauth.BaseClaims{
			Claims: jwt.Claims{
				Issuer:   Issuer,
				Subject:  uuid.New().String(),
				Expiry:   jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
				IssuedAt: jwt.NewNumericDate(time.Now()),
			},
			Persona: &jwtauth.Persona{
				PersonaID: personaID,
			},
			OCVID:            "OCVID",
			AuthContextClass: "acr",
		})
	// for unit tests, we can ignore not ending the span. (3rd value returned from below function.)
	ctxWithClaims, _ := auditlogtest.NewAuditLogContext(ctx, claims)
	return metadata.NewIncomingContext(ctxWithClaims, metadata.Pairs("authorization", "Bearer "+staticJWT))
}

func GetTestContextWithLogger(personaID *string) (context.Context, *util.SyncBuffer) {
	var ctx context.Context
	if personaID != nil {
		ctx = GetTestContextWithJWT(*personaID)
	} else {
		ctx = GetTestContext()
	}
	b := util.NewSyncBuffer()
	fabriclog.Init(fabriclog.WithLevel(fabriclog.InfoLevel), fabriclog.WithConsoleWriter(b))

	return ctx, b
}
