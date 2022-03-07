// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package trustedservices

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/veraison/common"
	"github.com/veraison/veraison/kvstore"
	"google.golang.org/protobuf/types/known/structpb"
)

// Local VTS client, used by executable that the VTS component is part of. All
// other, "remote", clients are basically proxies into this.

func NewLocalClientParamStore() (*common.ParamStore, error) {
	store := common.NewParamStore("vts-local")
	err := PopulateLocalClientParams(store)
	return store, err
}

func PopulateLocalClientParams(store *common.ParamStore) error {
	return store.AddParamDefinitions(map[string]*common.ParamDescription{
		"PluginLocations": {
			Kind:     uint32(reflect.Slice),
			Path:     "plugin.locations",
			Required: common.ParamNecessity_REQUIRED,
		},
		"EndorsementStoreName": {
			Kind:     uint32(reflect.String),
			Path:     "endorsement.store_name",
			Required: common.ParamNecessity_REQUIRED,
		},
		"EndorsementStoreParams": {
			Kind:     uint32(reflect.Map),
			Path:     "endorsement.store_params",
			Required: common.ParamNecessity_OPTIONAL,
		},
		"TrustAnchorStoreName": {
			Kind:     uint32(reflect.String),
			Path:     "trust_anchor.store_name",
			Required: common.ParamNecessity_REQUIRED,
		},
		"TrustAnchorStoreParams": {
			Kind:     uint32(reflect.Map),
			Path:     "trust_anchor.store_params",
			Required: common.ParamNecessity_OPTIONAL,
		},
	})
}

type LocalClientConnector struct {
}

func (c LocalClientConnector) Connect(
	host string,
	port int,
	params map[string]string,
) (common.ITrustedServicesClient, error) {
	paramStore, err := NewLocalClientParamStore()
	if err != nil {
		return nil, err
	}

	err = paramStore.PopulateFromStringMapString(params)
	if err != nil {
		return nil, err
	}

	var client LocalClient

	err = client.Init(paramStore)

	return &client, err
}

type LocalClient struct {
	PluginLocations  []string
	Schemes          map[common.AttestationFormat]*common.SchemePlugin
	TrustAnchorStore kvstore.KVStore
	EndorsementStore kvstore.KVStore
}

func (c *LocalClient) Init(params *common.ParamStore) error {
	var err error

	c.PluginLocations, err = params.TryGetStringSlice("PluginLocations")
	if err != nil {
		return err
	}

	c.Schemes = make(map[common.AttestationFormat]*common.SchemePlugin)

	storeName, err := params.TryGetString("TrustAnchorStoreName")
	if err != nil {
		return err
	}

	switch storeName {
	case "memory":
		c.TrustAnchorStore = new(kvstore.Memory)
	case "sql":
		c.TrustAnchorStore = new(kvstore.SQL)
	default:
		return fmt.Errorf("unknown TrustAnchorStoreName: %q", storeName)
	}

	storeParams, err := params.TryGetStringMap("TrustAnchorStoreParams")
	if err != nil {
		return err
	}

	err = c.TrustAnchorStore.Init(kvstore.Config(storeParams))
	if err != nil {
		return err
	}

	storeName, err = params.TryGetString("EndorsementStoreName")
	if err != nil {
		return err
	}

	switch storeName {
	case "memory":
		c.EndorsementStore = new(kvstore.Memory)
	case "sql":
		c.EndorsementStore = new(kvstore.SQL)
	default:
		return fmt.Errorf("unknown EndorsementStoreName: %q", storeName)
	}

	storeParams, err = params.TryGetStringMap("EndorsementStoreParams")
	if err != nil {
		return err
	}

	err = c.EndorsementStore.Init(kvstore.Config(storeParams))
	if err != nil {
		return err
	}

	return nil
}

func (c *LocalClient) GetAttestation(token *common.AttestationToken) (*common.Attestation, error) {
	scheme, err := c.getSchemePlugin(token.Format)
	if err != nil {
		return nil, err
	}

	ec, err := c.extractEvidence(scheme, token)

	endorsementString, err := c.EndorsementStore.Get(ec.SoftwareId)
	if err != nil {
		return nil, err
	}

	return scheme.GetAttestation(ec, endorsementString)
}

func (c LocalClient) Close() error {
	var msg []string
	err := c.TrustAnchorStore.Close()
	if err != nil {
		msg = append(msg, fmt.Sprintf("probelm closing trust anchor store: %s", err.Error()))
	}

	err = c.EndorsementStore.Close()
	if err != nil {
		msg = append(msg, fmt.Sprintf("problem closing endorsement store: %s", err.Error()))
	}

	for format, sp := range c.Schemes {
		err := sp.Close()
		if err != nil {
			msg = append(msg, fmt.Sprintf("error closing %q: %s", format.String(), err.Error()))
		}

	}

	if len(msg) > 0 {
		return errors.New(strings.Join(msg, "; "))
	} else {
		return nil
	}

}

func (c *LocalClient) getSchemePlugin(format common.AttestationFormat) (*common.SchemePlugin, error) {
	sp, ok := c.Schemes[format]
	if ok {
		return sp, nil
	}

	sp, err := common.LoadSchemePlugin(c.PluginLocations, format)
	if err != nil {
		return nil, err
	}

	c.Schemes[format] = sp
	return sp, nil
}

func (c *LocalClient) extractEvidence(
	scheme *common.SchemePlugin,
	token *common.AttestationToken,
) (*common.EvidenceContext, error) {
	var err error
	ec := new(common.EvidenceContext)

	ec.TenantId = token.TenantId
	ec.Format = token.Format
	ec.TrustAnchorId, err = scheme.GetTrustAnchorID(token)
	if err != nil {
		return nil, err
	}

	trustAnchor, err := c.TrustAnchorStore.Get(ec.TrustAnchorId)
	if err != nil {
		return nil, err
	}

	extracted, err := scheme.ExtractEvidence(token, trustAnchor)
	if err != nil {
		return nil, err
	}

	ec.Evidence, err = structpb.NewStruct(extracted.Evidence)
	if err != nil {
		return nil, err
	}

	ec.SoftwareId = extracted.SoftwareID

	return ec, nil

}
