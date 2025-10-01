package main

import (
	"log"
	"net"

	"scf-chaincode/engine"

	"google.golang.org/grpc"
)

func main() {
	// Start gRPC server
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	engine.RegisterSmartContractService(s)

	log.Println("Smart Contract Engine running on :9090")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
