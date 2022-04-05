// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/veraison/common"
)

type Plugin struct {
}

func (p *Plugin) GetName() string {
	return "good"
}

func main() {
	var pluginMap = map[string]plugin.Plugin{
		"test": &common.TestPlugin{
			Impl: &Plugin{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: common.DefaultPluginHandshakeConfig,
		Plugins:         pluginMap,
	})
}
