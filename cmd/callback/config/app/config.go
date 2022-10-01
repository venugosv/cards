package app

import (
	"fmt"

	"github.com/anzx/fabric-cards/pkg/integration/commandcentre"
	"github.com/anzx/fabric-cards/pkg/integration/vault_external"
	"github.com/anzx/fabric-cards/pkg/middleware/certvalidator"

	"github.com/anzx/fabric-cards/pkg/integration/fakerock"

	"github.com/anzx/fabric-cards/pkg/integration/forgerock"

	"github.com/anzx/fabric-cards/pkg/feature"

	"github.com/anzx/fabric-cards/pkg/middleware/logging"

	"github.com/anzx/fabric-cards/pkg/integration/ctm"
	flag "github.com/spf13/pflag"
)

type Spec struct {
	AppName        string                 `json:"appName"                      yaml:"appName"                      mapstructure:"appName"           validate:"required"`
	Port           int                    `json:"port"                         yaml:"port"                         mapstructure:"port"              validate:"gte=0"`
	Log            logging.Config         `json:"log"                          yaml:"log"                          mapstructure:"log"               validate:"required"`
	CTM            *ctm.Config            `json:"ctm,omitempty"                yaml:"ctm,omitempty"                mapstructure:"ctm"`
	Vault          *vault_external.Config `json:"vault,omitempty"              yaml:"vault,omitempty"              mapstructure:"vault"`
	CommandCentre  *commandcentre.Config  `json:"commandCentre,omitempty" yaml:"commandCentre,omitempty" mapstructure:"commandCentre"`
	FeatureToggles feature.Config         `json:"featureToggles"               yaml:"featureToggles"               mapstructure:"featureToggles"`
	Forgerock      *forgerock.Config      `json:"forgerock"                    yaml:"forgerock"                    mapstructure:"forgerock"`
	Fakerock       *fakerock.Config       `json:"fakerock"                     yaml:"fakerock"                     mapstructure:"fakerock"`
	Certificates   *certvalidator.Config  `json:"certificates"                 yaml:"certificates"                 mapstructure:"certificates"`
}

const (
	defaultAppName        = "Callback"
	defaultPort           = 8080
	defaultLogLevel       = "info"
	defaultInsecureIssuer = false
)

// Default returns a spec with default values.
func Default() Spec {
	return Spec{
		AppName: defaultAppName,
		Port:    defaultPort,
		Log: logging.Config{
			Level: defaultLogLevel,
		},
		FeatureToggles: feature.Config{},
	}
}

const (
	appHelp            = "Visa Callbacks"
	portHelp           = "Port configuration for gRPC server to listen on"
	logLevelHelp       = "Set the log level"
	insecureIssuerHelp = "Toggles whether the insecure issuer from jwttest is used, useful for test"
)

// Flags maps cli flags to configuration.
func Flags(f *flag.FlagSet, prefix string) {
	f.StringP(fmt.Sprintf("%s.appName", prefix), "n", defaultAppName, appHelp)
	f.IntP(fmt.Sprintf("%s.port", prefix), "p", defaultPort, portHelp)
	f.StringP(fmt.Sprintf("%s.log.level", prefix), "l", defaultLogLevel, logLevelHelp)
	f.BoolP(fmt.Sprintf("%s.InsecureIssuer", prefix), "i", defaultInsecureIssuer, insecureIssuerHelp)
}
