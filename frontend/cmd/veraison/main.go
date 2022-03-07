// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"flag"

	"github.com/veraison/frontend"
	"go.uber.org/zap"
)

func main() {
	var pluginDir, dbPath string

	flag.StringVar(&pluginDir, "p", "", "Path to directory containing plugin binaries.")
	flag.StringVar(&dbPath, "d", "", "Path to the directory containing sqlite database files.")
	flag.Parse()

	logger, err := zap.NewDevelopment() // TODO configurable
	if err != nil {
		logger.Fatal("Error: %v", zap.Error(err))
	}
	defer logger.Sync() //nolint:errcheck

	verifier, err := frontend.NewVerifier(pluginDir, dbPath, logger)
	if err != nil {
		logger.Fatal("Could not init verifier", zap.Error(err))
	}

	router := frontend.NewRouter(logger, verifier)
	err = router.Run()
	if err != nil {
		logger.Error("error running server", zap.Error(err))
	}
}
