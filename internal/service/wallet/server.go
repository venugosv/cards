package wallet

import (
	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/gpay"
	"github.com/anzx/fabric-cards/pkg/integration/selfservice"

	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"

	"github.com/anzx/fabric-cards/pkg/integration/apcam"
	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"

	"github.com/anzx/fabric-cards/pkg/integration/vault"
)

type server struct {
	cpb.UnimplementedWalletAPIServer
	ctm          ctm.Client
	vault        vault.Client
	apcam        apcam.Client
	eligibility  *eligibility.Client
	entitlements *entitlements.Client
	selfService  *selfservice.Client
	gPay         gpay.Client
	auditLog     *auditlogger.Client
}

// NewServer constructs a new CustomerRulesAPI from configured clients
func NewServer(ctm ctm.Client, vault vault.Client, apcam apcam.Client, eligibility *eligibility.Client,
	entitlements *entitlements.Client, auditlog *auditlogger.Client, gpay gpay.Client,
	selfservice *selfservice.Client,
) cpb.WalletAPIServer {
	return &server{
		ctm:          ctm,
		vault:        vault,
		apcam:        apcam,
		eligibility:  eligibility,
		entitlements: entitlements,
		auditLog:     auditlog,
		gPay:         gpay,
		selfService:  selfservice,
	}
}

const (
	serviceUnavailable     = "service unavailable"
	pushProvisioningFailed = "Push Provisioning Failed"
	apple                  = "APPLE"
	google                 = "GOOGLE"
)
