package main

import (
	"flag"
	"fmt"

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
		fmt.Printf("Error: %v", err)
		return
	}
	defer logger.Sync()

	tokenProcessor, err := frontend.NewTokenProcessor(pluginDir, dbPath)
	if err != nil {
		logger.Error("Could not init token processor", zap.Error(err))
		return
	}

	verifier, err := frontend.NewVerifier(pluginDir, dbPath, logger)
	if err != nil {
		logger.Error("Could not init verifier", zap.Error(err))
		return
	}

	router := frontend.NewRouter(logger, tokenProcessor, verifier)
	router.Run()
}
