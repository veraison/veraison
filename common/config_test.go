// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store1 := NewParamStore("one")
	require.Nil(store1.AddParamDefinitions(map[string]*ParamDescription{
		"serverHost": {
			Kind:     uint32(reflect.String),
			Path:     "server.host",
			Required: ParamNecessity_REQUIRED,
		},
		"serverPort": {
			Kind:     uint32(reflect.Int),
			Path:     "server.port",
			Required: ParamNecessity_REQUIRED,
		},
		"serverDebug": {
			Kind:     uint32(reflect.Bool),
			Path:     "debug",
			Required: ParamNecessity_REQUIRED,
		},
	}))

	store2 := NewParamStore("two")
	require.Nil(store2.AddParamDefinitions(map[string]*ParamDescription{
		"clientHost": {
			Kind:     uint32(reflect.String),
			Path:     "client.host",
			Required: ParamNecessity_REQUIRED,
		},
		"clientPort": {
			Kind:     uint32(reflect.Int),
			Path:     "client.port",
			Required: ParamNecessity_REQUIRED,
		},
		"clientDebug": {
			Kind:     uint32(reflect.Bool),
			Path:     "debug",
			Required: ParamNecessity_REQUIRED,
		},
	}))

	store3 := NewParamStore("three")
	require.Nil(store3.AddParamDefinitions(map[string]*ParamDescription{
		"clientHost": {
			Kind:     uint32(reflect.String),
			Path:     "other_client.host",
			Required: ParamNecessity_REQUIRED,
		},
	}))

	var configPaths []string
	config, err := NewConfig(configPaths, store1, store2)
	require.Nil(err)

	configText := []byte(`
debug: true
server:
  host: server.random
  port: 9999
client:
  host: client.random
  port: 6666
other_client:
  host: other.random
`)
	require.Nil(config.ReadConfig(bytes.NewBuffer(configText)))
	assert.Equal("server.random", config.GetString("serverHost"))

	store := config.GetParamStore("one")
	require.NotNil(store)
	assert.Equal(9999, store.GetInt("serverPort"))

	err = config.AddStore(NewParamStore("two"))
	require.NotNil(err)
	assert.Equal("store with name \"two\" already added", err.Error())

	err = config.AddStore(store3)
	require.NotNil(err)
	assert.Equal("param \"clientHost\" already registered by store \"two\"", err.Error())
}
