// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/hashicorp/go-plugin"

	"github.com/veraison/common"
)

type Manager struct {
	StoreName string
	Store     common.IPolicyStore
	RPCClient plugin.ClientProtocol
	Client    *plugin.Client
}

func NewManager() *Manager {
	return &Manager{
		StoreName: "[none]",
	}
}

func (pm *Manager) InitializeStore(
	pluginLocaitons []string,
	name string,
	params common.PolicyStoreParams,
) error {
	name = common.Canonize(name)

	lp, err := common.LoadPlugin(pluginLocaitons, "policystore", name)
	if err != nil {
		return err
	}

	pm.Store = lp.Raw.(common.IPolicyStore)
	pm.RPCClient = lp.RPCClient
	pm.Client = lp.PluginClient

	if err = pm.Store.Init(params); err != nil {
		pm.Client.Kill()
		return err
	}

	return nil
}

func (pm *Manager) GetPolicy(tenantID int, tokenFormat common.TokenFormat) (*common.Policy, error) {
	return pm.Store.GetPolicy(tenantID, tokenFormat)
}

func (pm *Manager) PutPolicy(tenantID int, policy *common.Policy) error {
	return pm.Store.PutPolicy(tenantID, policy)
}

func (pm *Manager) PutPolicyBytes(tenantID int, policyBytes []byte) error {
	policies, err := common.ReadPoliciesFromBytes(policyBytes)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		err = pm.PutPolicy(tenantID, policy)
		if err != nil {
			// TODO: implement transactions and roll back here
			return err
		}
	}

	return nil
}

func (pm *Manager) Close() {
	pm.Client.Kill()
	pm.RPCClient.Close()
}
