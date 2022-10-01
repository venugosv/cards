package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/anzx/fabric-cards/pkg/middleware/logging"

	"github.com/anzx/pkg/jsontime"
	"github.com/anzx/pkg/jwtauth"
	"github.com/brehv/r"
	"github.com/mitchellh/mapstructure"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
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
		{
			description: "Auth IssuerKeyURL long",
			flags:       []string{"--spec.auth.IssuerKeyURL=localhost/IssuerKeyURL"},
			expected:    config{AppSpec: R(defaultAppSpec, "Auth.IssuerKeyURL", "localhost/IssuerKeyURL").(Spec)},
		},
		{
			description: "Auth IssuerName long",
			flags:       []string{"--spec.auth.IssuerName=IssuerName"},
			expected:    config{AppSpec: R(defaultAppSpec, "Auth.IssuerName", "IssuerName").(Spec)},
		},
	}

	for i, test := range tests {
		i := i
		test := test
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

var defaultAppSpec = Spec{
	AppName: "Cards",
	Port:    8070,
	Log: logging.Config{
		Level: "info",
	},
	Auth: jwtauth.Config{
		Issuers: []jwtauth.IssuerConfig{
			{
				Name:     "fakerock.sit.fabric.gcpnp.anz",
				JWKSURL:  "http://localhost:9080/.well-known/jwks.json",
				CacheTTL: 30 * jsontime.Duration(time.Minute),
			},
		},
	},
}
