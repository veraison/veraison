package frontend

import (
	"path"

	"github.com/veraison/common"
	"github.com/veraison/verifier"

	"go.uber.org/zap"
)

func NewVerifier(pluginDir string, dbPath string, logger *zap.Logger) (*verifier.Verifier, error) {

	policyDbPath := path.Join(dbPath, "policy.sqlite3")
	endorsementDbPath := path.Join(dbPath, "endorsements.sqlite3")

	// TODO make engine/store names configurable
	var vc = verifier.Config{
		PluginLocations:      []string{pluginDir},
		PolicyEngineName:     "opa",
		PolicyStoreName:      "sqlite",
		EndorsementStoreName: "sqlite",
		PolicyStoreParams: common.PolicyStoreParams{
			"dbPath": policyDbPath,
		},
		EndorsementStoreParams: common.EndorsementStoreParams{
			"dbPath": endorsementDbPath,
		},
	}

	v, err := verifier.NewVerifier(logger)
	if err != nil {
		return nil, err
	}

	err = v.Initialize(vc)
	if err != nil {
		return nil, err
	}

	return v, nil
}
