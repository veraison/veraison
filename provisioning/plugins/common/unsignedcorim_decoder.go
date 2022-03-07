// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/endorsement"
	"github.com/veraison/veraison/provisioning/decoder"
)

// IExtractor is the interface that CoRIM plugins need to implement to hook into
// the UnsignedCorimDecoder logics.
// Each extractor consumes a specific CoMID triple and produces a corresponding
// Veraison Endorsement format (or an error).
//
// Note: At the moment the interface is limited by the known use cases.  We
// anticipate that in the future there will to be an extractor for each of the
// defined CoMID triples, plus maybe a way to handle cross-triples checks as
// well as extraction from the "global" CoRIM context.
// See also https://github.com/veraison/veraison/issues/112
type IExtractor interface {
	SwCompExtractor(comid.ReferenceValue) ([]*endorsement.SwComponent, error)
	TaExtractor(comid.AttestVerifKey) (*endorsement.TrustAnchor, error)
}

func UnsignedCorimDecoder(data []byte, xtr IExtractor) (*decoder.EndorsementDecoderResponse, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	var uc corim.UnsignedCorim

	if err := uc.FromCBOR(data); err != nil {
		return nil, fmt.Errorf("CBOR decoding failed: %w", err)
	}

	if err := uc.Valid(); err != nil {
		return nil, fmt.Errorf("invalid unsigned corim: %w", err)
	}

	// TODO(tho) check profile

	rsp := decoder.EndorsementDecoderResponse{}

	for i, tag := range uc.Tags {
		// need at least 3 bytes for the tag and 1 for the smallest bstr
		if len(tag) < 3+1 {
			return nil, fmt.Errorf("malformed tag at index %d", i)
		}

		// split tag from data
		cborTag, cborData := tag[:3], tag[3:]

		// The EnactTrust profile only knows about CoMIDs
		if !bytes.Equal(cborTag, corim.ComidTag) {
			return nil, fmt.Errorf("unknown CBOR tag %x detected at index %d", cborTag, i)
		}

		var c comid.Comid

		err := c.FromCBOR(cborData)
		if err != nil {
			return nil, fmt.Errorf("decoding failed for CoMID at index %d: %w", i, err)
		}

		if err := c.Valid(); err != nil {
			return nil, fmt.Errorf("decoding failed for CoMID at index %d: %w", i, err)
		}

		if c.Triples.ReferenceValues != nil {
			for _, rv := range *c.Triples.ReferenceValues {
				swComp, err := xtr.SwCompExtractor(rv)
				if err != nil {
					return nil, fmt.Errorf("bad software component in CoMID at index %d: %w", i, err)
				}

				for i := range swComp {
					rsp.SwComponents = append(rsp.SwComponents, swComp[i])
				}
			}
		}

		if c.Triples.AttestVerifKeys != nil {
			for _, avk := range *c.Triples.AttestVerifKeys {
				k, err := xtr.TaExtractor(avk)
				if err != nil {
					return nil, fmt.Errorf("bad key in CoMID at index %d: %w", i, err)
				}

				rsp.TrustAnchors = append(rsp.TrustAnchors, k)
			}
		}

		// silently ignore any other triples
	}

	return &rsp, nil
}
