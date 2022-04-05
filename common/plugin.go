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

// DefaultPluginSet maps the plugin type onto the corresponding plugin struct
// for the plugins used by Veraison.
var DefaultPluginSet = plugin.PluginSet{
	"endorsementstore": &EndorsementBackendPlugin{},
	"policyengine":     &PolicyEnginePlugin{},
	"policystore":      &PolicyStorePlugin{},
	"scheme":           &SchemePlugin{},
}

// DefaultPluginHandshakeConfig defins the handshake used by Veraison plugins.
var DefaultPluginHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "VERAISON_PLUGIN",
	MagicCookieValue: "VERAISON",
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

// PluginManager loads and keeps track of plugins, cleaning them up when
// Close() is called.
type PluginManager struct {
	locations       []string
	logLevel        hclog.Level
	plugins         []*LoadedPlugin
	pluginSet       plugin.PluginSet
	handshakeConfig plugin.HandshakeConfig
}

// NewPluginManager returns a pointer to a PluginManger initialized with the
// specified plugin locations and log level.
func NewPluginManager(
	locations []string,
	logLevel hclog.Level,
) (*PluginManager, error) {
	manager := new(PluginManager)
	err := manager.Init(locations, logLevel)
	return manager, err
}

// Init sets the locatins where the PluginManager will look for plugins, and
// whether to quieten logging output from the same.
func (m *PluginManager) Init(locations []string, logLevel hclog.Level) error {
	var validatedLocations []string

	for _, path := range locations {
		stat, err := os.Stat(path)

		if err == nil {
			if stat.IsDir() {
				validatedLocations = append(validatedLocations, path)
			} else {
				return fmt.Errorf("'%s' is not a directory", path)
			}
		} else { // err != nil
			if os.IsNotExist(err) {
				return fmt.Errorf("directory '%s' does not exist", path)
			}
			return err
		}
	}

	m.locations = validatedLocations
	m.logLevel = logLevel
	m.pluginSet = DefaultPluginSet
	m.handshakeConfig = DefaultPluginHandshakeConfig

	return nil
}

// SetPlugins specifies the new plugin set the PluginManager will use whent
// creating the plugin client.
func (m *PluginManager) SetPlugins(newPlugins plugin.PluginSet) {
	m.pluginSet = newPlugins
}

func (m *PluginManager) SetHandshakeConfig(newConfig plugin.HandshakeConfig) {
	m.handshakeConfig = newConfig
}

// Load returns an interface{} from the loaded plugin. The plugin itself is
// kept track of by the PluginManager, and will be cleaned up when Close() is
// called. The plugin to be loaded is specified by its type (see PluginMap) and
// name.
func (m *PluginManager) Load(pluginType, pluginName string) (interface{}, error) {
	loaded, err := DoLoadPlugin(
		m.locations,
		pluginType,
		pluginName,
		m.logLevel,
		m.handshakeConfig,
		m.pluginSet,
	)
	if err != nil {
		return nil, err
	}

	m.plugins = append(m.plugins, loaded)

	return loaded.Raw, nil
}

// Close cleans up all loaded plugins, terminating thier processes.
// Note: this will also terminate plugins loaded with LoadPlugin.
func (m PluginManager) Close() {
	plugin.CleanupClients()
}

type PluginLogWriter struct {
	Logger hclog.Logger
	Level  hclog.Level
}

func NewPluginLogWriter(logger hclog.Logger, level hclog.Level) *PluginLogWriter {
	logWriter := new(PluginLogWriter)
	logWriter.Logger = logger
	logWriter.Level = level
	return logWriter
}

func (w PluginLogWriter) Write(p []byte) (int, error) {
	w.Logger.Log(w.Level, string(p))
	return len(p), nil
}

// LoadPlugin returns a pointer to a LoadedPlugin based on the plugin type and
// names specfied, by search for a suitable plugin binary inside the provided
// locations.
func LoadPlugin(
	locations []string,
	plugType, plugName string,
	quiet bool,
) (*LoadedPlugin, error) {

	var logLevel hclog.Level
	if quiet {
		logLevel = hclog.Error
	} else {
		logLevel = hclog.Warn
	}

	return DoLoadPlugin(
		locations, plugType, plugName,
		logLevel, DefaultPluginHandshakeConfig, DefaultPluginSet)

}

// DoLoadPlugin returns a pointer ot a LoadedPlugin, similarly to LoadedPlugin.
// It differes in that it allows to specify the handshake config and the plugin
// map, rather than using the default ones assumed by Vereaison. This is
// primarily to enable testing of PluginManager.
func DoLoadPlugin(
	locations []string,
	plugType, plugName string,
	logLevel hclog.Level,
	handshakeConfig plugin.HandshakeConfig,
	pluginSet plugin.PluginSet,
) (*LoadedPlugin, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   plugName,
		Output: os.Stdout,
		Level:  logLevel,
	})

	writer := NewPluginLogWriter(logger, logLevel)

	for _, location := range locations {
		files, err := ioutil.ReadDir(location)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range files {
			pluginPath := filepath.Join(location, fileInfo.Name())

			client := plugin.NewClient(&plugin.ClientConfig{
				HandshakeConfig: handshakeConfig,
				Plugins:         pluginSet,
				Managed:         true,
				Cmd:             exec.Command(pluginPath),
				Logger:          logger,
				SyncStdout:      writer,
				SyncStderr:      writer,
			})

			rpcClient, err := client.Client()
			if err != nil {
				logger.Debug("failed to load RPC client from",
					pluginPath, ":", err)
				client.Kill()
				continue
			}

			raw, err := rpcClient.Dispense(plugType)
			if err != nil {
				logger.Debug("plugin", pluginPath,
					"does not implement a", plugType)
				client.Kill()
				continue
			}

			named := raw.(INamed)
			if !strings.EqualFold(
				named.GetName(),
				strings.ToLower(plugName),
			) {
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
