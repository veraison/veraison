// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package tokenprocessor

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/veraison/common"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

var FWID = []uint8{
	0x6e, 0x34, 0x0b, 0x9c, 0xff, 0xb3, 0x7a, 0x98, 0x9c, 0xa5, 0x44,
	0xe6, 0xbb, 0x78, 0x0a, 0x2c, 0x78, 0x90, 0x1d, 0x3f, 0xb3, 0x37,
	0x38, 0x76, 0x85, 0x11, 0xa3, 0x06, 0x17, 0xaf, 0xa0, 0x1d,
}

var DeviceID = []uint8{
	0x04, 0x9f, 0x34, 0x66, 0x25, 0x8b, 0x71, 0x06, 0x23, 0x7f, 0xeb,
	0x64, 0x8d, 0xdf, 0xd5, 0x56, 0xb8, 0xb9, 0x38, 0xc3, 0x07, 0x6a,
	0x61, 0x89, 0x40, 0x97, 0x5f, 0x20, 0x98, 0x9d, 0x61, 0xf3, 0x79,
	0x9d, 0x04, 0x82, 0xf4, 0x6c, 0x8c, 0x4a, 0x94, 0xba, 0x5e, 0x00,
	0x3d, 0xda, 0x66, 0x8f, 0x58, 0xc2, 0x90, 0xca, 0x6f, 0x63, 0xda,
	0xe5, 0xcd, 0x5a, 0x5a, 0x54, 0xcc, 0x0c, 0x07, 0x2f, 0xae,
}

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

func Test_TokenProcessor_ProcessDice(t *testing.T) {
	require := require.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	schemaFile := filepath.Join(wd, "test", "trustanchor.sql")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer finiDb(dbPath)

	pluginDir := filepath.Join(wd, "..", "plugins", "bin")

	config := Config{
		PluginLocations:      []string{pluginDir},
		TrustAnchorStoreName: "sqlite",
		TrustAnchorStoreParams: common.TrustAnchorStoreParams{
			"dbpath": dbPath,
		},
	}

	var tp TokenProcessor
	err = tp.Init(config)
	if err != nil {
		t.Fatalf("%v", err)
	}

	tokenFile := filepath.Join(wd, "test", "DeviceCerts.pem")
	tokenData, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expectedFWID := base64.StdEncoding.EncodeToString(FWID)
	expectedDeviceID := base64.StdEncoding.EncodeToString(DeviceID)

	ec, err := tp.Process(1, common.DiceToken, tokenData)
	require.Nil(err)
	require.Equal(1, ec.TenantID)
	require.Equal(common.DiceToken, ec.Format)
	require.Equal(expectedFWID, ec.Evidence["FWID"])
	require.Equal(expectedDeviceID, ec.Evidence["DeviceID"])
}
