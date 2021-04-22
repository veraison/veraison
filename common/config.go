package common

import (
	"fmt"

	"github.com/spf13/viper"
)

// DefaultConfigPaths is a set of default paths that can be specified for
// Config.Init().
var DefaultConfigPaths = []string{
	"$HOME/veraison",
	".",
}

var expectedConfigKey = []string{
	"plugin.locations",
	"policy.store_name",
	"policy.engine_name",
	"endorsements.store_name",
}

// Config encapsulates Veraison configuration, maintaining the complete set of
// configuration points and populating them from config files discovered in
// pre-defined locations.
type Config struct {

	// If set to "true" enable debug level logging.
	Debug bool

	// A slice of paths that will be checked when searching for plugins.
	PluginLocations []string

	// The name of the PolicStorePlugin implementation that should be loaded
	// for this deployment.
	PolicyStoreName string

	// The name of the PolicyEnginePlugin that should be loaded for this
	// deployment.
	PolicyEngineName string

	// The name of the EndorsementStorePlugin that should be loaded for
	// this deployment.
	EndorsementStoreName string

	// Parameters to be passed for initializing the policy store. See the
	// documentation for a particular store implementation for which
	// parameters are valid, and what values are accepted.
	PolicyStoreParams PolicyStoreParams

	// Parameters to be passed for initializing the policy engine. See the
	// documentation for a particular engine implementation for which
	// parameters are valid, and what values are accepted.
	PolicyEngineParams PolicyEngineParams

	// Parameters to be passed for initializing the endorsements store. See the
	// documentation for a particular store implementation for which
	// parameters are valid, and what values are accepted.
	EndorsementStoreParams EndorsementStoreParams

	viper *viper.Viper
}

// NewConfig creates a new Config instance initialized with the default set of
// config paths.
func NewConfig() *Config {
	config := &Config{}
	config.Init(DefaultConfigPaths)
	return config
}

// Init initializes a Config instances, setting the search paths to the
// specified locations and setting configuration type and file name.
func (c *Config) Init(paths []string) {
	c.viper = viper.New()
	c.viper.SetConfigName("config.yaml")
	c.viper.SetConfigType("yaml")
	for _, path := range paths {
		c.viper.AddConfigPath(path)
	}
}

// AddPath adds an additional path to search for config files without removing
// paths added during Init().
func (c Config) AddPath(path string) {
	c.viper.AddConfigPath(path)
}

// Update the name of the file that contains configuration from the default config.yaml.
func (c Config) SetFileName(name string) {
	c.viper.SetConfigName(name)
}

// Reload re-runs configuration discovery based on config paths currently set,
// and repo-poulates the config with discovered values.
func (c *Config) Reload() error {
	if err := c.viper.ReadInConfig(); err != nil {
		return err
	}

	if err := c.validate(); err != nil {
		return err
	}

	c.Debug = c.viper.GetBool("debug")
	c.PluginLocations = c.viper.GetStringSlice("plugin.locations")
	c.PolicyStoreName = c.viper.GetString("policy.store_name")
	c.PolicyStoreParams = c.viper.GetStringMapString("policy.store_params")
	c.PolicyEngineName = c.viper.GetString("policy.engine_name")
	c.PolicyEngineParams = c.viper.GetStringMapString("policy.engine_params")
	c.EndorsementStoreName = c.viper.GetString("endorsements.store_name")
	c.EndorsementStoreParams = c.viper.GetStringMap("endorsements.store_params")

	return nil
}

func (c Config) validate() error {
	for _, key := range expectedConfigKey {
		if !c.viper.IsSet(key) {
			return fmt.Errorf("key %q not set in configuration", key)
		}
	}
	return nil
}
