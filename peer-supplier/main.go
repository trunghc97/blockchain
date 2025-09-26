package main

import (
	"log"
	"os"

	"shared/peer-base"
)

func main() {
	// Set default environment variables for supplier peer
	if os.Getenv("PEER_NODE_TYPE") == "" {
		os.Setenv("PEER_NODE_TYPE", "supplier")
	}
	if os.Getenv("PEER_NODE_ID") == "" {
		os.Setenv("PEER_NODE_ID", "supplier-peer-1")
	}
	if os.Getenv("PEER_PORT") == "" {
		os.Setenv("PEER_PORT", "8083")
	}
	if os.Getenv("PEER_GRPC_PORT") == "" {
		os.Setenv("PEER_GRPC_PORT", "9093")
	}
	if os.Getenv("ORDERER_ADDR") == "" {
		os.Setenv("ORDERER_ADDR", "orderer-cluster-ord1:7050")
	}

	log.Println("Starting Supplier Peer Node...")
	peer_base.Main()
}