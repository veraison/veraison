// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsement

import (
	"github.com/hashicorp/go-plugin"

	"veraison/common"
)

// EndorsementManager handles interadctions with the endorsement store. It is
// responsible for loading the appropriate plugin and maintaing the client
// connection to it.
type EndorsementManager struct {

	// StoreName is the name of the endorsement store plugin that has been
	// loaded, or "[none]" if there isn't one.
	StoreName string

	// Store is the interface to the underyling store.
	Store common.IEndorsementStore

	// RpcClient is an instance of ClientProtocol responsible for handling
	// the underlying RPC channel to the plugin.
	RpcClient plugin.ClientProtocol

	// Client is the client interface to the plugin used for aquring the
	// handle for the interface implemented by the plugin and establishing
	// the RPC channel.
	Client *plugin.Client
}

// NewEndorsementManager creates a new uninitialized EndorsementManager.
func NewEndorsementManager() *EndorsementManager {
	return &EndorsementManager{
		StoreName: "[none]",
	}
}

// InitializeStore establishes a connection to the underlying store via a
// plugin specified by name using the provided params.
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
