// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// PluginMap maps the name of a plugin type onto the corresponding plugin struct.
var PluginMap = map[string]plugin.Plugin{
	"endorsementstore": &EndorsementBackendPlugin{},
	"policyengine":     &PolicyEnginePlugin{},
	"policystore":      &PolicyStorePlugin{},
	"scheme":           &SchemePlugin{},
}

// LoadedPlugin encapsulates a loaded Hashicorp plugin.
type LoadedPlugin struct {
	Raw          interface{}
	PluginClient *plugin.Client
	RPCClient    plugin.ClientProtocol
}

// INamed defines inteface for named entities.
type INamed interface {

	// GetName returns a string with the name of this entity.
	GetName() string
}

// LoadPlugin returns a pointer to a LoadedPlugin based on the plugin type and
// names specfied, by search for a suitable plugin binary inside the provided
// locations.
func LoadPlugin(locations []string, plugType, plugName string, quiet bool) (*LoadedPlugin, error) {

	handshakeConfig := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var logLevel hclog.Level
	if quiet {
		logLevel = hclog.Error
	} else {
		logLevel = hclog.Warn
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   plugName,
		Output: os.Stdout,
		Level:  logLevel,
	})

	for _, location := range locations {
		files, err := ioutil.ReadDir(location)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range files {
			pluginPath := filepath.Join(location, fileInfo.Name())

			client := plugin.NewClient(&plugin.ClientConfig{
				HandshakeConfig: handshakeConfig,
				Plugins:         PluginMap,
				Cmd:             exec.Command(pluginPath),
				Logger:          logger,
			})

			rpcClient, err := client.Client()
			if err != nil {
				logger.Debug("failed to load RPC client from", pluginPath, ":", err)
				client.Kill()
				continue
			}

			raw, err := rpcClient.Dispense(plugType)
			if err != nil {
				logger.Debug("plugin", pluginPath, "does not implement a", plugType)
				client.Kill()
				continue
			}

			named := raw.(INamed)
			if !strings.EqualFold(named.GetName(), strings.ToLower(plugName)) {
				logger.Debug("wrong name in", pluginPath)
				client.Kill()
				continue
			}

			return &LoadedPlugin{
				Raw:          raw,
				PluginClient: client,
				RPCClient:    rpcClient,
			}, nil

		}
	}

	return nil, fmt.Errorf("could not find %v plugin '%v'", plugType, plugName)
}
