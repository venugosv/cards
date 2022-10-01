package v1beta2

import (
	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"
	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/forgerock"
	"github.com/anzx/fabric-cards/pkg/integration/ocv"
	"github.com/anzx/fabric-cards/pkg/integration/vault"
	"github.com/anzx/fabric-cards/pkg/integration/visagateway/customerrules"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2"
)

type External struct {
	Vault     vault.Client
	CTM       ctm.ControlAPI
	AuditLog  *auditlogger.Client
	OCV       ocv.Client
	Forgerock forgerock.Clienter
}

type Fabric struct {
	CommandCentre *commandcentre.Client
	Eligibility   *eligibility.Client
	Entitlements  entitlements.Carder
	Visa          *customerrules.Client
}

type Internal struct{}

type server struct {
	Fabric
	Internal
	External
	ccpb.UnimplementedCardControlsAPIServer
}

// NewServer constructs a new CustomerRulesAPI from configured clients
func NewServer(fabric Fabric, internal Internal, external External) ccpb.CardControlsAPIServer {
	return &server{Fabric: fabric, Internal: internal, External: external}
}
