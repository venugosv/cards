package config

import (
	"os"
	"testing"

	"github.com/anzx/pkg/validator"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfig         = `../config/local.yaml`
	testConfigFileKey     = "TEST_CONFIG_FILE"
	pubsubEmulatorHostKey = "PUBSUB_EMULATOR_HOST"
	configPrefix          = "TEST"
)

// Load begins by attempting to load in config by file as defined in `TEST_CONFIG_FILE` env var. If this returns a nil
// string, config will be loaded in from environment variables entirely
func Load(t *testing.T) (cfg *Config, errResp error) {
	defer func() {
		setRequiredVars(cfg)
	}()

	file := os.Getenv(testConfigFileKey)
	if file != "" {
		return fromFile(t, file)
	}

	cfg, err := fromEnv()
	if err == nil {
		return cfg, err
	}

	config, err := fromFile(t, defaultConfig)

	return config, err
}

// fromFile loads in a yaml file from the local system and decodes it into a Config struct.
func fromFile(t *testing.T, in string) (cfg *Config, errResp error) {
	defer func() {
		if err := cfg.Validate(); err != nil {
			errResp = errors.Wrap(errResp, err.Error())
		}
	}()

	t.Logf("loading config from file: %s", in)
	file, err := os.Open(in)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open config in")
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)

	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, errors.Wrap(err, "unable to parse config in")
	}

	return &config, nil
}

// fromEnv constructs a Config object from environment variables
func fromEnv() (config *Config, errResp error) {
	defer func() {
		if err := config.Validate(); err != nil {
			errResp = errors.Wrap(errResp, err.Error())
		}
	}()

	var cfg *Config
	if err := envconfig.Process(configPrefix, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setRequiredVars(config *Config) {
	if config == nil {
		return
	}
	if config.CommandCentre.EmulatorHost != os.Getenv(pubsubEmulatorHostKey) {
		os.Setenv(pubsubEmulatorHostKey, config.CommandCentre.EmulatorHost)
	}
}

// Validate config struct
func (c *Config) Validate() error {
	return validator.Validate(c)
}
