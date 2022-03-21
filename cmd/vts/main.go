// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/veraison/common"
	"github.com/veraison/trustedservices"
	"google.golang.org/grpc"
)

// TODO this a very minimal "frontend" implementation.
func main() {
	configPaths := common.NewConfigPaths()

	flag.Var(configPaths, "config", "Path to direcotory containing the config file(s).")
	flag.Parse()

	clientParams, err := trustedservices.NewLocalClientParamStore()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}
	serverParams, err := trustedservices.NewRPCServerParamStore()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	config, err := common.NewConfig(*configPaths, clientParams, serverParams)
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}
	if err = config.ReadInConfig(); err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serverParams.GetInt("Port"))) //nolint
	if err != nil {
		log.Fatalf("could not create listener: %v", err)
	}

	server := trustedservices.RPCServer{}

	grpcServer := grpc.NewServer()
	common.RegisterVTSServer(grpcServer, &server)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
