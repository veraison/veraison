// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"github.com/veraison/common"
)

type Config struct {
	common.BaseConfig

	// A slice of paths that will be checked when searching for plugins.
	PluginLocations []string

	// The name of the PolicStorePlugin implementation that should be loaded
	// for this deployment.
	PolicyStoreName string

	// The name of the PolicyEnginePlugin that should be loaded for this
	// deployment.
	PolicyEngineName string

	// Parameters to be passed for initializing the policy store. See the
	// documentation for a particular store implementation for which
	// parameters are valid, and what values are accepted.
	PolicyStoreParams common.PolicyStoreParams

	// Parameters to be passed for initializing the policy engine. See the
	// documentation for a particular engine implementation for which
	// parameters are valid, and what values are accepted.
	PolicyEngineParams common.PolicyEngineParams

	// Address of the entorsement store. The format depends on the type store
	// used, but would typically be in the form "<host>:<port>" or similar.
	EndorsementStoreHost string
	EndorsementStorePort int
}

func (c *Config) Reload() error {
	if err := c.BaseConfig.Reload(); err != nil {
		return err
	}

	c.PluginLocations = c.Viper.GetStringSlice("plugin.locations")
	c.PolicyStoreName = c.Viper.GetString("policy.store_name")
	c.PolicyStoreParams = c.Viper.GetStringMapString("policy.store_params")
	c.PolicyEngineName = c.Viper.GetString("policy.engine_name")
	c.PolicyEngineParams = c.Viper.GetStringMapString("policy.engine_params")
	c.EndorsementStoreHost = c.Viper.GetString("endorsements.store_host")
	c.EndorsementStorePort = c.Viper.GetInt("endorsements.store_port")

	return nil
}

func (c Config) ExpectedKeys() []string {
	return append(c.BaseConfig.ExpectedKeys(),
		"plugin.locations",
		"policy.store_name",
		"policy.store_params",
		"policy.engine_name",
		"endorsements.store_host",
		"endorsements.store_port",
	)
}

func (c Config) Kind() common.ConfigKind {
	return common.VerifierConfig
}

// NewConfig creates a new Config instance initialized with the default set of
// config paths.
func NewConfig(paths []string) *Config {
	config := &Config{}
	config.Init(paths)
	config.BaseConfig.SetDerived(config)
	return config
}
