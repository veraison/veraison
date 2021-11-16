package frontend

import (
	"path"

	"github.com/veraison/common"
	"github.com/veraison/policy"
	"github.com/veraison/trustedservices"
	"github.com/veraison/verifier"

	"go.uber.org/zap"
)

func NewVerifier(pluginDir string, dbPath string, logger *zap.Logger) (*verifier.Verifier, error) {

	policyDbPath := path.Join(dbPath, "policy.sqlite3")

	verifierParams, err := verifier.NewVerifierParams()
	if err != nil {
		return nil, err
	}

	pluginLocations := []string{pluginDir}

	// TODO make configurable
	verifierParams.SetStringSlice("PluginLocations", pluginLocations)
	verifierParams.SetString("VtsHost", "")
	verifierParams.SetString("VtsPort", "")
	verifierParams.SetStringMapString("VtsParams", make(map[string]string))

	v, err := verifier.NewVerifier(logger)
	if err != nil {
		return nil, err
	}

	connector := new(trustedservices.LocalClientConnector)

	policyManagerParams, err := policy.NewManagerParamStore()
	if err != nil {
		return nil, err
	}
	// TODO make configurable
	policyManagerParams.SetStringSlice("PluginLocations", pluginLocations)
	policyManagerParams.SetString("PolicyStoreName", "sqlite")
	policyManagerParams.SetStringMapString("PolicyStoreParams", map[string]string{"dbPath": policyDbPath})

	pm := policy.NewManager()
	err = pm.Init(policyManagerParams)
	if err != nil {
		return nil, err
	}

	pe, err := common.LoadPolicyEnginePlugin(pluginLocations, "opa")
	if err != nil {
		return nil, err
	}

	err = v.Init(verifierParams, connector, pm, pe)
	return v, err
}
