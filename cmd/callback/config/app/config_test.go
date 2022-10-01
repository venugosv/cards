package app

import (
	"fmt"
	"testing"

	"github.com/brehv/r"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/anzx/fabric-cards/pkg/feature"
	"github.com/anzx/fabric-cards/pkg/middleware/logging"

	"github.com/anzx/pkg/jsontime"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

var R = r.R

type config struct {
	AppSpec Spec `mapstructure:"spec"`
}

func TestDefault(t *testing.T) {
	t.Parallel()

	expected := defaultAppSpec

	actual := Default()

	assert.Equal(t, expected, actual)
}

var defaultAppSpec = Spec{
	AppName: "Callback",
	Port:    8080,
	Log: logging.Config{
		Level: "info",
	},
	FeatureToggles: feature.Config{
		RPCs:     nil,
		Features: nil,
	},
}

func TestFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		flags       []string
		expected    config
	}{
		{
			description: "Default values",
			expected:    config{AppSpec: defaultAppSpec},
		},
		{
			description: "App name Short",
			flags:       []string{"-nNever"},
			expected:    config{AppSpec: R(defaultAppSpec, "AppName", "Never").(Spec)},
		},
		{
			description: "App name Long",
			flags:       []string{"--spec.appName=Never"},
			expected:    config{AppSpec: R(defaultAppSpec, "AppName", "Never").(Spec)},
		},
		{
			description: "Port Short",
			flags:       []string{"-p=9090"},
			expected:    config{AppSpec: R(defaultAppSpec, "Port", 9090).(Spec)},
		},
		{
			description: "Port Long",
			flags:       []string{"--spec.port=9090"},
			expected:    config{AppSpec: R(defaultAppSpec, "Port", 9090).(Spec)},
		},
		{
			description: "Log level short",
			flags:       []string{"-l=info"},
			expected:    config{AppSpec: R(defaultAppSpec, "Log.Level", "info").(Spec)},
		},
		{
			description: "Log level long",
			flags:       []string{"--spec.log.level=info"},
			expected:    config{AppSpec: R(defaultAppSpec, "Log.Level", "info").(Spec)},
		},
		{
			description: "Insecure Issuer short",
			flags:       []string{"-i=false"},
			expected:    config{AppSpec: R(defaultAppSpec, "InsecureIssuer", false).(Spec)},
		},
		{
			description: "Insecure Issuer short",
			flags:       []string{"--spec.InsecureIssuer=false"},
			expected:    config{AppSpec: R(defaultAppSpec, "InsecureIssuer", false).(Spec)},
		},
		{
			description: "Insecure Issuer short",
			flags:       []string{"-i=true"},
			expected:    config{AppSpec: R(defaultAppSpec, "InsecureIssuer", true).(Spec)},
		},
		{
			description: "Insecure Issuer short",
			flags:       []string{"--spec.InsecureIssuer=true"},
			expected:    config{AppSpec: R(defaultAppSpec, "InsecureIssuer", true).(Spec)},
		},
	}

	for i, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()

			var cfg config

			v := viper.New()
			v.SetDefault("spec", getMap(defaultAppSpec))
			f := flag.NewFlagSet(fmt.Sprintf("TestFlags%d", i), flag.PanicOnError)
			Flags(f, "spec")
			_ = f.Parse(test.flags)

			_ = v.BindPFlags(f)

			_ = v.Unmarshal(&cfg, viper.DecodeHook(jsontime.DurationMapstructureDecodeHookFunc))

			assert.Equal(t, test.expected, cfg)
		})
	}
}

// Map values to a struct based on mapstructure.
func getMap(encoded interface{}) map[string]interface{} {
	var inInterface map[string]interface{}
	_ = mapstructure.Decode(encoded, &inInterface)

	return inInterface
}
