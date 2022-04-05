// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import "fmt"

// ITrustedServicesConnector provides a means of establishing connection to the trusted services component.
type ITrustedServicesConnector interface {
	Connect(host string, port int, params map[string]string) (ITrustedServicesClient, error)
}

// ITrustedServicesClient specifies the client interface for the trusted services component.
type ITrustedServicesClient interface {
	Init(params *ParamStore) error

	// GetAttestation returns attestation information -- evidences,
	// endorsed claims, trust vector, etc -- for the provided attestation
	// token data.
	GetAttestation(token *AttestationToken) (*Attestation, error)

	Close() error
}



type ClientDescriptor  struct {
	Name  string
	Client ITrustedServicesClient
	Store *ParamStore
}

func (d *ClientDescriptor) Connect(params *ParamStore) (ITrustedServicesClient, error) {
	d.Store.PopulateFromStore(params)
	if err := d.Store.Validate(true); err != nil {
		return nil, err
	}

	if err := d.Client.Init(d.Store); err != nil {
		return nil, err
	}

	return  d.Client, nil
}

type TrustedServicesConnector struct {
	Clients map[string]*ClientDescriptor
}


func (c *TrustedServicesConnector) Register(name string,  client ITrustedServicesClient, store *ParamStore) error {
	if _, ok := c.Clients[name]; ok {
		return fmt.Errorf("client with name %q already registered", name)
	}

	c.Clients[name] = &ClientDescriptor{Name: name, Client: client, Store: store}

	return nil
}

func (c *TrustedServicesConnector) Connect(name string, params *ParamStore) (ITrustedServicesClient, error) {
	cliDesc, ok := c.Clients[name]
	if !ok {
		return nil, fmt.Errorf("client %q does not exist", name)
	}

	return cliDesc.Connect(params)
}
