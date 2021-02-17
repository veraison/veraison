// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"veraison/common"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func initDb(schemaFile string) (string, error) {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	dbf, err := ioutil.TempFile(os.TempDir(), "trustanchor-db-")
	if err != nil {
		return "", err
	}
	dbPath := dbf.Name()
	dbf.Close()

	dbConfig := fmt.Sprintf("file:%s?cache=shared", dbPath)
	db, err := sql.Open("sqlite3", dbConfig)
	if err != nil {
		return dbPath, err
	}
	defer db.Close()

	commands := strings.Split(string(schema), ";")
	for _, command := range commands {
		_, err := db.Exec(command)
		if err != nil {
			return dbPath, err
		}
	}

	return dbPath, nil
}

func finiDb(path string) {
	if path != "" {
		os.RemoveAll(path)
	}
}

func Test_Cert_RoundTrip(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	schemaFile := filepath.Join(wd, "test", "schema.sqlite")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer finiDb(dbPath)

	var taStore SqliteTrustAnchorStore
	err = taStore.Init(common.TrustAnchorStoreParams{"dbPath": dbPath})
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.Equal("sqlite", taStore.GetName())

	certsFile := filepath.Join(wd, "test", "certs.pem")
	certData, err := ioutil.ReadFile(certsFile)
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = taStore.AddCertsFromPEM(1, certData)
	assert.Nil(err)

	taData, err := taStore.GetTrustAnchor(1, common.TrustAnchorID{Type: common.TaTypeCert})
	assert.Nil(err)

	expected := []string{"Vendor Intermediate CA 1", "Vendor Intermediate CA 2", "Vendor Root CA O=MSR_TEST"}

	var returned []string
	var block *pem.Block
	rest := taData
	for len(rest) != 0 {
		block, rest = pem.Decode(rest)
		if block == nil {
			t.Fatalf("Could not decode PEM block from returned Trust Anchor data.")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("%v", err)
		}
		returned = append(returned, cert.Subject.CommonName)
	}

	assert.Equal(expected, returned)
}

func Test_Key_RoundTrip(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	schemaFile := filepath.Join(wd, "test", "schema.sqlite")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer finiDb(dbPath)

	var taStore SqliteTrustAnchorStore
	err = taStore.Init(common.TrustAnchorStoreParams{"dbPath": dbPath})
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.Equal("sqlite", taStore.GetName())

	keyFile := filepath.Join(wd, "test", "key.pem")
	keyData, err := ioutil.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("%v", err)
	}

	kid := []byte{0x01, 0x02, 0x03, 0x04}

	err = taStore.AddPublicKeyFromPEM(1, kid, keyData)
	assert.Nil(err)

	taID := map[string]interface{}{"key-id": kid}
	taData, err := taStore.GetTrustAnchor(1, common.TrustAnchorID{Type: common.TaTypeKey, Value: taID})
	assert.Nil(err)
	assert.Equal(taData, keyData)
}
