package main

import (
	"github.com/veraison/common"
)

type Config struct {
	common.BaseConfig
	PluginLocations          []string
	EndorsementBackendName   string
	EndorsementBackendParams common.EndorsementBackendParams
	Port                     int
	Quiet                    bool
}

func (c *Config) Reload() error {
	if err := c.BaseConfig.Reload(); err != nil {
		return err
	}

	c.PluginLocations = c.Viper.GetStringSlice("plugin.locations")
	c.EndorsementBackendName = c.Viper.GetString("endorsements.backend_name")
	c.EndorsementBackendParams = common.EndorsementBackendParams(c.Viper.GetStringMap("endorsements.backend_params"))
	c.Port = c.Viper.GetInt("endorsements.store_port")
	c.Quiet = c.Viper.GetBool("quiet")

	return nil
}

func (c Config) ExpectedKeys() []string {
	return append(c.BaseConfig.ExpectedKeys(),
		"plugin.locations",
		"endorsements.backend_name",
		"endorsements.backend_params",
		"endorsements.store_port",
	)
}

func (c Config) GetPluginLocations() []string {
	return c.PluginLocations
}

func (c Config) GetEndorsementBackendName() string {
	return c.EndorsementBackendName
}

func (c Config) GetEndorsementBackendParams() common.EndorsementBackendParams {
	return c.EndorsementBackendParams
}

func (c Config) GetQuiet() bool {
	return c.Quiet
}

func NewConfig(paths []string) (*Config, error) {
	config := &Config{}
	if err := config.Init(paths); err != nil {
		return nil, err
	}
	config.BaseConfig.SetDerived(config)
	return config, nil
}
