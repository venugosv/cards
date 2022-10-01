package config

import (
	"time"

	"github.com/anzx/anzdata"

	"gopkg.in/yaml.v2"

	"github.com/anzx/fabric-cards/pkg/integration/vault_external"
	"github.com/anzx/utils/forgejwt/v2"
)

type Config struct {
	Timeout       time.Duration          `yaml:"timeout"       envconfig:"TIMEOUT"    default:"5s"` // TEST_TIMEOUT
	Callback      Callback               `yaml:"callback"      envconfig:"CALLBACK"`                // TEST_CALLBACK
	Cards         Service                `yaml:"cards"         envconfig:"CARDS"`                   // TEST_CARDS
	CardControls  Service                `yaml:"cardcontrols"  envconfig:"CARDCONTROLS"`            // TEST_CARDCONTROLS
	CommandCentre CommandCentre          `yaml:"commandcentre" envconfig:"COMMANDCENTRE"`           // TEST_COMMANDCENTRE
	Vault         *vault_external.Config `yaml:"vault"         envconfig:"VAULT"`                   // TEST_VAULT
	MaxUser       int                    `yaml:"maxuser"       envconfig:"MAXUSER"`                 // TEST_PNV
}

type CommandCentre struct {
	ProjectID    string `yaml:"projectid"    envconfig:"PROJECTID"`             // TEST_COMMANDCENTRE_PROJECTID
	Topic        string `yaml:"topic"        envconfig:"TOPIC"`                 // TEST_COMMANDCENTRE_TOPIC
	EmulatorHost string `yaml:"emulatorHost" envconfig:"SPANNER_EMULATOR_HOST"` // TEST_COMMANDCENTRE_EMULATORHOST
	Subscription string `yaml:"subscription" envconfig:"SUBSCRIPTION"`          // TEST_COMMANDCENTRE_SUBSCRIPTION
}

type Service struct {
	Insecure bool              `yaml:"insecure"   envconfig:"INSECURE"`                                          // TEST_(CALLBACK|CARDS|CARDCONTROLS)_INSECURE
	BaseURL  string            `yaml:"baseUrl"    envconfig:"BASEURL"    validate:"required"    required:"true"` // TEST_(CALLBACK|CARDS|CARDCONTROLS)_BASEURL
	Auth     AuthConfig        `yaml:"auth"       envconfig:"AUTH"       validate:"required"    required:"true"` // TEST_(CALLBACK|CARDS|CARDCONTROLS)_AUTH
	Headers  []string          `yaml:"headers"    envconfig:"HEADERS"`                                           // TEST_(CALLBACK|CARDS|CARDCONTROLS)_HEADERS
	TearDown bool              `yaml:"tearDown"   envconfig:"TEARDOWN"`                                          // TEST_(CALLBACK|CARDS|CARDCONTROLS)_TEARDOWN
	Toggle   map[TestName]bool `yaml:"toggle"     envconfig:"TOGGLE"`                                            // TEST_(CALLBACK|CARDS|CARDCONTROLS)_TOGGLE
	Scheme   string
}

type Callback struct {
	CurrentCard   string        `yaml:"currentCard"   envconfig:"CURRENT_CARD"    validate:"required"    required:"true"` // TEST_CALLBACK_CURRENT_CARD
	PubsubTimeout time.Duration `yaml:"pubsubTimeout" envconfig:"PUBSUB_TIMEOUT"`
	PubsubSkip    bool          `yaml:"pubsubSkip"    envconfig:"PUBSUB_SKIP"`
	Service
}

type AuthConfig struct {
	Env       forgejwt.Env   `yaml:"env"        envconfig:"ENV"        validate:"required"            required:"true"`                     // TEST_CARDS_AUTH_ENV
	Region    anzdata.Region `yaml:"region"     envconfig:"REGION"`                                                                        // TEST_REGION
	Method    AuthMethod     `yaml:"method"     envconfig:"METHOD"     validate:"oneof=fakejwt basic forgejwt forgesso" default:"fakejwt"` // TEST_CARDS_AUTH_METHOD
	PersonaID string         `yaml:"personaID"  envconfig:"PERSONA_ID"`                                                                    // TEST_CARDS_AUTH_PERSONA_ID
	FromPool  bool           `yaml:"frompool"   envconfig:"FROMPOOL"`                                                                      // TEST_CARDS
}

type AuthMethod string

const (
	AuthMethodFakeJWT   AuthMethod = "fakejwt"
	AuthMethodCustomJWT AuthMethod = "customjwt"
	AuthMethodBasic     AuthMethod = "basic"
	AuthMethodForgejwt  AuthMethod = "forgejwt"
	AuthMethodForgesso  AuthMethod = "forgesso"
	AuthMethodNone      AuthMethod = "none"
)

func (c *Config) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}
