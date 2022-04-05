// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// DefaultConfigPaths is a set of default paths that can be specified for
// Config.Init().
var DefaultConfigPaths = []string{
	"$HOME/veraison",
	".",
}

type ConfigPaths []string

func (p *ConfigPaths) Add(value string) error {
	*p = append(*p, value)
	return nil
}

func (p *ConfigPaths) Strings() []string {
	return []string(*p)[:]
}

func (p *ConfigPaths) String() string {
	result, err := json.Marshal(p)
	if err == nil {
		return string(result)
	}

	return ""
}

func (p *ConfigPaths) Set(text string) error {
	paths := strings.Split(text, string(os.PathListSeparator))

	for _, path := range paths {
		if err := p.Add(path); err != nil {
			return err
		}
	}

	return nil
}

func NewConfigPaths() *ConfigPaths {
	paths := new(ConfigPaths)
	for _, p := range DefaultConfigPaths {
		*paths = append(*paths, p)
	}

	return paths
}

// Config encapsulates Veraison configuration, maintaining the complete set of
// configuration points and populating them from config files discovered in
// pre-defined locations.
type Config struct {
	v      *viper.Viper
	params map[string]*ParamStore
	stores map[string]*ParamStore
}

func NewConfig(paths []string, stores ...*ParamStore) (*Config, error) {
	c := &Config{}
	err := c.Init(paths, stores...)
	return c, err
}

// Init initializes a Config instances, setting the search paths to the
// specified locations and setting configuration type and file name.
func (c *Config) Init(paths []string, stores ...*ParamStore) error {
	c.v = viper.New()
	c.v.SetConfigName("config.yaml")
	c.v.SetConfigType("yaml")
	for _, path := range paths {
		c.v.AddConfigPath(path)
	}

	c.params = make(map[string]*ParamStore)
	c.stores = make(map[string]*ParamStore)

	for _, store := range stores {
		if err := c.AddStore(store); err != nil {
			return err
		}
	}

	return nil
}

// AddPath adds an additional path to search for config files without removing
// paths added during Init().
func (c Config) AddPath(path string) {
	c.v.AddConfigPath(path)
}

// SetFileName update the name of the file that contains configuration from the default config.yaml.
func (c Config) SetFileName(name string) {
	c.v.SetConfigName(name)
}

func (c Config) AddStore(store *ParamStore) error {
	storeName := store.GetName()
	if _, ok := c.stores[storeName]; ok {
		return fmt.Errorf("store with name %q already added", storeName)
	}

	for _, name := range store.GetParamNames() {
		otherStore, ok := c.params[name]
		if ok {
			return fmt.Errorf("param %q already registered by store %q", name, otherStore.GetName())
		}
	}

	store.Freeze() // prevent subsequent modifications to parameter definitions.

	c.stores[storeName] = store
	for _, name := range store.GetParamNames() {
		c.params[name] = store
	}

	return nil
}

func (c *Config) ReadInConfig() error {
	if err := c.v.ReadInConfig(); err != nil {
		return err
	}

	return c.Reload()
}

func (c *Config) ReadConfig(in io.Reader) error {
	if err := c.v.ReadConfig(in); err != nil {
		return err
	}

	return c.Reload()
}

// Reload re-populates the config with discovered values, and validates to make
// sure that read values match expected types, and all mandatory parameters
// have been created.
func (c *Config) Reload() error {
	for _, store := range c.stores {
		if err := store.PopulateFromViper(c.v); err != nil {
			return err
		}
	}

	if err := c.validate(); err != nil {
		return err
	}

	return nil
}

func (c Config) Get(name string) interface{} {
	cv := reflect.ValueOf(c)
	retv := cv.FieldByName(name)
	return retv.Interface()
}

func (c Config) GetBool(name string) bool {
	store, ok := c.params[name]
	if !ok {
		return false
	}

	return store.GetBool(name)
}

func (c Config) GetInt(name string) int {
	store, ok := c.params[name]
	if !ok {
		return 0
	}

	return store.GetInt(name)
}

func (c Config) GetString(name string) string {
	store, ok := c.params[name]
	if !ok {
		return ""
	}

	return store.GetString(name)
}

func (c Config) GetStringSlice(name string) []string {
	store, ok := c.params[name]
	if !ok {
		return nil
	}

	return store.GetStringSlice(name)
}

func (c Config) GetStringMapString(name string) map[string]string {
	store, ok := c.params[name]
	if !ok {
		return nil
	}

	return store.GetStringMapString(name)
}

func (c Config) GetParamStore(name string) (*ParamStore, error) {
	store, ok := c.stores[name]
	if !ok {
		return nil, fmt.Errorf("param store %q not in config", name)
	}
	return store, nil
}

func (c Config) validate() error {
	for _, store := range c.stores {
		if err := store.Validate(true); err != nil {
			return fmt.Errorf("%s params validation failed: %s", store.Name, err.Error())
		}
	}

	return nil
}
