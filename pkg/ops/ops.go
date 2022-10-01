package ops

import (
	"fmt"

	"github.com/anzx/pkg/opentelemetry"

	flag "github.com/spf13/pflag"
)

// Spec defines the configuration relevant to operation of card controls (profilers, tracing, etc)
type Spec struct {
	Port          int                   `json:"port"                 yaml:"port"                 mapstructure:"port" validate:"required,gt=0"`
	OpenTelemetry *opentelemetry.Config `json:"opentelemetry,omitempty" yaml:"opentelemetry,omitempty" mapstructure:"opentelemetry"`
}

const (
	defaultOpsPort = 8082
)

// Default returns a spec with default values.
func Default() Spec {
	return Spec{
		Port:          defaultOpsPort,
		OpenTelemetry: nil,
	}
}

const (
	opsPortHelp = "Ops port"
)

// Flags maps cli flags to configuration.
func Flags(f *flag.FlagSet, prefix string) {
	f.IntP(fmt.Sprintf("%s.port", prefix), "o", defaultOpsPort, opsPortHelp)
}
