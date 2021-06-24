package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"github.com/veraison/common"
)

func initDb(schemaFile string) (string, error) {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	dbf, err := ioutil.TempFile(os.TempDir(), "endorsement-db-")
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

func doRunCommand(command string, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	schemaFile := filepath.Join(wd, "test", "endorsements.sqlite")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		return err
	}
	defer finiDb(dbPath)

	config := common.NewConfig()

	config.AddPath(filepath.Join(wd, "test"))
	err = config.Reload()
	if err != nil {
		return err
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	err = runCommand(config, command, args, logger)
	if err != nil {
		return err
	}

	return nil
}

func Example_runQueryCommand() {
	queryParams := []string{
		"-a", "platform_id=76543210fedcba9817161514131211101f1e1d1c1b1a1918", "hardware_id",
	}

	err := doRunCommand("query", queryParams)
	if err != nil {
		fmt.Printf("%v", err)
	}

	// Output: ["acme-rr-trap"]
}

func Example_runListQueriesCommand() {
	queryParams := []string{}

	err := doRunCommand("list-queries", queryParams)
	if err != nil {
		fmt.Printf("%v", err)
	}

	// Output:
	// Endorsement Store "sqlite" supports the following queries:
	//	hardware_id
	//	software_components
}

func Example_runAddCommand() {
	var swComps = `
	[{
		"sw_component_type": "M4",
		"signer_id": "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		"sw_component_version": "1.0.1",
		"measurement_value": "76543210fedcba9817161514131211101f1e1d1c1b1a1920"
	}]
`

	queryParams := []string{
		"-a", "platform_id=76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		"-a", fmt.Sprintf("software_components=%s", swComps),
		"software_components",
	}

	err := doRunCommand("add", queryParams)
	if err != nil {
		fmt.Printf("%v", err)
	}

	// Output: Successfully added software_components endorsement.
}
