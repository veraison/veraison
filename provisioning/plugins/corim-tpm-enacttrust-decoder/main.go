// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/veraison/veraison/provisioning/decoder"
)

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PROVISIONING_DECODER_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"decoder": &decoder.Plugin{
			Impl: &Decoder{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
