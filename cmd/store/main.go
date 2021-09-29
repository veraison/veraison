package main

import (
	"log"
	"net"

	grpc "google.golang.org/grpc"

	"github.com/veraison/endorsement"
)

// TODO this a very minimal "frontend" implementation.
func main() {
	listener, err := net.Listen("tcp", ":50051") //nolint
	if err != nil {
		log.Fatalf("could not create listener: %v", err)
	}

	store := endorsement.Store{}

	grpcServer := grpc.NewServer()
	endorsement.RegisterStoreServer(grpcServer, &store)
	endorsement.RegisterFetcherServer(grpcServer, &store)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
