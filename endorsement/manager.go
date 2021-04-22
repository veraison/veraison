// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsement

import (
	"github.com/hashicorp/go-plugin"

	"github.com/veraison/common"
)

// Manager handles interadctions with the endorsement store. It is responsible
// for loading the appropriate plugin and maintaining the client connection to
// it.
type Manager struct {

	// StoreName is the name of the endorsement store plugin that has been
	// loaded, or "[none]" if there isn't one.
	StoreName string

	// Store is the interface to the underlying store.
	Store common.IEndorsementStore

	// RpcClient is an instance of ClientProtocol responsible for handling
	// the underlying RPC channel to the plugin.
	RPCClient plugin.ClientProtocol

	// Client is the client interface to the plugin used for aquring the
	// handle for the interface implemented by the plugin and establishing
	// the RPC channel.
	Client *plugin.Client
}

// NewManager creates a new uninitialized Manager.
func NewManager() *Manager {
	return &Manager{
		StoreName: "[none]",
	}
}

// InitializeStore establishes a connection to the underlying store via a
// plugin specified by name using the provided params.
func (em *Manager) InitializeStore(
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
	em.RPCClient = lp.RPCClient

	if err = em.Store.Init(params); err != nil {
		em.Client.Kill()
		return err
	}

	return nil
}

func (em *Manager) GetName() string {
	return em.Store.GetName()
}

func (em *Manager) GetSupportedQueries() []string {
	return em.Store.GetSupportedQueries()
}

func (em *Manager) RunQuery(name string, args common.QueryArgs) (common.QueryResult, error) {
	return em.Store.RunQuery(name, args)
}

func (em *Manager) GetEndorsements(qds ...common.QueryDescriptor) (common.EndorsementMatches, error) {
	return em.Store.GetEndorsements(qds...)
}

func (em *Manager) AddEndorsement(name string, args common.QueryArgs, update bool) error {
	return em.Store.AddEndorsement(name, args, update)
}

func (em *Manager) Close() {
	em.Client.Kill()
	em.RPCClient.Close()
	em.Store.Close()
}
