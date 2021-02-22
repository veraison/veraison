package frontend

import (
	"path"

	"github.com/veraison/common"
	"github.com/veraison/tokenprocessor"
)

func NewTokenProcessor(pluginDir string, dbPath string) (*tokenprocessor.TokenProcessor, error) {

	trustAnchorDbPath := path.Join(dbPath, "trustanchors.sqlite3")

	// TODO make processor params configurable
	var config = tokenprocessor.TokenProcessorConfig{
		PluginLocations:      []string{pluginDir},
		TrustAnchorStoreName: "sqlite",
		TrustAnchorStoreParams: common.TrustAnchorStoreParams{
			"dbPath": trustAnchorDbPath,
		},
	}

	var tp tokenprocessor.TokenProcessor
	err := tp.Init(config)
	if err != nil {
		return nil, err
	}

	return &tp, nil
}
