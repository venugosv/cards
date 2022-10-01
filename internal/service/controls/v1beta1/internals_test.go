// This is the test file for all the pkg methods in the cardcontrols service layer
package v1beta1

import (
	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"
	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/test/fixtures"
	ccpb "github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1"
)

func buildCardControlsServer(c *fixtures.ServerBuilder) ccpb.CardControlsAPIServer {
	fabric := Fabric{
		CommandCentre: &commandcentre.Client{
			Publisher: c.CommandCentreEnv,
		},
		Eligibility: &eligibility.Client{
			CardEligibilityAPIClient: c.CardEligibilityAPIClient,
		},
		Entitlements: entitlements.Client{
			CardEntitlementsAPIClient: c.CardEntitlementsAPIClient,
		},
	}
	internal := Internal{}
	external := External{
		Vault: c.VaultClient,
		CTM:   c.CTMClient,
		Visa:  c.VisaClient,
		AuditLog: &auditlogger.Client{
			Publisher: c.AuditLogPublisher,
		},
		OCV: c.OCVClient,
	}
	return NewServer(fabric, internal, external)
}
