// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginManger_Load(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	manager, err := NewPluginManager([]string{"test/plugin/bin/"}, hclog.Info)
	require.Nil(err)
	defer manager.Close()

	manager.SetPlugins(TestPluginSet)

	test_load_good(manager, assert, require)
	test_load_nonexistent(manager, assert, require)
	test_load_bad_type(manager, assert, require)
}

func TestPluginManger_Load_bad_paths(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	_, err := NewPluginManager([]string{"does/not/exist/"}, hclog.Info)
	assert.EqualError(err, "directory 'does/not/exist/' does not exist")

	tmpFile, err := ioutil.TempFile(os.TempDir(), "veraison-test-")
	require.Nil(err)
	defer os.Remove(tmpFile.Name())

	_, err = NewPluginManager([]string{tmpFile.Name()}, hclog.Info)
	expected := fmt.Sprintf("'%s' is not a directory", tmpFile.Name())
	assert.EqualError(err, expected)
}

func test_load_good(
	manager *PluginManager,
	assert *assert.Assertions,
	require *require.Assertions,
) {
	loaded, err := manager.Load("test", "good")
	require.Nil(err)

	plugin, ok := loaded.(ITestPluginInterface)
	require.True(ok)
	assert.Equal("good", plugin.GetName())
}

func test_load_nonexistent(
	manager *PluginManager,
	assert *assert.Assertions,
	require *require.Assertions,
) {
	loaded, err := manager.Load("test", "nonexistent")
	assert.Nil(loaded)
	assert.EqualError(err, "could not find test plugin 'nonexistent'")
}

func test_load_bad_type(
	manager *PluginManager,
	assert *assert.Assertions,
	require *require.Assertions,
) {
	loaded, err := manager.Load("bad", "good")
	assert.Nil(loaded)
	assert.EqualError(err, "could not find bad plugin 'good'")
}
