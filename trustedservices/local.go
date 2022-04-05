// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package trustedservices

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/veraison/common"
	"github.com/veraison/veraison/kvstore"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	DummyTenantID = "0"
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
		"EndorsementKVStoreConfig": {
			Kind:     uint32(reflect.Map),
			Path:     "endorsement.kvstore_config",
			Required: common.ParamNecessity_REQUIRED,
		},
		"TrustAnchorKVStoreConfig": {
			Kind:     uint32(reflect.Map),
			Path:     "trust_anchor.kvstore_config",
			Required: common.ParamNecessity_REQUIRED,
		},
	})
}

func initStore(params *common.ParamStore, basename string) (kvstore.IKVStore, error) {
	var (
		store       kvstore.IKVStore
		storeConfig kvstore.Config
		err         error
	)

	storeConfig, err = params.TryGetStringMap(basename + "KVStoreConfig")
	if err != nil {
		return nil, err
	}

	store, err = kvstore.New(storeConfig)
	if err != nil {
		return nil, err
	}

	return store, nil
}

type LocalClient struct {
	Schemes          map[common.AttestationFormat]common.IScheme
	TrustAnchorStore kvstore.IKVStore
	EndorsementStore kvstore.IKVStore
	PluginManager    *common.PluginManager
}

func NewLocalClient(pluginManager *common.PluginManager) *LocalClient {
	client := new(LocalClient)
	client.PluginManager = pluginManager
	return client
}

func (c *LocalClient) Init(params *common.ParamStore) error {
	var err error

	c.Schemes = make(map[common.AttestationFormat]common.IScheme)

	c.TrustAnchorStore, err = initStore(params, "TrustAnchor")
	if err != nil {
		return err
	}

	c.EndorsementStore, err = initStore(params, "Endorsement")
	if err != nil {
		return err
	}

	return nil
}

func (o *LocalClient) AddSwComponents(req *common.AddSwComponentsRequest) (*common.AddSwComponentsResponse, error) {
	var (
		keys []string
		val  []byte
	)

	for _, swComp := range req.GetSwComponents() {
		attestFormat := swComp.GetScheme()

		loaded, err := o.PluginManager.Load("scheme", attestFormat.String())
		if err != nil {
			return addSwComponentErrorResponse(err), nil
		}

		scheme, ok := loaded.(common.IScheme)
		if !ok {
			err = fmt.Errorf(
				"plugin '%s' does not implement IScheme",
				attestFormat.String(),
			)
			return addSwComponentErrorResponse(err), nil
		}

		keys, err = scheme.SynthKeysFromSwComponent(DummyTenantID, swComp)
		if err != nil {
			return addSwComponentErrorResponse(err), nil
		}

		val, err = json.Marshal(swComp)
		if err != nil {
			return addSwComponentErrorResponse(err), nil
		}
	}

	for _, key := range keys {
		if err := o.EndorsementStore.Add(key, string(val)); err != nil {
			if err != nil {
				return addSwComponentErrorResponse(err), nil
			}
		}
	}

	return addSwComponentSuccessResponse(), nil
}

func addSwComponentSuccessResponse() *common.AddSwComponentsResponse {
	return &common.AddSwComponentsResponse{
		Status: &common.Status{
			Result: true,
		},
	}
}

func addSwComponentErrorResponse(err error) *common.AddSwComponentsResponse {
	return &common.AddSwComponentsResponse{
		Status: &common.Status{
			Result:      false,
			ErrorDetail: fmt.Sprintf("%v", err),
		},
	}
}

func (o *LocalClient) AddTrustAnchor(req *common.AddTrustAnchorRequest) (*common.AddTrustAnchorResponse, error) {
	var (
		err    error
		keys   []string
		scheme common.IScheme
		ta     *common.Endorsement
		val    []byte
	)

	ta = req.TrustAnchor
	if ta == nil {
		err = errors.New("nil TrustAnchor in request")
		return addTrustAnchorErrorResponse(err), nil
	}

	scheme, err = o.getScheme(ta.GetScheme())
	if err != nil {
		return addTrustAnchorErrorResponse(err), nil
	}

	keys, err = scheme.SynthKeysFromTrustAnchor(DummyTenantID, ta)
	if err != nil {
		return addTrustAnchorErrorResponse(err), nil
	}

	val, err = json.Marshal(ta)
	if err != nil {
		return addTrustAnchorErrorResponse(err), nil
	}

	for _, key := range keys {
		if err := o.TrustAnchorStore.Add(key, string(val)); err != nil {
			if err != nil {
				return addTrustAnchorErrorResponse(err), nil
			}
		}
	}

	return addTrustAnchorSuccessResponse(), nil
}

func addTrustAnchorSuccessResponse() *common.AddTrustAnchorResponse {
	return &common.AddTrustAnchorResponse{
		Status: &common.Status{
			Result: true,
		},
	}
}

func addTrustAnchorErrorResponse(err error) *common.AddTrustAnchorResponse {
	return &common.AddTrustAnchorResponse{
		Status: &common.Status{
			Result:      false,
			ErrorDetail: fmt.Sprintf("%v", err),
		},
	}
}

func (c *LocalClient) GetAttestation(token *common.AttestationToken) (*common.Attestation, error) {
	scheme, err := c.getScheme(token.Format)
	if err != nil {
		return nil, err
	}

	ec, err := c.extractEvidence(scheme, token)
	if err != nil {
		return nil, err
	}

	endorsements, err := c.EndorsementStore.Get(ec.SoftwareId)
	if err != nil {
		return nil, err
	}

	return scheme.GetAttestation(ec, endorsements)
}

func (c *LocalClient) getScheme(format common.AttestationFormat) (common.IScheme, error) {
	scheme, ok := c.Schemes[format]
	if ok {
		return scheme, nil
	}

	loaded, err := c.PluginManager.Load("scheme", format.String())
	if err != nil {
		return nil, err
	}

	scheme, ok = loaded.(common.IScheme)
	if !ok {
		err = fmt.Errorf(
			"plugin '%s' does not implement IScheme",
			format.String(),
		)
		return nil, err
	}

	c.Schemes[format] = scheme
	return scheme, nil
}

func (c LocalClient) Close() error {
	var (
		msg []string
		err error
	)

	err = c.TrustAnchorStore.Close()
	if err != nil {
		msg = append(msg, fmt.Sprintf("problem closing trust anchor store: %s", err.Error()))
	}

	err = c.EndorsementStore.Close()
	if err != nil {
		msg = append(msg, fmt.Sprintf("problem closing endorsement store: %s", err.Error()))
	}

	if len(msg) > 0 {
		return errors.New(strings.Join(msg, "; "))
	}

	return nil
}

func (c *LocalClient) extractEvidence(
	scheme common.IScheme,
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

	if len(trustAnchor) != 1 {
		return nil, fmt.Errorf("found %d trust anchors, want 1", len(trustAnchor))
	}

	extracted, err := scheme.ExtractEvidence(token, trustAnchor[0])
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
