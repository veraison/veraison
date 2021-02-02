// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsement

import (
	"github.com/hashicorp/go-plugin"

	"veraison/common"
)

type EndorsementManager struct {
	StoreName string
	Store     common.IEndorsementStore
	RpcClient plugin.ClientProtocol
	Client    *plugin.Client
}

func NewEndorsementManager() *EndorsementManager {
	return &EndorsementManager{
		StoreName: "[none]",
	}
}

func (em *EndorsementManager) InitializeStore(
	pluginLocaitons []string,
	name string,
	params common.EndorsementStoreParams,
) error {
	lp, err := common.LoadPlugin(pluginLocaitons, "endorsementstore", name)
	if err != nil {
		return err
	}

	em.Store = lp.Raw.(common.IEndorsementStore)
	em.Client = lp.PluginClient
	em.RpcClient = lp.RpcClient

	if err = em.Store.Init(params); err != nil {
		em.Client.Kill()
		return err
	}

	return nil
}

func (em *EndorsementManager) GetSupportedQueries() []string {
	return em.Store.GetSupportedQueries()
}

func (em *EndorsementManager) RunQuery(name string, args common.QueryArgs) (common.QueryResult, error) {
	return em.Store.RunQuery(name, args)
}

func (em *EndorsementManager) GetEndorsements(qds ...common.QueryDescriptor) (common.EndorsementMatches, error) {
	return em.Store.GetEndorsements(qds...)
}

func (em *EndorsementManager) Close() {
	em.Client.Kill()
	em.RpcClient.Close()
	em.Store.Close()
}
