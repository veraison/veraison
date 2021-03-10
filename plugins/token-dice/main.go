// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"

	"github.com/hashicorp/go-plugin"
	"github.com/veraison/dice"

	"github.com/veraison/common"
)

var altNameID = asn1.ObjectIdentifier{2, 5, 29, 17}

type EvidenceExtractor struct {
}

func (ee EvidenceExtractor) GetName() string {
	return "dice"
}

// Init is a no-op for this plugin.
func (ee EvidenceExtractor) Init(params common.EvidenceExtractorParams) error {
	return nil
}

// Close is a no-op for this plugin.
func (ee EvidenceExtractor) Close() error {
	return nil
}

func (ee EvidenceExtractor) GetTrustAnchorID(token []byte) (common.TrustAnchorID, error) {
	return common.TrustAnchorID{Type: common.TaTypeCert}, nil
}

func (ee EvidenceExtractor) ExtractEvidence(token []byte, trustAnchor []byte) (map[string]interface{}, error) {
	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	if err := parseTrustAnchor(trustAnchor, roots, intermediates); err != nil {
		return nil, err
	}

	aliasCert, err := parseTokenCerts(token, intermediates, roots)
	if err != nil {
		return nil, err
	}

	opts := x509.VerifyOptions{
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		Roots:         roots,
		Intermediates: intermediates,
	}

	claims, err := extractEvidenceClaims(aliasCert)
	if err != nil {
		return nil, err
	}

	// note: must verify this after extracting claims so that the Subject Alternative Name
	// gets processed; otherwise, it will be raised an unhandled critical extension.
	if _, err = aliasCert.Verify(opts); err != nil {
		return nil, errors.New("failed to verify alias cert: " + err.Error())
	}

	return claims, err
}

func extractEvidenceClaims(cert *x509.Certificate) (map[string]interface{}, error) {
	claims := make(map[string]interface{})

	for _, ext := range cert.Extensions {
		if ext.Id.Equal(altNameID) {
			if err := processAltName(ext.Value, &claims); err != nil {
				return nil, err
			}
			break
		}
	}

	// Remove Subject Alternative Name from Unhandled critical extensions list, as
	// we've now "handled" it. This will allow the cert to be verified.
	altNameIdx := -1
	for i, extOID := range cert.UnhandledCriticalExtensions {
		if extOID.Equal(altNameID) {
			altNameIdx = i
			break
		}
	}

	if altNameIdx != -1 {
		cert.UnhandledCriticalExtensions = append(cert.UnhandledCriticalExtensions[:altNameIdx],
			cert.UnhandledCriticalExtensions[altNameIdx+1:]...)
	}

	return claims, nil
}

func processAltName(data []byte, claims *map[string]interface{}) error {

	var dice dice.DiceExtension

	rest, err := dice.UnmarshalDER(data)
	if err != nil {
		return err
	}
	if len(rest) != 0 {
		return errors.New("trailing data after DICE extension")
	}

	(*claims)["FWID"] = dice.CompositeDeviceID.Fwid.Fwid
	(*claims)["DeviceID"] = dice.CompositeDeviceID.DeviceID.SubjectPublicKey.Bytes

	return nil
}

func parseTokenCerts(token []byte, intermediates *x509.CertPool, roots *x509.CertPool) (*x509.Certificate, error) {
	block, rest := pem.Decode(token)
	if block == nil {
		return nil, errors.New("problem extracting token cert PEM block")
	}

	aliasCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	block, rest = pem.Decode(rest)
	if block == nil {
		return nil, errors.New("problem extrating token cert PEM block")
	}

	deviceCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	// self signed cert should not have any intermediates presented with it.
	if deviceCert.Subject.String() == deviceCert.Issuer.String() {
		if len(rest) != 0 {
			return nil, errors.New("additional data found alongside a self-signed Cert")
		}

		roots.AddCert(deviceCert)

		return aliasCert, nil
	}

	// Device cert is not self-signed. Add it as an intermediate and process
	// the rest of the certs if any.

	intermediates.AddCert(deviceCert)

	for len(rest) != 0 {
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, errors.New("problem extracting token intermediate PEM block")
		}

		intCert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		intermediates.AddCert(intCert)
	}

	return aliasCert, nil
}

func parseTrustAnchor(trustAnchor []byte, roots *x509.CertPool, intermediates *x509.CertPool) error {
	var block *pem.Block
	rest := trustAnchor
	for len(rest) != 0 {
		block, rest = pem.Decode(rest)
		if block == nil {
			return errors.New("problem extracting trust anchor PEM block")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return err
		}

		if cert.Subject.String() == cert.Issuer.String() {
			// self-signed
			roots.AddCert(cert)
		} else {
			intermediates.AddCert(cert)
		}
	}

	return nil
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"evidenceextractor": &common.EvidenceExtractorPlugin{
			Impl: &EvidenceExtractor{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
