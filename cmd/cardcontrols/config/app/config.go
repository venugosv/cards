package app

import (
	"fmt"
	"time"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/fakerock"
	"github.com/anzx/fabric-cards/pkg/integration/forgerock"
	"github.com/anzx/fabric-cards/pkg/integration/vault_external"

	"github.com/anzx/fabric-cards/pkg/integration/visagateway"

	"github.com/anzx/fabric-cards/pkg/integration/ocv"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/middleware/logging"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"
	"github.com/anzx/fabric-cards/pkg/integration/visa"
	"github.com/anzx/pkg/auditlog"
	"github.com/anzx/pkg/jsontime"
	"github.com/anzx/pkg/jwtauth"
	flag "github.com/spf13/pflag"
)

type Spec struct {
	AppName        string                 `json:"appName"                       yaml:"appName"                     mapstructure:"appName"         validate:"required"`
	Port           int                    `json:"port"                          yaml:"port"                        mapstructure:"port"            validate:"gte=0"`
	Log            logging.Config         `json:"log"                           yaml:"log"                         mapstructure:"log"             validate:"required"`
	Entitlements   *entitlements.Config   `json:"entitlements,omitempty"        yaml:"entitlements,omitempty"      mapstructure:"entitlements"`
	Eligibility    *eligibility.Config    `json:"eligibility,omitempty"         yaml:"eligibility,omitempty"       mapstructure:"eligibility"`
	Auth           jwtauth.Config         `json:"auth"                          yaml:"auth"                        mapstructure:"auth"            validate:"required_without=Insecure"` //nolint:lll
	Visa           *visa.Config           `json:"visa,omitempty"                yaml:"visa,omitempty"              mapstructure:"visa"`
	VisaGateway    *visagateway.Config    `json:"visaGateway,omitempty"         yaml:"visaGateway,omitempty"       mapstructure:"visaGateway"`
	CTM            *ctm.Config            `json:"ctm,omitempty"                 yaml:"ctm,omitempty"               mapstructure:"ctm"`
	CommandCentre  *commandcentre.Config  `json:"commandCentre,omitempty" yaml:"commandCentre,omitempty" mapstructure:"commandCentre"`
	Vault          *vault_external.Config `json:"vault,omitempty"               yaml:"vault,omitempty"             mapstructure:"vault"`
	FeatureToggles feature.Config         `json:"featureToggles"                yaml:"featureToggles"              mapstructure:"featureToggles"`
	AuditLog       *auditlog.Config       `json:"auditlog,omitempty"            yaml:"auditlog,omitempty"          mapstructure:"auditlog"`
	OCV            *ocv.Config            `json:"ocv,omitempty"                 yaml:"ocv,omitempty"               mapstructure:"ocv"`
	Forgerock      *forgerock.Config      `json:"forgerock"                     yaml:"forgerock"                   mapstructure:"forgerock"`
	Fakerock       *fakerock.Config       `json:"fakerock"                      yaml:"fakerock"                    mapstructure:"fakerock"`
}

const (
	defaultAppName          = "CardControls"
	defaultControlPort      = 8080
	defaultLogLevel         = "info"
	defaultInsecureIssuer   = false
	defaultAuthIssuerKeyURL = "http://localhost:9080/.well-known/jwks.json"
	defaultAuthIssuerName   = "fakerock.sit.fabric.gcpnp.anz"
)

// Default returns a spec with default values.
func Default() Spec {
	return Spec{
		AppName: defaultAppName,
		Port:    defaultControlPort,
		Log: logging.Config{
			Level: defaultLogLevel,
		},
		Auth: jwtauth.Config{
			Issuers: []jwtauth.IssuerConfig{
				{
					Name:     defaultAuthIssuerName,
					JWKSURL:  defaultAuthIssuerKeyURL,
					CacheTTL: 30 * jsontime.Duration(time.Minute),
				},
			},
		},
	}
}

const (
	appHelp              = "Visa Transaction Controls (VTC) provide a powerful and convenient way for cardholders to track and manage all payment activity on enrolled accounts and tokens."
	portHelp             = "Port configuration for gRPC server to listen on"
	logLevelHelp         = "Set the log level"
	insecureIssuerHelp   = "Toggles whether the insecure issuer from jwttest is used, useful for test"
	authIssuerKeyURLHelp = "Sets the subpath to the auth issuer's key to validate incoming JWTs with"
	authIssuerNameHelp   = "Sets default auth issuer name"
)

// Flags maps cli flags to configuration.
func Flags(f *flag.FlagSet, prefix string) {
	f.StringP(fmt.Sprintf("%s.appName", prefix), "n", defaultAppName, appHelp)
	f.IntP(fmt.Sprintf("%s.port", prefix), "p", defaultControlPort, portHelp)
	f.StringP(fmt.Sprintf("%s.log.level", prefix), "l", defaultLogLevel, logLevelHelp)
	f.BoolP(fmt.Sprintf("%s.InsecureIssuer", prefix), "i", defaultInsecureIssuer, insecureIssuerHelp)
	f.String(fmt.Sprintf("%s.auth.IssuerKeyURL", prefix), defaultAuthIssuerKeyURL, authIssuerKeyURLHelp)
	f.String(fmt.Sprintf("%s.auth.IssuerName", prefix), defaultAuthIssuerName, authIssuerNameHelp)
}
