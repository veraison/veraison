// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"github.com/veraison/common"

	"go.uber.org/zap"
)

func getInput(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)
	return err
}

func initDb(schemaFile string) (string, error) {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	dbf, err := ioutil.TempFile(os.TempDir(), "veraison-db-")
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

func TestVerifier(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	schemaFile := filepath.Join(wd, "test", "iat-policy.sqlite")
	policyDbPath, err := initDb(schemaFile)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer finiDb(policyDbPath)

	schemaFile = filepath.Join(wd, "test", "iat-endorsement.sqlite")
	endorsementDbPath, err := initDb(schemaFile)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer finiDb(endorsementDbPath)

	pluginDir := filepath.Join(wd, "..", "plugins", "bin")

	var vc = Config{
		PluginLocations:      []string{pluginDir},
		PolicyEngineName:     "opa",
		PolicyStoreName:      "sqlite",
		EndorsementStoreName: "sqlite",
		PolicyStoreParams: common.PolicyStoreParams{
			"dbpath": policyDbPath,
		},
		EndorsementStoreParams: common.EndorsementStoreParams{
			"dbpath": endorsementDbPath,
		},
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("%v", err)
	}

	v, err := NewVerifier(logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = v.Initialize(vc)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Close()

	evidenceDir := filepath.Join(wd, "test", "evidence")
	fis, err := ioutil.ReadDir(evidenceDir)
	if err != nil {
		t.Errorf("%v", err)
	}

	ec := common.EvidenceContext{
		TenantID: 1,
		Format:   common.PsaIatToken,
	}

	for _, fi := range fis {
		if !strings.HasSuffix(fi.Name(), ".json") {
			continue
		}

		evidencePath := filepath.Join(evidenceDir, fi.Name())
		var evidence map[string]interface{}
		err = getInput(evidencePath, &evidence)
		if err != nil {
			t.Fatalf("%v", err)
		}

		ec.Evidence = evidence
		result, err := v.Verify(&ec, true)
		if err != nil {
			t.Fatalf("%v", err)
		}

		assert.NotNil(result)

		if strings.HasPrefix(fi.Name(), "valid-") && result.Status != common.Status_SUCCESS {
			t.Fatalf("%v resported as invalid", fi.Name())
		}
		if strings.HasPrefix(fi.Name(), "invalid-") && result.Status == common.Status_SUCCESS {
			t.Fatalf("%v resported as valid", fi.Name())
		}
	}
}
