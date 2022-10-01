package cards

import (
	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"
	"github.com/anzx/fabric-cards/pkg/integration/cardcontrols"
	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/echidna"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/forgerock"
	"github.com/anzx/fabric-cards/pkg/integration/ocv"
	"github.com/anzx/fabric-cards/pkg/integration/selfservice"
	"github.com/anzx/fabric-cards/pkg/integration/vault"
	"github.com/anzx/fabric-cards/pkg/integration/visagateway/dcvv2"
	"github.com/anzx/fabric-cards/pkg/ratelimit"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
)

type server struct {
	cpb.UnimplementedCardAPIServer
	Fabric
	Internal
	External
}

type Fabric struct {
	CommandCentre *commandcentre.Client
	Eligibility   *eligibility.Client
	Entitlements  *entitlements.Client
	SelfService   *selfservice.Client
	DCVV2         *dcvv2.Client
	CardControls  *cardcontrols.Client
}

type Internal struct {
	RateLimit ratelimit.RateLimit
}

type External struct {
	CTM       ctm.Client
	Echidna   echidna.Echidna
	Vault     vault.Client
	AuditLog  *auditlogger.Client
	OCV       ocv.Client
	Forgerock forgerock.Clienter
}

// NewServer constructs a new CustomerRulesAPI from configured clients
func NewServer(fabric Fabric, internal Internal, external External) cpb.CardAPIServer {
	return &server{Fabric: fabric, Internal: internal, External: external}
}
