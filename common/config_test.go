package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LoadConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	wd, err := os.Getwd()
	require.Nil(err)

	configsDirs := []string{filepath.Join(wd, "test", "configs")}

	var config Config
	config.Init(configsDirs)
	require.Nil(config.Reload())

	assert.Equal("opa", config.PolicyEngineName)
	assert.Equal("sqlite", config.PolicyStoreName)
	assert.Equal("sqlite", config.EndorsementStoreName)
	assert.Equal(true, config.Debug)
	assert.Equal([]string{"../plugins/bin/"}, config.PluginLocations)
}

func Test_LoadBadConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	wd, err := os.Getwd()
	require.Nil(err)

	configsDirs := []string{filepath.Join(wd, "test", "configs")}

	var config Config
	config.Init(configsDirs)
	config.SetFileName("no-engine.yaml")

	err = config.Reload()
	assert.NotNil(err)
	assert.Equal("key \"policy.engine_name\" not set in configuration", err.Error())
}
