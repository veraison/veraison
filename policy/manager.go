// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"reflect"

	"github.com/veraison/common"
)

func NewManagerParamStore() (*common.ParamStore, error) {
	store := common.NewParamStore("policy_manager")
	err := PopulateManagerParams(store)
	return store, err
}

func PopulateManagerParams(store *common.ParamStore) error {
	return store.AddParamDefinitions(map[string]*common.ParamDescription{
		"PolicyStoreName": {
			Kind:     uint32(reflect.String),
			Path:     "policy.store_name",
			Required: common.ParamNecessity_REQUIRED,
		},
		"PolicyStoreParams": {
			Kind:     uint32(reflect.Map),
			Path:     "policy.store_params",
			Required: common.ParamNecessity_OPTIONAL,
		},
	})
}

type Manager struct {
	StoreName     string
	Store         common.IPolicyStore
	PluginManager *common.PluginManager
}

func NewManager(pluginManager *common.PluginManager) *Manager {
	return &Manager{
		StoreName:     "[none]",
		PluginManager: pluginManager,
	}
}

func (pm *Manager) Init(params *common.ParamStore) error {
	storeName := common.Canonize(params.GetString("PolicyStoreName"))
	storeParamMap := params.GetStringMapString("PolicyStoreParams")

	loaded, err := pm.PluginManager.Load("policystore", storeName)
	if err != nil {
		return err
	}

	var ok bool
	pm.Store, ok = loaded.(common.IPolicyStore)
	if !ok {
		return fmt.Errorf(
			"policystore plugin '%s' does not implement IPolicyStore",
			storeName,
		)
	}

	storeParams := common.NewParamStore("policystore")
	pdesc, err := pm.Store.GetParamDescriptions()
	if err != nil {
		return err
	}

	if err = storeParams.AddParamDefinitions(pdesc); err != nil {
		return err
	}

	if err = storeParams.PopulateFromStringMapString(storeParamMap); err != nil {
		return err
	}

	if err = pm.Store.Init(storeParams); err != nil {
		return err
	}

	return nil
}

func (pm *Manager) ListPolicies(tenantID int) ([]common.PolicyListEntry, error) {
	return pm.Store.ListPolicies(tenantID)
}

func (pm *Manager) GetPolicy(tenantID int, tokenFormat common.AttestationFormat) (*common.Policy, error) {
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

func (pm *Manager) DeletePolicy(tenantID int, tokenFormat common.AttestationFormat) error {
	return pm.Store.DeletePolicy(tenantID, tokenFormat)
}

func (pm *Manager) Close() error {
	return nil
}
