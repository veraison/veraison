// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package tokenprocessor

import (
	"github.com/hashicorp/go-plugin"

	"github.com/veraison/common"
)

type Config struct {
	PluginLocations        []string
	TrustAnchorStoreName   string
	TrustAnchorStoreParams common.TrustAnchorStoreParams
}

type TokenProcessor struct {
	TrustAnchorStoreName string
	TrustAnchorStore     common.ITrustAnchorStore
	Client               *plugin.Client
	RPCClient            plugin.ClientProtocol
	PluginLocations      []string

	extractors map[common.TokenFormat]LoadedExtractorPlugin
}

type LoadedExtractorPlugin struct {
	Extractor common.IEvidenceExtractor
	Client    *plugin.Client
	RPCClient plugin.ClientProtocol
}

func (tp *TokenProcessor) Init(config Config) error {
	lp, err := common.LoadPlugin(config.PluginLocations, "trustanchorstore", config.TrustAnchorStoreName, false)
	if err != nil {
		return err
	}
	tp.TrustAnchorStoreName = config.TrustAnchorStoreName

	tp.TrustAnchorStore = lp.Raw.(common.ITrustAnchorStore)
	tp.RPCClient = lp.RPCClient
	tp.Client = lp.PluginClient
	tp.PluginLocations = config.PluginLocations
	tp.extractors = make(map[common.TokenFormat]LoadedExtractorPlugin)

	if err = tp.TrustAnchorStore.Init(config.TrustAnchorStoreParams); err != nil {
		tp.Client.Kill()
		return err
	}

	return nil
}

func (tp *TokenProcessor) Close() {
	tp.Client.Kill()
	tp.RPCClient.Close()
}

func (tp TokenProcessor) GetExtractor(format common.TokenFormat) (common.IEvidenceExtractor, error) {
	extractorPlugin, ok := tp.extractors[format]
	if ok {
		return extractorPlugin.Extractor, nil
	}

	lp, err := common.LoadPlugin(tp.PluginLocations, "evidenceextractor", format.String(), false)
	if err != nil {
		return nil, err
	}

	loadedExtractor := LoadedExtractorPlugin{
		Extractor: lp.Raw.(common.IEvidenceExtractor),
		Client:    lp.PluginClient,
		RPCClient: lp.RPCClient,
	}

	if err = loadedExtractor.Extractor.Init(common.EvidenceExtractorParams{}); err != nil {
		lp.PluginClient.Kill()
		return nil, err
	}

	tp.extractors[format] = loadedExtractor

	return loadedExtractor.Extractor, nil
}

func (tp TokenProcessor) Process(
	tenantID int,
	format common.TokenFormat,
	token []byte,
) (*common.EvidenceContext, error) {
	extractor, err := tp.GetExtractor(format)
	if err != nil {
		return nil, err
	}

	taID, err := extractor.GetTrustAnchorID(token)
	if err != nil {
		return nil, err
	}

	trustAnchor, err := tp.TrustAnchorStore.GetTrustAnchor(tenantID, taID)
	if err != nil {
		return nil, err
	}

	evidence, err := extractor.ExtractEvidence(token, trustAnchor)
	if err != nil {
		return nil, err
	}

	return &common.EvidenceContext{TenantID: tenantID, Format: format, Evidence: evidence}, nil
}
