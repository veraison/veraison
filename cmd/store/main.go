package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"

	"github.com/veraison/common"
	"github.com/veraison/endorsement"
)

// TODO this a very minimal "frontend" implementation.
func main() {

	configPaths := common.NewConfigPaths()

	flag.Var(configPaths, "config", "Path to direcotory containing the config file(s).")
	flag.Parse()

	config, err := NewConfig(*configPaths)
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}
	if err = config.ReadInConfig(); err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port)) //nolint
	if err != nil {
		log.Fatalf("could not create listener: %v", err)
	}

	store := endorsement.Store{}

	if err := store.Init(config); err != nil {
		log.Fatalf("could not initialize store: %v", err)
	}

	grpcServer := grpc.NewServer()
	endorsement.RegisterStoreServer(grpcServer, &store)
	endorsement.RegisterFetcherServer(grpcServer, &store)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("server error: %v", err)
	}

	store.Fini()
}
