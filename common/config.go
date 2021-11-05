package common

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/spf13/viper"
)

// DefaultConfigPaths is a set of default paths that can be specified for
// Config.Init().
var DefaultConfigPaths = []string{
	"$HOME/veraison",
	".",
}

var expectedConfigKey = []string{
	"plugin.locations",
	"policy.store_name",
	"policy.engine_name",
	"endorsements.store_name",
}

type ConfigKind int
const (
	GenericConfig ConfigKind = iota
	VerifierConfig
	EndorsementStoreConfig
)


type ConfigPaths []string


func (p *ConfigPaths) Set(value string) error {
	*p = append(*p, value)
	return nil
}

func (p *ConfigPaths) String() string {
	result, err := json.Marshal(p)
	if err == nil {
		return string(result)
	}

	return ""
}

func NewConfigPaths() *ConfigPaths {
	paths := new(ConfigPaths)
	for _, p := range DefaultConfigPaths {
		*paths = append(*paths, p)
	}

	return paths
}

type IConfig interface {
	Init(paths []string) error
	AddPath(path string)
	SetFileName(name string)
	Reload() error
	ReadConfig(in io.Reader) error

	Kind() ConfigKind
	ExpectedKeys() []string

	Get(name string) interface{}
	GetInt(name string) int64
	GetString(name string) string
	GetStringSlice(name string) []string
	GetStringMapString(name string) map[string]string
}

// Config encapsulates Veraison configuration, maintaining the complete set of
// configuration points and populating them from config files discovered in
// pre-defined locations.
type BaseConfig struct {
	// If set to "true" enable debug level logging.
	Debug bool

	derived IConfig
	Viper *viper.Viper
}

func (c *BaseConfig) SetDerived(derived IConfig) {
	c.derived = derived
}

// Init initializes a Config instances, setting the search paths to the
// specified locations and setting configuration type and file name.
func (c *BaseConfig) Init(paths []string) error {
	c.Viper = viper.New()
	c.Viper.SetConfigName("config.yaml")
	c.Viper.SetConfigType("yaml")
	for _, path := range paths {
		c.Viper.AddConfigPath(path)
	}

	return nil
}

// AddPath adds an additional path to search for config files without removing
// paths added during Init().
func (c BaseConfig) AddPath(path string) {
	c.Viper.AddConfigPath(path)
}

// Update the name of the file that contains configuration from the default config.yaml.
func (c BaseConfig) SetFileName(name string) {
	c.Viper.SetConfigName(name)
}

func (c *BaseConfig) ReadInConfig() error {
	if err := c.Viper.ReadInConfig(); err != nil {
		return err
	}

	return  c.derived.Reload()
}

func (c *BaseConfig) ReadConfig(in io.Reader) error {
	if err := c.Viper.ReadConfig(in); err != nil {
		return err
	}

	return  c.derived.Reload()
}

// Reload re-populates the config with discovered values.
func (c *BaseConfig) Reload() error {

	if err := c.validate(); err != nil {
		return err
	}

	c.Debug = c.Viper.GetBool("debug")

	return nil
}

func (c BaseConfig) Kind() ConfigKind {
	return GenericConfig
}

func (c BaseConfig) ExpectedKeys() []string {
	return []string{"debug"}
}

func (c BaseConfig) Get(name string) interface{} {
	cv := reflect.ValueOf(c)
	retv  := cv.FieldByName(name)
	return retv.Interface()
}

func (c BaseConfig) GetInt(name string) int64 {
	cv := reflect.ValueOf(c)
	retv  := cv.FieldByName(name)

	switch retv.Kind() {
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		return retv.Int()
	}

	return 0
}

func (c BaseConfig) GetString(name string) string {
	cv := reflect.ValueOf(c)
	retv  := cv.FieldByName(name)

	if retv.Kind() == reflect.String {
		return retv.String()
	} else {
		return ""
	}
}

func (c BaseConfig) GetStringSlice(name string) []string {
	cv := reflect.ValueOf(c)
	retv  := cv.FieldByName(name)

	if retv.Kind() == reflect.Slice {
		slice, ok :=  retv.Interface().([]string)
		if ok {
			return slice
		}
	}

	return []string{}
}

func (c BaseConfig) GetStringMapString(name string) map[string]string {
	cv := reflect.ValueOf(c)
	retv  := cv.FieldByName(name)

	if retv.Kind() == reflect.Map {
		m, ok := retv.Interface().(map[string]string)
		if ok {
			return m
		}
	}

	return nil
}

func (c BaseConfig) validate() error {
	for _, key := range c.derived.ExpectedKeys() {
		if !c.Viper.IsSet(key) {
			return fmt.Errorf("key %q not set in configuration", key)
		}
	}
	return nil
}
