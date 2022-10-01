package v1beta1

import (
	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"
	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/ocv"
	"github.com/anzx/fabric-cards/pkg/integration/vault"

	"github.com/anzx/fabric-cards/pkg/integration/visa"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

type server struct {
	Fabric
	Internal
	External
	ccpb.UnimplementedCardControlsAPIServer
}

type Fabric struct {
	CommandCentre *commandcentre.Client
	Eligibility   *eligibility.Client
	Entitlements  entitlements.Carder
}

type Internal struct{}

type External struct {
	Vault    vault.Client
	CTM      ctm.ControlAPI
	Visa     visa.CustomerRulesAPI
	AuditLog *auditlogger.Client
	OCV      ocv.Client
}

// NewServer constructs a new CustomerRulesAPI from configured clients
func NewServer(fabric Fabric, internal Internal, external External) ccpb.CardControlsAPIServer {
	return &server{Fabric: fabric, Internal: internal, External: external}
}
