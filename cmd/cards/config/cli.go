package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/anzx/fabric-cards/cmd/cards/config/app"
	"github.com/anzx/fabric-cards/pkg/ops"
	"github.com/anzx/pkg/jsontime"
	"github.com/anzx/pkg/validator"
	"github.com/mitchellh/mapstructure"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Config defines the configuration structure for
type Config struct {
	AppSpec app.Spec `json:"spec"          yaml:"spec"           mapstructure:"spec"  validate:"required,dive"`
	OpsSpec ops.Spec `json:"ops,omitempty" yaml:"ops,omitempty"  mapstructure:"ops"   validate:"required,dive"`
}

const (
	appSpec = "spec"
	opsSpec = "ops"
)

// Load configuration from file, env and flags and return compiled and validated config.
func Load() (*Config, error) {
	v := viper.New()
	f := flag.CommandLine

	flags(f)

	flag.Parse()

	config, err := create(v, f)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Create validated configuration from file, env, and flags.
func create(v *viper.Viper, f *flag.FlagSet) (*Config, error) {
	var config Config

	v.SetDefault(appSpec, getMap(app.Default()))
	v.SetDefault(opsSpec, getMap(ops.Default()))
	configFile, _ := f.GetString("config")

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigType("yaml")
		v.SetConfigName("config")
		v.AddConfigPath("/config")
		v.AddConfigPath("/app/config")
	}

	// No error will be returned by BindPFlags(), it will continue/panic/exit if something goes wrong
	_ = v.BindPFlags(f)

	// Merge file, env and flag configuration into config object. Error is intentionally ignored
	// because it's likely config file not found and the default config will be used instead
	_ = v.ReadInConfig()

	err := v.Unmarshal(&config, viper.DecodeHook(jsontime.DurationMapstructureDecodeHookFunc))
	if err != nil {
		log.Println(fmt.Errorf("error while unmarshalling configuration file: %w", err))

		return nil, err
	}

	err = validator.Validate(config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Flags enables shorthand flags for specific config values.
func flags(f *flag.FlagSet) {
	f.StringP("config", "c", "", "The configuration file to use to configure this application")
	app.Flags(f, appSpec)
	ops.Flags(f, opsSpec)
}

func getMap(config interface{}) map[string]interface{} {
	var inInterface map[string]interface{}
	_ = mapstructure.Decode(config, &inInterface)
	return inInterface
}

func (c *Config) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (c *Config) ServeJSON(w http.ResponseWriter, r *http.Request) {
	// Error is ignored for test coverage purposes and the fact that errors here do not cause a panic and don't serve actual requests
	payload, _ := json.Marshal(c)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (c *Config) ServeYAML(w http.ResponseWriter, r *http.Request) {
	// Error is ignored for test coverage purposes and the fact that errors here do not cause a panic and don't serve actual requests
	payload, _ := yaml.Marshal(c)
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
