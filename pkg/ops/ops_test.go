package ops

import (
	"fmt"
	"testing"

	"github.com/anzx/pkg/jsontime"
	"github.com/brehv/r"
	"github.com/mitchellh/mapstructure"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var R = r.R

type testConfig struct {
	OpsSpec Spec `mapstructure:"ops"`
}

func TestDefault(t *testing.T) {
	t.Parallel()

	expected := defaultOpsSpec

	actual := Default()

	assert.Equal(t, expected, actual)
}

func TestFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		flags       []string
		expected    testConfig
	}{
		{
			description: "Port Short",
			flags:       []string{"-o=9090"},
			expected:    testConfig{OpsSpec: R(defaultOpsSpec, "Port", 9090).(Spec)},
		},
		{
			description: "Port Long",
			flags:       []string{"--ops.port=9090"},
			expected:    testConfig{OpsSpec: R(defaultOpsSpec, "Port", 9090).(Spec)},
		},
	}

	for i, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()

			var cfg testConfig

			v := viper.New()
			v.SetDefault("ops", getMap(defaultOpsSpec))
			f := flag.NewFlagSet(fmt.Sprintf("TestFlags%d", i), flag.PanicOnError)
			Flags(f, "ops")
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

var defaultOpsSpec = Spec{
	Port: 8082,
}
