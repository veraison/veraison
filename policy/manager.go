// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/hashicorp/go-plugin"

	"veraison/common"
)

type PolicyManager struct {
	StoreName string
	Store     common.IPolicyStore
	RpcClient plugin.ClientProtocol
	Client    *plugin.Client
}

func NewPolicyManager() *PolicyManager {
	return &PolicyManager{
		StoreName: "[none]",
	}
}

func (pm *PolicyManager) InitializeStore(
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
	pm.RpcClient = lp.RpcClient
	pm.Client = lp.PluginClient

	if err = pm.Store.Init(params); err != nil {
		pm.Client.Kill()
		return err
	}

	return nil
}

func (pm *PolicyManager) GetPolicy(tenantId int, tokenFormat common.TokenFormat) (*common.Policy, error) {
	return pm.Store.GetPolicy(tenantId, tokenFormat)
}

func (pm *PolicyManager) PutPolicy(tenantId int, policy *common.Policy) error {
	return pm.Store.PutPolicy(tenantId, policy)
}

func (pm *PolicyManager) PutPolicyBytes(tenantId int, policyBytes []byte) error {
	policies, err := common.ReadPoliciesFromBytes(policyBytes)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		err = pm.PutPolicy(tenantId, policy)
		if err != nil {
			// TODO: implement transactions and roll back here
			return err
		}
	}

	return nil
}

func (pm *PolicyManager) Close() {
	pm.Client.Kill()
	pm.RpcClient.Close()
}
