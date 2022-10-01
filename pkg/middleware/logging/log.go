package logging

import (
	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging"
)

type Config struct {
	Level                             string `json:"level"     yaml:"level"     mapstructure:"level"`
	grpclogging.PayloadLoggingDecider `json:"payloadDecider" yaml:"payloadDecider" mapstructure:"payloadDecider"`
}
